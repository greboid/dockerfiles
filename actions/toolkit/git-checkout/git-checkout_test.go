package gitcheckout

import (
	"reflect"
	"strings"
	"testing"
)

func TestCandidateRefs(t *testing.T) {
	sha := strings.Repeat("a", 40)

	cases := []struct {
		ref  string
		want []string
	}{
		{"refs/heads/main", []string{"refs/heads/main", sha}},
		{"refs/tags/v1.2.3", []string{"refs/tags/v1.2.3", sha}},
		{"refs/pull/7/merge", []string{"refs/pull/7/merge", "refs/pull/7/head", sha}},
		{"refs/pull/7/head", []string{"refs/pull/7/head", "refs/pull/7/merge", sha}},
		{"main", []string{"main", "refs/heads/main", "refs/tags/main", sha}},
		{"feature/x", []string{"feature/x", "refs/heads/feature/x", "refs/tags/feature/x", sha}},
	}
	for _, c := range cases {
		if got := candidateRefs(c.ref, "", sha); !reflect.DeepEqual(got, c.want) {
			t.Errorf("candidateRefs(%q) = %v want %v", c.ref, got, c.want)
		}
	}

	got := candidateRefs("refs/pull/9/merge", "pr-branch", sha)
	want := []string{"refs/pull/9/merge", "refs/pull/9/head", "refs/heads/pr-branch", sha}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("with HEAD_REF: got %v want %v", got, want)
	}

	if got := candidateRefs("", "", ""); got != nil {
		t.Errorf("empty ref and sha should give no candidates, got %v", got)
	}
}

func TestIsCommitSHA(t *testing.T) {
	if !isCommitSHA(strings.Repeat("0", 40)) {
		t.Error("40-hex should be a SHA")
	}
	if !isCommitSHA(strings.Repeat("0", 64)) {
		t.Error("64-hex (sha256 repos) should be a SHA")
	}
	for _, s := range []string{"", "abc", strings.Repeat("g", 40), strings.Repeat("0", 39)} {
		if isCommitSHA(s) {
			t.Errorf("%q should not be a SHA", s)
		}
	}
}

func TestSanitizeRef(t *testing.T) {
	if got := sanitizeRef("refs/pull/12/merge"); got != "refs_pull_12_merge" {
		t.Errorf("got %q", got)
	}
}
