// Package dockerlogin logs registry credentials into the docker
// config that buildah — and the docker CLI — read, verifying them
// against the registry (Registry HTTP API v2): a replacement for
// docker/login-action on Forgejo Runner. The config is exported via
// DOCKER_CONFIG so later steps — and docker-push — pick the
// credentials up.
//
// Credentials come from explicit inputs, from a Forgejo Authorized
// Integration (an OIDC JWT fetched from the runner, used as the
// registry password) or, with no inputs at all, from the workflow's
// job token.
package dockerlogin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/buildah"
	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/oidc"
)

// Inputs carries the action inputs; secrets arrive through the
// environment (Password from PASSWORD, GHCR from GHCR), the rest as
// command line flags.
type Inputs struct {
	Registry string
	Username string
	Password string
	Audience string
	GHCR     string
}

// Run logs in to the registry and exports the docker config directory.
func Run(ctx context.Context, c *gha.Context, log *slog.Logger, in Inputs) error {
	gha.Mask(in.Password)
	gha.Mask(in.GHCR)

	registry, username, password := in.Registry, in.Username, in.Password

	if in.Audience != "" && password == "" {
		// Authorized Integration mode: fetch an OIDC JWT from the
		// runner and use it as the registry password. Forgejo v16+
		// accepts the JWT in the basic-auth password field for its
		// container registry; the username is ignored.
		jwt, err := oidc.Token(ctx, log, c.IDTokenURL, c.IDTokenToken, in.Audience)
		if err != nil {
			return err
		}
		password = jwt
		gha.Mask(password)
		if username == "" {
			username = "jwt"
		}
		if registry == "" {
			if c.Host() == "" {
				return fmt.Errorf("registry input is required when GITHUB_SERVER_URL is not set")
			}
			registry = c.Host()
		}
		log.Info(fmt.Sprintf("using an authorized integration ID token (audience %s) to log in to %s", in.Audience, registry))
	}

	if registry == "" && username == "" && password == "" {
		// Zero-config mode: log in to the Forgejo instance the
		// workflow runs on, using the job token and the repository
		// owner.
		if c.Host() == "" || c.Token == "" {
			return fmt.Errorf("registry, username and password inputs are required")
		}
		registry = c.Host()
		username = gha.FirstNonEmpty(c.Owner(), c.Actor, "git")
		password = c.Token
		log.Info(fmt.Sprintf("no inputs given; logging in to %s as %s with GITHUB_TOKEN", registry, username))
	}
	if username == "" || password == "" {
		return fmt.Errorf("both username and password are required")
	}
	if registry == "" {
		registry = "docker.io"
	}

	configDir := os.Getenv("DOCKER_CONFIG")
	if configDir == "" {
		configDir = c.ResolvePath(filepath.Join(".actions", "docker"))
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	authfile := filepath.Join(configDir, "config.json")

	if err := login(ctx, authfile, registry, username, password, ""); err != nil {
		return err
	}

	// A second login when a GitHub PAT is given: store ghcr.io
	// credentials in the same config, which is what docker-push's
	// ghcr mirror reads. GHCR authenticates on the token alone, so
	// the username is cosmetic.
	if in.GHCR != "" && NormalizeRegistry(registry) != "ghcr.io" {
		ghcrUser := gha.FirstNonEmpty(c.Owner(), c.Actor, "ghcr")
		if err := login(ctx, authfile, "ghcr.io", ghcrUser, in.GHCR, "ghcr login"); err != nil {
			return err
		}
		log.Info("also logged in to ghcr.io as " + ghcrUser)
	}

	// buildah stores Docker Hub credentials under the docker.io key;
	// rewrite it to the legacy well-known one so the docker CLI finds
	// them too (buildah reads both).
	switch NormalizeRegistry(registry) {
	case "docker.io", "registry-1.docker.io", "index.docker.io":
		if err := legacyDockerIOKey(authfile); err != nil {
			return err
		}
	}

	if err := c.ExportEnv("DOCKER_CONFIG", configDir); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("logged in as %s; credentials in %s (DOCKER_CONFIG exported)", username, authfile))
	return nil
}

// login runs buildah login, feeding the password over stdin. label
// names the operation in the error ("" means "login to <registry>").
func login(ctx context.Context, authfile, registry, username, password, label string) error {
	if err := buildah.Login(ctx, authfile, registry, username, password); err != nil {
		if label == "" {
			label = "login to " + registry
		}
		return fmt.Errorf("%s failed (invalid credentials?)", label)
	}
	return nil
}

// legacyDockerIOKey rewrites the authfile's docker.io key to the
// legacy well-known one, in place, preserving everything else.
func legacyDockerIOKey(authfile string) error {
	b, err := os.ReadFile(authfile)
	if err != nil {
		return err
	}
	var config struct {
		Auths map[string]json.RawMessage `json:"auths"`
	}
	if err := json.Unmarshal(b, &config); err != nil {
		return fmt.Errorf("authfile %s is not valid JSON: %w", authfile, err)
	}
	raw, ok := config.Auths["docker.io"]
	if !ok {
		return nil
	}
	delete(config.Auths, "docker.io")
	config.Auths["https://index.docker.io/v1/"] = raw
	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(authfile, out, 0o600)
}

// NormalizeRegistry strips any scheme and path from a registry
// address, mapping empty input to docker.io.
func NormalizeRegistry(reg string) string {
	if i := strings.Index(reg, "://"); i >= 0 {
		reg = reg[i+3:]
	}
	if i := strings.IndexByte(reg, '/'); i >= 0 {
		reg = reg[:i]
	}
	if reg == "" {
		return "docker.io"
	}
	return reg
}
