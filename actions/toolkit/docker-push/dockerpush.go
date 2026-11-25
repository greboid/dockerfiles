// Package dockerpush builds OCI images with buildah in the action
// container — one platform by default, a multi-platform manifest list
// when the platforms input names several — and pushes each tag
// straight to its registry, optionally mirroring to GHCR: a
// replacement for docker/build-push-action on Forgejo Runner. With no
// tags input, tags are derived from the workflow ref (the rules of
// greboid/docker-tags-action, see internal/imagetags) as
// <registry>/<owner>/<name>:<version> against the registries input,
// defaulting to the Forgejo instance of this workflow.
package dockerpush

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/buildah"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/imagetags"
)

// Inputs carries the action inputs; secrets arrive through the
// environment (Token from TOKEN, Password from PASSWORD, GHCRToken
// from GHCR_TOKEN), the rest as command line flags.
type Inputs struct {
	Tags           string
	Registries     string
	Repository     string
	Token          string
	Context        string
	File           string
	Target         string
	Platforms      string
	BuildArgs      string
	Push           string
	NoCache        string
	Registry       string
	Username       string
	Password       string
	GHCR           string
	GHCRRepository string
	GHCRToken      string
	GHCRUsername   string
}

// Run builds the image, pushes its tags and reports the results.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	gha.Mask(in.Password)
	gha.Mask(in.GHCRToken)
	gha.Mask(in.Token)

	tags := strings.TrimSpace(in.Tags)
	if tags == "" {
		var err error
		tags, err = deriveTags(ctx, c, log, in)
		if err != nil {
			return err
		}
		log.Info("derived tags from " + c.Ref + ": " + tags)
	}
	tagList := strings.Fields(tags)

	version := imagetags.RefVersion(c.Ref, c.SHA)

	contextDir, relFile, err := buildah.ResolveBuildFile(c.Workspace, in.Context, in.File)
	if err != nil {
		return err
	}

	// Registry authentication: the docker config docker-login wrote
	// (via DOCKER_CONFIG), or an explicit registry/username/password
	// login.
	authfile := os.Getenv("REGISTRY_AUTH_FILE")
	if authfile == "" {
		if dockerConfig := os.Getenv("DOCKER_CONFIG"); dockerConfig != "" {
			authfile = filepath.Join(dockerConfig, "config.json")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			authfile = filepath.Join(home, ".docker", "config.json")
		}
	}
	if err := os.MkdirAll(filepath.Dir(authfile), 0o755); err != nil {
		return err
	}
	if !fileExists(authfile) {
		if err := os.WriteFile(authfile, []byte("{}\n"), 0o600); err != nil {
			return err
		}
	}
	log.Info("using registry auth: " + authfile)

	if in.Registry != "" && in.Username != "" && in.Password != "" {
		if err := buildah.Login(ctx, authfile, in.Registry, in.Username, in.Password); err != nil {
			return fmt.Errorf("login to %s failed (invalid credentials?)", in.Registry)
		}
	}

	tool, err := buildah.Prepare(ctx, log)
	if err != nil {
		return err
	}

	// Build. More than one platform builds once per platform into a
	// shared manifest list, which is what the tags then point at; a
	// single (or unset) platform keeps the plain single-image flow.
	platforms := buildah.SplitRefs(in.Platforms)
	if len(platforms) == 0 {
		platforms = []string{""} // unset: build the host platform once
	}
	manifest := ""
	if len(platforms) > 1 {
		manifest = buildah.UniqueName("action-manifest")
		defer tool.RemoveManifest(ctx, manifest)
		log.Info(fmt.Sprintf("building platforms %s into manifest list %s", strings.Join(platforms, ", "), manifest))
	}

	iidFile, err := os.CreateTemp("", "iid-*")
	if err != nil {
		return err
	}
	iidPath := iidFile.Name()
	iidFile.Close()
	defer os.Remove(iidPath)

	imageid := ""
	for _, platform := range platforms {
		bud := buildah.BudOptions{
			AuthFile:  authfile,
			IIDFile:   iidPath,
			Tags:      tagList,
			RelFile:   relFile,
			Target:    in.Target,
			NoCache:   gha.Truthy(in.NoCache, false),
			BuildArgs: buildArgLines(in.BuildArgs),
			Context:   contextDir,
		}
		if manifest != "" {
			bud.Manifest = manifest
		}
		if platform != "" {
			bud.Platform = platform
		}
		if err := tool.Bud(ctx, bud); err != nil {
			return err
		}

		rawID, err := os.ReadFile(iidPath)
		if err != nil {
			return err
		}
		imageid = strings.TrimSpace(string(rawID))
		if !strings.Contains(imageid, ":") {
			imageid = "sha256:" + imageid
		}
		if platform != "" {
			log.Info("built image " + imageid + " for " + platform)
		} else {
			log.Info("built image " + imageid)
		}
	}

	// Push each tag; the digest of the first push becomes the digest
	// output — the manifest list digest in the multi-platform flow.
	digest := ""
	push := gha.Truthy(in.Push, true)
	if push {
		first := true
		for _, tag := range tagList {
			log.Info("pushing " + tag)
			digestFile, err := os.CreateTemp("", "digest-*")
			if err != nil {
				return err
			}
			digestPath := digestFile.Name()
			digestFile.Close()
			if manifest != "" {
				err = tool.ManifestPush(ctx, buildah.ManifestPushOptions{
					AuthFile:   authfile,
					DigestFile: digestPath,
					List:       manifest,
					Dest:       tag,
				})
			} else {
				// By ID: a concurrent job may retag this name in a shared store.
				err = tool.Run(ctx, "push", "--authfile", authfile, "--digestfile", digestPath, imageid, tag)
			}
			if content, readErr := os.ReadFile(digestPath); readErr == nil && first && len(content) > 0 {
				digest = strings.TrimSpace(string(content))
			}
			os.Remove(digestPath)
			if err != nil {
				return err
			}
			first = false
			log.Info("pushed " + tag)
		}
	} else {
		log.Info("push disabled")
	}

	// Optional GHCR mirror: each distinct tag suffix once, regardless
	// of how many registries the tags list names it for.
	if gha.Truthy(in.GHCR, false) {
		if err := mirrorToGHCR(ctx, c, log, tool, in, tagList, authfile, push, manifest, imageid); err != nil {
			return err
		}
	}

	// Outputs.
	if err := c.Output("imageid", imageid); err != nil {
		return err
	}
	if err := c.Output("digest", digest); err != nil {
		return err
	}
	if err := c.Output("version", version); err != nil {
		return err
	}
	if err := c.Output("tags", strings.Join(tagList, ",")); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("done: version=%s imageid=%s digest=%s", version, imageid, digest))
	return nil
}

// deriveTags builds the tag list from the workflow ref when the tags
// input is empty: <registry>/<image>:<version> for every derived
// version across every registry.
func deriveTags(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) (string, error) {
	registries := buildah.SplitRefs(in.Registries)
	if len(registries) == 0 {
		if c.Host() == "" {
			return "", fmt.Errorf("no tags input and no registries input; GITHUB_SERVER_URL would supply the default registry but is not set")
		}
		registries = []string{c.Host()}
	}
	image := in.Repository
	if image == "" {
		if c.Repository == "" {
			return "", fmt.Errorf("no tags input and no repository input; GITHUB_REPOSITORY would supply the default image name but is not set")
		}
		image = c.Repository // instance registries namespace under owner/name
	}
	if c.Ref == "" {
		return "", fmt.Errorf("no tags input and GITHUB_REF is not set")
	}

	// Only the "latest" rule needs the repository's tag list; master
	// and main builds are just "dev".
	latest := ""
	if strings.HasPrefix(c.Ref, "refs/tags/") {
		token := in.Token
		if token == "" {
			token = c.Token
		}
		if token != "" {
			names, err := imagetags.ListRepoTags(ctx, log, c.APIURL, c.Repository, token)
			if err != nil {
				return "", err
			}
			latest = imagetags.LatestStableMajor(names)
		} else {
			log.Warn("no token to list repository tags; \"latest\" cannot be determined")
		}
	}

	versions := imagetags.DeriveTags(c.Ref, latest)
	if len(versions) == 0 {
		return "", fmt.Errorf("ref %s derives no image tags (only master/main and semver tags do); set the tags input explicitly", c.Ref)
	}

	var tags []string
	for _, reg := range registries {
		for _, v := range versions {
			tags = append(tags, reg+"/"+image+":"+v)
		}
	}
	return strings.Join(tags, " "), nil
}

// mirrorToGHCR pushes the tags' suffixes to ghcr.io/<repository>; in
// the multi-platform flow it pushes the manifest list itself, in the
// single-platform flow it tags the built image.
func mirrorToGHCR(ctx context.Context, c *gha.Context, log *slog.Logger, tool *buildah.Toolchain, in Inputs, tagList []string, authfile string, push bool, manifest, imageid string) error {
	ghcrRepo := in.GHCRRepository
	if ghcrRepo == "" {
		if c.Repository == "" {
			return fmt.Errorf("GITHUB_REPOSITORY is not set and the ghcr-repository input was not provided")
		}
		ghcrRepo = strings.ToLower(c.Repository)
	}
	ghcrUser := in.GHCRUsername
	if ghcrUser == "" {
		ghcrUser = c.Actor
	}
	ghcrAuthfile := authfile
	if in.GHCRToken != "" {
		if ghcrUser == "" {
			return fmt.Errorf("the ghcr-username input (or GITHUB_ACTOR) is required when ghcr-token is set")
		}
		tmp, err := os.CreateTemp("", "ghcr-auth-*")
		if err != nil {
			return err
		}
		ghcrAuthfile = tmp.Name()
		defer os.Remove(ghcrAuthfile)
		tmp.Close()
		// buildah login reads the existing authfile to merge into, and
		// an empty one is not valid JSON.
		if err := os.WriteFile(ghcrAuthfile, []byte("{}\n"), 0o600); err != nil {
			return err
		}
		if err := buildah.Login(ctx, ghcrAuthfile, "ghcr.io", ghcrUser, in.GHCRToken); err != nil {
			return fmt.Errorf("ghcr login failed (invalid credentials?)")
		}
	}
	seen := map[string]bool{}
	for _, tag := range tagList {
		suffix := buildah.TagSuffix(tag)
		if seen[suffix] {
			continue
		}
		seen[suffix] = true
		ghcrRef := "ghcr.io/" + ghcrRepo + ":" + suffix
		log.Info("mirroring " + tag + " -> " + ghcrRef)
		if manifest != "" {
			if push {
				if err := tool.ManifestPush(ctx, buildah.ManifestPushOptions{
					AuthFile: ghcrAuthfile,
					List:     manifest,
					Dest:     ghcrRef,
				}); err != nil {
					return err
				}
			}
			continue
		}
		if err := tool.Run(ctx, "tag", imageid, ghcrRef); err != nil {
			return err
		}
		if push {
			if err := tool.Run(ctx, "push", "--authfile", ghcrAuthfile, ghcrRef); err != nil {
				return err
			}
		}
	}
	return nil
}

// buildArgLines splits the build-args value into one argument per
// non-empty line, skipping lines that are not KEY=VALUE.
func buildArgLines(s string) []string {
	var out []string
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") {
			out = append(out, line)
		}
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
