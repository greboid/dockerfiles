package oidc

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParse(t *testing.T) {
	tok, err := Parse([]byte(`{"value":"aaa.bbb.ccc","count":1}`))
	if err != nil || tok != "aaa.bbb.ccc" {
		t.Errorf("got %q, %v", tok, err)
	}
	if _, err := Parse([]byte(`{"count":1}`)); err == nil {
		t.Error("response without a token should error")
	}
	if _, err := Parse([]byte(`nope`)); err == nil {
		t.Error("non-JSON response should error")
	}
}

func TestClean(t *testing.T) {
	clean, stripped := Clean("ab\ncd ef.gh")
	if clean != "abcdef.gh" || stripped != 2 {
		t.Errorf("got %q (%d stripped), want abcdef.gh (2)", clean, stripped)
	}
	clean, stripped = Clean("untouched.token")
	if clean != "untouched.token" || stripped != 0 {
		t.Errorf("clean token should pass through, got %q (%d)", clean, stripped)
	}
}

func TestToken(t *testing.T) {
	var gotAuth, gotAudience string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAudience = r.URL.Query().Get("audience")
		w.Header().Set("Content-Type", "application/json")
		// The escaped newlines decode into the string value, the way a
		// misbehaving endpoint would deliver them.
		_, _ = w.Write([]byte(`{"value":"hd.\n\njwt"}`))
	}))
	defer srv.Close()

	tok, err := Token(context.Background(), slog.Default(), srv.URL, "req-token", "my audience")
	if err != nil {
		t.Fatal(err)
	}
	if tok != "hd.jwt" {
		t.Errorf("token not cleaned, got %q", tok)
	}
	if gotAuth != "Bearer req-token" {
		t.Errorf("authorization header wrong: %q", gotAuth)
	}
	if gotAudience != "my audience" {
		t.Errorf("audience not passed: %q", gotAudience)
	}
}

func TestTokenWithoutEndpoint(t *testing.T) {
	if _, err := Token(context.Background(), nil, "", "", "aud"); err == nil {
		t.Error("missing endpoint should error")
	}
}

func TestTokenServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	}))
	defer srv.Close()

	if _, err := Token(context.Background(), nil, srv.URL, "tok", "aud"); err == nil {
		t.Error("non-200 response should error")
	}
}
