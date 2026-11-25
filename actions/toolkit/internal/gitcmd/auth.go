// Authentication of git CLI invocations against the workflow
// instance, without persisting the token: it never touches
// .git/config or argv, so later git calls have no credentials to
// piggyback on and nothing leaks through them.
//
// Static tokens ride a chmod 700 GIT_ASKPASS helper that answers git's
// username/password prompts; an Authorized Integration JWT is instead
// sent as an HTTP Bearer header (http.extraheader exported through
// GIT_CONFIG_*, git 2.31+), the way Forgejo's git endpoint expects
// JWTs to be presented.
package gitcmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/oidc"
)

// Auth carries git authentication for one action run.
type Auth struct {
	env     []string // environment additions for git invocations
	askpass string   // path of the throwaway askpass helper, when used
}

// Setup resolves the authentication token — the explicit token input,
// an authorized-integration ID token minted for audience, then the job
// token — masks it, and prepares the environment additions that
// authenticate git CLI invocations. Cleanup must be called when the
// action is done.
func Setup(ctx context.Context, c *gha.Context, log *slog.Logger, token, audience, username string) (*Auth, error) {
	bearer := false
	if token == "" && audience != "" {
		jwt, err := oidc.Token(ctx, log, c.IDTokenURL, c.IDTokenToken, audience)
		if err != nil {
			return nil, err
		}
		token = jwt
		bearer = true
	}
	if token == "" {
		token = c.Token
	}
	if token == "" {
		return nil, fmt.Errorf("no token for git authentication; set the token or audience input (or provide GITHUB_TOKEN/GITEA_TOKEN)")
	}
	if username == "" {
		username = c.Actor
	}
	if username == "" {
		username = "x-access-token"
	}
	gha.Mask(token)

	env := []string{"GIT_TERMINAL_PROMPT=0"}
	a := &Auth{env: env}
	if bearer {
		// Export http.extraheader through the numbered GIT_CONFIG_*
		// channel, honoring whatever entries the environment already
		// carries.
		n, _ := strconv.Atoi(os.Getenv("GIT_CONFIG_COUNT"))
		a.env = append(a.env,
			"GIT_CONFIG_COUNT="+strconv.Itoa(n+1),
			"GIT_CONFIG_KEY_"+strconv.Itoa(n)+"=http.extraheader",
			"GIT_CONFIG_VALUE_"+strconv.Itoa(n)+"=Authorization: Bearer "+token,
		)
		return a, nil
	}

	// Static token: a throwaway askpass helper answers git's prompts.
	// It lives only for the run and is removed by Cleanup.
	file, err := os.CreateTemp(os.TempDir(), "git-askpass-*")
	if err != nil {
		return nil, err
	}
	if _, err := file.WriteString(askpassScript(username, token)); err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}
	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return nil, err
	}
	if err := os.Chmod(file.Name(), 0o700); err != nil {
		os.Remove(file.Name())
		return nil, err
	}
	a.askpass = file.Name()
	a.env = append(a.env, "GIT_ASKPASS="+file.Name())
	return a, nil
}

// Environ returns the environment additions authenticating git
// invocations; extend exec.Cmd.Env with them.
func (a *Auth) Environ() []string { return a.env }

// Cleanup removes the throwaway askpass helper. It is a no-op when
// Setup used a bearer token instead.
func (a *Auth) Cleanup() {
	if a != nil && a.askpass != "" {
		os.Remove(a.askpass)
	}
}

// askpassScript renders the sh script answering git's credential
// prompts with the username and token.
func askpassScript(username, token string) string {
	return "#!/bin/sh\ncase \"$1\" in\n" +
		"\t*Username*) printf %s " + shQuote(username) + " ;;\n" +
		"\t*Password*) printf %s " + shQuote(token) + " ;;\n" +
		"\t*) printf \"\\n\" ;;\nesac\n"
}

// shQuote single-quotes s for embedding in a sh script.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
