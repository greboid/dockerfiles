// Package dockerbuild builds an OCI image with buildah in the action
// container — the reference OCI builder, so no Docker Engine compat
// API quirks. The platforms input selects the target platforms. It is
// the build-only sibling of docker-push: when tags were given, a
// single-platform result is pushed into the runner's engine over
// DOCKER_HOST so later steps can use it; an untagged build is only
// reported via the imageid output. A multi-platform build is never
// loaded, because the docker-daemon transport takes only a single
// image. Nothing is pushed to a registry and no registry credentials
// are read.
package dockerbuild

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/buildah"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
)

// Inputs carries the action inputs; the build-args value is one
// KEY=VALUE per line (values may contain spaces).
type Inputs struct {
	Tags      string
	Context   string
	File      string
	Target    string
	Platforms string
	BuildArgs string
	NoCache   string
}

// Run builds the image and reports its ID.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	tags := buildah.SplitRefs(in.Tags)

	platforms := buildah.SplitRefs(in.Platforms)
	if len(platforms) == 0 {
		platforms = []string{""} // unset: build the host platform once
	}

	contextDir, relFile, err := buildah.ResolveBuildFile(c.Workspace, in.Context, in.File)
	if err != nil {
		return err
	}

	tool, err := buildah.Prepare(ctx, log)
	if err != nil {
		return err
	}

	// Without tags the image still gets an internal one, so it can be
	// inspected and pushed into the engine.
	refs := tags
	if len(refs) == 0 {
		refs = []string{buildah.UniqueName("action-untagged")}
	}

	manifest := ""
	if len(platforms) > 1 {
		manifest = buildah.UniqueName("action-manifest")
		defer tool.RemoveManifest(ctx, manifest)
		log.Info(fmt.Sprintf("building platforms %s into manifest list %s", strings.Join(platforms, ", "), manifest))
	}

	for _, platform := range platforms {
		err = tool.Bud(ctx, buildah.BudOptions{
			Tags:      refs,
			RelFile:   relFile,
			Platform:  platform,
			Manifest:  manifest,
			Target:    in.Target,
			NoCache:   gha.Truthy(in.NoCache, false),
			BuildArgs: buildArgLines(in.BuildArgs),
			Context:   contextDir,
		})
		if err != nil {
			return err
		}
	}

	// Each build retags refs, so this is the last platform's image.
	imageid, err := tool.Output(ctx, "inspect", "--format", "{{.FromImageID}}", refs[0])
	if err != nil {
		return err
	}
	if !strings.Contains(imageid, ":") {
		imageid = "sha256:" + imageid
	}
	if err := c.Output("imageid", imageid); err != nil {
		return err
	}

	// With tags the image lives in this container's storage; hand it
	// to the runner's engine so later steps can use it. Without tags
	// there is nothing to load (docker-daemon references need a tag),
	// so the imageid output is all later steps get.
	switch {
	case manifest != "":
		if len(tags) > 0 {
			log.Warn("multi-platform builds cannot be loaded into the runner engine; the tags will not be available to later steps")
		}
		log.Info("image " + imageid + " built (multi-platform; not loaded into the engine)")
	case len(tags) > 0:
		for _, ref := range refs {
			if err := tool.Run(ctx, "push", "--format", "oci", ref, "docker-daemon:"+ref); err != nil {
				return fmt.Errorf("buildah push %s to daemon failed: %w", ref, err)
			}
		}
		log.Info("image " + imageid + " built")
	default:
		log.Info("image " + imageid + " built (untagged; not loaded into the engine)")
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
