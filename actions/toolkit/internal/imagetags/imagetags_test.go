package imagetags

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in   string
		want Version
		ok   bool
	}{
		{"1", Version{Major: 1}, true},
		{"1.2", Version{Major: 1, Minor: 2}, true},
		{"1.2.3", Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"v1.2.3", Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"go1.22.5", Version{Major: 1, Minor: 22, Patch: 5}, true},
		{"1.2.3-rc1", Version{Major: 1, Minor: 2, Patch: 3, Pre: "rc1"}, true},
		{"1.2.3-rc-1", Version{Major: 1, Minor: 2, Patch: 3, Pre: "rc-1"}, true},
		{"1.2.3+build.7", Version{Major: 1, Minor: 2, Patch: 3}, true},
		{" v1.2.3 ", Version{Major: 1, Minor: 2, Patch: 3}, true},
		{"007", Version{Major: 7}, true}, // Atoi semantics
		{"1.2.3.4", Version{}, false},
		{"", Version{}, false},
		{"v", Version{}, false},
		{"1.x", Version{}, false},
		{"1.2.", Version{}, false},
		{"-1", Version{}, false},
		{"1.2.3-", Version{Major: 1, Minor: 2, Patch: 3}, true}, // empty pre-release
	}
	for _, tc := range cases {
		got, ok := ParseVersion(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("ParseVersion(%q) = %+v, %v; want %+v, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestDeriveTags(t *testing.T) {
	cases := []struct {
		ref    string
		latest string
		want   []string
	}{
		{"refs/heads/master", "", []string{"dev"}},
		{"refs/heads/main", "", []string{"dev"}},
		{"refs/tags/v1.2.3", "", []string{"1.2.3"}},
		{"refs/tags/v1.2.3", "1.2.3", []string{"latest", "1.2.3"}},
		{"refs/tags/v1.2.3", "9.9.9", []string{"1.2.3"}},
		{"refs/tags/v1.2.3-rc1", "1.9.9", []string{"latest", "1.2.3"}}, // pre dropped, major still matches
		{"refs/tags/not-a-version", "", nil},
		{"refs/heads/topic", "", nil},
		{"refs/merge-request/7/head", "", nil},
	}
	for _, tc := range cases {
		got := DeriveTags(tc.ref, tc.latest)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("DeriveTags(%q, %q) = %v, want %v", tc.ref, tc.latest, got, tc.want)
		}
	}
}

func TestRefVersion(t *testing.T) {
	if got := RefVersion("refs/tags/v2.3.4-rc1", ""); got != "2.3.4" {
		t.Errorf("got %q", got)
	}
	if got := RefVersion("refs/heads/master", "abc123"); got != "abc123" {
		t.Errorf("got %q", got)
	}
	if got := RefVersion("refs/heads/topic", "abc123"); got != "unknown" {
		t.Errorf("got %q", got)
	}
}

func TestLatestStableMajor(t *testing.T) {
	cases := []struct {
		names []string
		want  string
	}{
		{nil, ""},
		{[]string{"topic"}, ""},
		{[]string{"v1.2.3", "v1.10.0", "v2.0.0-rc1"}, "1"}, // pre-releases are ignored
		{[]string{"v0.9.0", "1.0"}, "1"},
		{[]string{"v10.0.0", "v9.99.99"}, "10"},
	}
	for _, tc := range cases {
		if got := LatestStableMajor(tc.names); got != tc.want {
			t.Errorf("LatestStableMajor(%v) = %q, want %q", tc.names, got, tc.want)
		}
	}
}

func TestListRepoTags(t *testing.T) {
	var page int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		switch page {
		case 1:
			if got := r.Header.Get("Authorization"); got != "token tok" {
				t.Errorf("authorization header wrong: %q", got)
			}
			_, _ = w.Write([]byte(`[{"name":"v1.0.0"},{"name":"v0.9.0"}]`))
		case 2:
			_, _ = w.Write([]byte(`[{"name":"not-json-objects-will-just-fail-later"}]`))
			w.WriteHeader(http.StatusBadGateway)
		default:
			t.Errorf("unexpected page %d", page)
		}
	}))
	defer srv.Close()

	// Page two fails; whatever page one delivered is kept, without error.
	tags, err := ListRepoTags(t.Context(), slog.Default(), srv.URL, "owner/repo", "tok")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tags, []string{"v1.0.0", "v0.9.0"}) {
		t.Errorf("got %v", tags)
	}
}

func TestListRepoTagsNoAPI(t *testing.T) {
	if _, err := ListRepoTags(t.Context(), nil, "", "", ""); err == nil {
		t.Error("missing API URL should error")
	}
}
