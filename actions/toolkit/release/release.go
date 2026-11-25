// Package release creates a release on the Forgejo instance the
// workflow runs on — its own repository, or any repository given via
// input — and uploads files as the release's assets: a compact
// replacement for softprops/action-gh-release (no Node.js). A previous
// release for the same tag is replaced wholesale, so reruns never
// leave stale assets behind. Authentication prefers an explicit access
// token, then an Authorized Integration JWT minted for the audience
// input (the workflow itself is the credential), then the job token.
package release

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/oidc"
)

// Inputs carries the action inputs; the token arrives through the
// environment (TOKEN) so it never shows in argv.
type Inputs struct {
	Tag        string
	Name       string
	Body       string
	Commitish  string
	Assets     string
	Replace    string
	Draft      string
	Prerelease string
	Repository string
	Token      string
	Audience   string
}

// asset is one file to upload: its path on disk and the name it gets
// on the release.
type asset struct {
	path string
	name string
}

// statusError is a failed Forgejo API call, carrying the HTTP status
// so callers can branch on it (409 on create means a concurrent run
// recreated the release).
type statusError struct {
	op     string
	status int
	body   string
}

func (e *statusError) Error() string {
	msg := strings.TrimSpace(e.body)
	if msg == "" {
		return fmt.Sprintf("%s: %s", e.op, http.StatusText(e.status))
	}
	return fmt.Sprintf("%s: %s: %s", e.op, http.StatusText(e.status), msg)
}

func (e *statusError) conflict() bool { return e.status == http.StatusConflict }

// Run creates the release and uploads the assets, reporting the
// release id and URL.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	token, scheme, err := resolveToken(ctx, c, log, in)
	if err != nil {
		return err
	}
	gha.Mask(token)

	repo := strings.TrimSpace(in.Repository)
	if repo == "" {
		repo = c.Repository
	}
	if repo == "" {
		return fmt.Errorf("no repository: set the repository input or run inside a workflow")
	}
	if c.APIURL == "" {
		return fmt.Errorf("no API URL: the runner did not advertise a server URL")
	}
	api := strings.TrimSuffix(c.APIURL, "/") + "/repos/" + repo

	tag := strings.TrimSpace(in.Tag)
	if tag == "" {
		tag = tagFromRef(c.Ref)
	}
	if tag == "" {
		return fmt.Errorf("the tag input is required outside of tag events")
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = tag
	}
	commitish := strings.TrimSpace(in.Commitish)
	if commitish == "" {
		commitish = c.SHA
	}

	assets, err := parseAssets(c.Workspace, in.Assets)
	if err != nil {
		return err
	}
	for _, a := range assets {
		info, err := os.Stat(a.path)
		if err != nil {
			return fmt.Errorf("asset %s: %w", a.name, err)
		}
		if info.IsDir() {
			return fmt.Errorf("asset %s: %s is a directory", a.name, a.path)
		}
	}

	client := &http.Client{Timeout: 30 * time.Minute}
	replace := gha.Truthy(in.Replace, true)
	option := createReleaseOption{
		TagName:      tag,
		Target:       commitish,
		Title:        name,
		Note:         in.Body,
		IsDraft:      gha.Truthy(in.Draft, false),
		IsPrerelease: gha.Truthy(in.Prerelease, false),
	}

	// A previous run of this exact release may still be around;
	// replace it wholesale so stale assets cannot survive. Deleting
	// the release keeps the tag, which the create below reattaches
	// to.
	if replace {
		if old, found, err := findRelease(ctx, client, api, tag, token, scheme); err != nil {
			return err
		} else if found {
			log.Info(fmt.Sprintf("replacing existing release %s (id %d)", tag, old.ID))
			if err := deleteRelease(ctx, client, api, old.ID, token, scheme); err != nil {
				return err
			}
		}
	}

	rel, err := createRelease(ctx, client, api, token, scheme, option)
	var conflict *statusError
	if err != nil && replace && errors.As(err, &conflict) && conflict.conflict() {
		// A release for the tag reappeared between the lookup and
		// the create (a concurrent run of the same build): replace
		// it and try once more.
		if old, found, ferr := findRelease(ctx, client, api, tag, token, scheme); ferr == nil && found {
			if derr := deleteRelease(ctx, client, api, old.ID, token, scheme); derr != nil {
				return derr
			}
			rel, err = createRelease(ctx, client, api, token, scheme, option)
		}
	}
	if err != nil {
		return err
	}

	for _, a := range assets {
		if err := uploadAsset(ctx, client, api, rel.ID, a, token, scheme); err != nil {
			return err
		}
		log.Info(fmt.Sprintf("uploaded %s", a.name))
	}

	if err := c.Output("id", fmt.Sprintf("%d", rel.ID)); err != nil {
		return err
	}
	if err := c.Output("url", rel.HTMLURL); err != nil {
		return err
	}
	if err := c.Output("assets", fmt.Sprintf("%d", len(assets))); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("release %s (%d asset(s)) at %s", tag, len(assets), rel.HTMLURL))
	return nil
}

// resolveToken picks the credential: the explicit token input, an
// authorized-integration ID token minted for the audience input, then
// the job token. The returned scheme is the Authorization header
// prefix each kind of token is presented with.
func resolveToken(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) (token, scheme string, err error) {
	switch {
	case in.Token != "":
		return in.Token, "token ", nil
	case in.Audience != "":
		jwt, err := oidc.Token(ctx, log, c.IDTokenURL, c.IDTokenToken, in.Audience)
		if err != nil {
			return "", "", err
		}
		log.Info(fmt.Sprintf("using an authorized integration ID token (audience %s)", in.Audience))
		return jwt, "Bearer ", nil
	case c.Token != "":
		return c.Token, "token ", nil
	}
	return "", "", fmt.Errorf("no token; set the token or audience input (or provide GITHUB_TOKEN/GITEA_TOKEN)")
}

// createReleaseOption is the POST /releases payload.
type createReleaseOption struct {
	TagName      string `json:"tag_name"`
	Target       string `json:"target_commitish,omitempty"`
	Title        string `json:"name"`
	Note         string `json:"body"`
	IsDraft      bool   `json:"draft"`
	IsPrerelease bool   `json:"prerelease"`
}

// releaseOut is the release object the API answers with, reduced to
// what the action reports.
type releaseOut struct {
	ID      int64  `json:"id"`
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// findRelease looks a release up by tag; not found is not an error.
func findRelease(ctx context.Context, client *http.Client, api, tag, token, scheme string) (releaseOut, bool, error) {
	var rel releaseOut
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		api+"/releases/tags/"+url.PathEscape(tag), nil)
	if err != nil {
		return rel, false, err
	}
	req.Header.Set("Authorization", scheme+token)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return rel, false, fmt.Errorf("looking up release %s: %w", tag, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	switch {
	case resp.StatusCode == http.StatusOK:
		if err := json.Unmarshal(body, &rel); err != nil {
			return rel, false, fmt.Errorf("release %s was not valid JSON: %w", tag, err)
		}
		return rel, true, nil
	case resp.StatusCode == http.StatusNotFound:
		return rel, false, nil
	}
	return rel, false, &statusError{op: "GET release " + tag, status: resp.StatusCode, body: string(body)}
}

// deleteRelease removes a release (the tag itself is kept).
func deleteRelease(ctx context.Context, client *http.Client, api string, id int64, token, scheme string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/releases/%d", api, id), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", scheme+token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("deleting release %d: %w", id, err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return &statusError{op: fmt.Sprintf("DELETE release %d", id), status: resp.StatusCode}
	}
	return nil
}

// createRelease posts the new release and returns its id and URL.
func createRelease(ctx context.Context, client *http.Client, api, token, scheme string, opt createReleaseOption) (releaseOut, error) {
	var rel releaseOut
	payload, err := json.Marshal(opt)
	if err != nil {
		return rel, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api+"/releases", bytes.NewReader(payload))
	if err != nil {
		return rel, err
	}
	req.Header.Set("Authorization", scheme+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return rel, fmt.Errorf("creating release %s: %w", opt.TagName, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusCreated {
		return rel, &statusError{op: "POST release " + opt.TagName, status: resp.StatusCode, body: apiMessage(body)}
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		return rel, fmt.Errorf("created release %s but the answer was not valid JSON: %w", opt.TagName, err)
	}
	return rel, nil
}

// uploadAsset attaches one file to the release, streamed from disk.
func uploadAsset(ctx context.Context, client *http.Client, api string, releaseID int64, a asset, token, scheme string) error {
	f, err := os.Open(a.path)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/releases/%d/assets?name=%s", api, releaseID, url.QueryEscape(a.name)), f)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", scheme+token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("uploading %s: %w", a.name, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusCreated {
		return &statusError{op: "upload asset " + a.name, status: resp.StatusCode, body: apiMessage(body)}
	}
	return nil
}

// apiMessage digs the human-readable message out of Forgejo's error
// JSON, falling back to the raw body.
func apiMessage(body []byte) string {
	var resp struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err == nil && resp.Message != "" {
		return resp.Message
	}
	return strings.TrimSpace(string(body))
}

// parseAssets turns the assets input — whitespace/newline-separated
// entries of "path" or "name=path" — into the upload list, resolving
// paths against the workspace and defaulting names to the path's
// basename. Duplicate asset names are an error: Forgejo would attach
// both and the download URL would pick one arbitrarily.
func parseAssets(ws, in string) ([]asset, error) {
	var out []asset
	seen := map[string]bool{}
	for field := range strings.FieldsSeq(in) {
		path, name := field, ""
		if before, after, ok := strings.Cut(field, "="); ok {
			path, name = after, before
			if name == "" {
				return nil, fmt.Errorf("asset %q has an empty name", field)
			}
			if strings.ContainsAny(name, `/\`) {
				return nil, fmt.Errorf("asset name %q must not contain path separators", name)
			}
		}
		if path == "" {
			return nil, fmt.Errorf("asset %q has an empty path", field)
		}
		if name == "" {
			name = filepath.Base(path)
		}
		if seen[name] {
			return nil, fmt.Errorf("two assets are named %q; use name=path to disambiguate", name)
		}
		seen[name] = true
		if !filepath.IsAbs(path) {
			path = filepath.Join(ws, path)
		}
		out = append(out, asset{path: path, name: name})
	}
	return out, nil
}

// tagFromRef returns the tag name of a refs/tags/... ref, or "" when
// the ref is not a tag.
func tagFromRef(ref string) string {
	if tag, ok := strings.CutPrefix(ref, "refs/tags/"); ok {
		return tag
	}
	return ""
}
