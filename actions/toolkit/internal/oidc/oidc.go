// Package oidc fetches OIDC ID tokens from the runner's ID token
// endpoint — the JWTs Forgejo Authorized Integrations are built on.
// The runner only exposes the endpoint (ACTIONS_ID_TOKEN_REQUEST_URL
// plus a request token) to workflows that set enable-openid-connect.
package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Token fetches an ID token minted for audience from the runner
// endpoint addressed by requestURL and requestToken (the
// ACTIONS_ID_TOKEN_REQUEST_URL / ACTIONS_ID_TOKEN_REQUEST_TOKEN pair).
// The token is returned cleaned: anything outside the base64url+dot
// alphabet an embedded newline could smuggle in is stripped, and the
// number of stripped characters is logged when non-zero.
func Token(ctx context.Context, log *slog.Logger, requestURL, requestToken, audience string) (string, error) {
	jwt, err := fetch(ctx, http.DefaultClient, requestURL, requestToken, audience)
	if err != nil {
		return "", err
	}
	clean, stripped := Clean(jwt)
	if stripped > 0 && log != nil {
		log.Warn(fmt.Sprintf("stripped %d non-base64url character(s) from the ID token", stripped))
	}
	if clean == "" {
		return "", fmt.Errorf("ID token response contained no usable token")
	}
	return clean, nil
}

// fetch performs the endpoint request and extracts the token from the
// JSON response.
func fetch(ctx context.Context, client *http.Client, requestURL, requestToken, audience string) (string, error) {
	if requestURL == "" || requestToken == "" {
		return "", fmt.Errorf("runner provided no ID token endpoint; set enable-openid-connect: true in the workflow to get one")
	}
	sep := "?"
	if strings.Contains(requestURL, "?") {
		sep = "&"
	}
	u := requestURL + sep + "audience=" + url.QueryEscape(audience)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ID token request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("ID token request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ID token request failed: %s", resp.Status)
	}
	return Parse(body)
}

// Parse extracts the JWT from an ID token endpoint response body.
func Parse(body []byte) (string, error) {
	var payload struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("ID token response was not valid JSON: %w", err)
	}
	if payload.Value == "" {
		return "", fmt.Errorf("ID token response contained no token")
	}
	return payload.Value, nil
}

// Clean strips everything outside the base64url+dot alphabet a JWT is
// written in — the runner cannot be trusted to keep the value on one
// line, and an embedded newline makes consumers of the token (git's
// curl, for one) abort with "a libcurl function was given a bad
// argument". It returns the cleaned token and how many characters were
// removed.
func Clean(jwt string) (string, int) {
	clean := strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			return r
		}
		return -1
	}, jwt)
	return clean, len(jwt) - len(clean)
}
