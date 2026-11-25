// Package gitcheckout is a pure-Go replacement for actions/checkout
// on Forgejo Runner: it fetches the workflow repository over git
// smart HTTP using go-git, so it needs neither node nor a git binary.
package gitcheckout

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

// Inputs carries the action inputs; the token arrives through the
// environment (TOKEN) so it never shows in argv.
type Inputs struct {
	Repository string
	Ref        string
	Token      string
	Username   string
	Path       string
	FetchDepth string
	Submodules string
}

// Run checks the repository out into the workspace.
func Run(c *gha.Context, log *slog.Logger, in Inputs) error {
	repository := in.Repository
	if repository == "" {
		repository = c.Repository
	}
	ref := in.Ref
	if ref == "" {
		ref = c.Ref
	}
	token := in.Token
	if token == "" {
		token = c.Token
	}
	username := in.Username
	if username == "" {
		username = gha.FirstNonEmpty(c.Actor, "x-access-token")
	}
	if c.ServerURL == "" {
		return fmt.Errorf("GITHUB_SERVER_URL is not set; cannot determine repository URL")
	}
	if repository == "" {
		return fmt.Errorf("no repository given (GITHUB_REPOSITORY unset and no repository input)")
	}
	repoURL := c.RepoURL() + ".git"

	depth, err := gha.IntValue("fetch-depth", in.FetchDepth, 1)
	if err != nil {
		return err
	}

	dest := c.ResolvePath(gha.FirstNonEmpty(in.Path, "."))

	var auth *githttp.BasicAuth
	if token != "" {
		// Forgejo accepts the token as the password with any
		// username, mirroring what actions/checkout does on GitHub.
		auth = &githttp.BasicAuth{Username: username, Password: token}
	}

	candidates := candidateRefs(ref, c.HeadRef, c.SHA)
	if len(candidates) == 0 {
		return fmt.Errorf("nothing to check out: GITHUB_REF is unset and no ref input was given")
	}
	log.Info(fmt.Sprintf("%s ref %q depth %d -> %s", repoURL, ref, depth, dest))

	// Always start from a clean directory; this action deliberately
	// has fresh-clone semantics rather than actions/checkout's re-use.
	// The destination is usually the workspace volume itself, which
	// cannot be removed, so its contents are cleared instead.
	if err := clearDir(dest); err != nil {
		return err
	}

	repo, err := git.PlainInit(dest, false)
	if err != nil {
		return fmt.Errorf("init %s: %w", dest, err)
	}
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})
	if err != nil {
		return fmt.Errorf("configure origin: %w", err)
	}

	var matched string
	var hash plumbing.Hash
	var lastErr error
	for _, cand := range candidates {
		local := "refs/remotes/action/" + sanitizeRef(cand)
		fetch := &git.FetchOptions{
			RefSpecs: []config.RefSpec{config.RefSpec("+" + cand + ":" + local)},
			Force:    true,
			Tags:     git.NoTags,
			Auth:     auth,
		}
		if depth > 0 {
			fetch.Depth = depth
		} else {
			fetch.Tags = git.AllTags
		}
		if err := remote.Fetch(fetch); err == nil {
			if refObj, err := repo.Reference(plumbing.ReferenceName(local), true); err == nil {
				matched, hash = cand, refObj.Hash()
				break
			} else {
				lastErr = err
			}
		} else {
			lastErr = err
		}
		log.Info(fmt.Sprintf("ref %s unavailable (%v), trying next candidate", cand, lastErr))
	}
	if matched == "" {
		if auth == nil {
			lastErr = fmt.Errorf("%w (no token available; private repositories need one)", lastErr)
		}
		return fmt.Errorf("could not fetch any of %v: %w", candidates, lastErr)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if branch, ok := strings.CutPrefix(matched, "refs/heads/"); ok {
		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
			Hash:   hash,
			Create: true,
			Force:  true,
		})
	} else {
		err = wt.Checkout(&git.CheckoutOptions{Hash: hash, Force: true})
	}
	if err != nil {
		return fmt.Errorf("checkout %s: %w", hash, err)
	}

	if in.Submodules == "true" || in.Submodules == "recursive" {
		log.Info(fmt.Sprintf("updating submodules (%s)", in.Submodules))
		subs, err := wt.Submodules()
		if err != nil {
			return err
		}
		if err := subs.Update(&git.SubmoduleUpdateOptions{
			Init:              true,
			RecurseSubmodules: recurseMode(in.Submodules),
			Auth:              auth,
		}); err != nil {
			return fmt.Errorf("submodules: %w", err)
		}
	}

	log.Info(fmt.Sprintf("checked out %s (commit %s)", matched, hash))
	if err := c.Output("ref", matched); err != nil {
		return err
	}
	return c.Output("commit", hash.String())
}

// clearDir removes the contents of dir, creating it when missing.
func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return os.MkdirAll(dir, 0o755)
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// candidateRefs returns the full refs to try fetching, in order: the
// given ref itself, pull-request alternates, bare-name interpretations
// and finally the event SHA. Forgejo uses refs/pull/<n>/head while
// GitHub uses refs/pull/<n>/merge, so both are tried.
func candidateRefs(ref, headRef, sha string) []string {
	var out []string
	add := func(s string) {
		if s == "" {
			return
		}
		if slices.Contains(out, s) {
			return
		}
		out = append(out, s)
	}
	if ref == "" {
		return nil
	}
	if strings.HasPrefix(ref, "refs/") {
		add(ref)
	}
	if strings.HasPrefix(ref, "refs/pull/") || strings.HasPrefix(ref, "refs/merge-request/") {
		if n := pullNumber(ref); n > 0 {
			add(fmt.Sprintf("refs/pull/%d/merge", n))
			add(fmt.Sprintf("refs/pull/%d/head", n))
		}
		if headRef != "" {
			add("refs/heads/" + headRef)
		}
	}
	if !strings.HasPrefix(ref, "refs/") {
		add(ref) // commit SHA, if the server allows fetching it directly
		add("refs/heads/" + ref)
		add("refs/tags/" + ref)
	}
	if isCommitSHA(sha) {
		add(sha)
	}
	return out
}

func pullNumber(ref string) int {
	parts := strings.Split(ref, "/")
	for i, p := range parts {
		if (p == "pull" || p == "merge-request") && i+1 < len(parts) {
			if n, err := strconv.Atoi(parts[i+1]); err == nil {
				return n
			}
		}
	}
	return 0
}

func isCommitSHA(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

// sanitizeRef munges a full ref into a safe local reference name.
func sanitizeRef(ref string) string {
	var b strings.Builder
	for _, r := range ref {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

// recurseMode maps the submodules input to go-git's recursion setting.
func recurseMode(submodules string) git.SubmoduleRescursivity {
	if submodules == "recursive" {
		return git.DefaultSubmoduleRecursionDepth
	}
	return git.NoRecurseSubmodules
}
