// Package imagetags derives container image tags from a workflow ref —
// the rules of greboid/docker-tags-action: master/main builds are
// tagged "dev"; semver tags yield "latest" (when the tag's major
// matches the newest stable tag in the repository) plus M.m.p, M.m and
// M; every other ref yields no tags. Pre-release suffixes are dropped
// from the generated tags.
package imagetags

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Version is a parsed x[.y[.z]] version with an optional pre-release
// suffix. Build metadata is not represented; it never affects tagging.
type Version struct {
	Major, Minor, Patch int
	Pre                 string
}

// String renders the version with its pre-release suffix, if any.
func (v Version) String() string {
	s := v.Numeric()
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	return s
}

// Numeric renders only the numeric components — the form image tags
// are generated from, where pre-release suffixes are dropped.
func (v Version) Numeric() string {
	return strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor) + "." + strconv.Itoa(v.Patch)
}

// Less reports whether v sorts before o by version precedence
// (pre-releases sort before their release).
func (v Version) Less(o Version) bool {
	if v.Major != o.Major {
		return v.Major < o.Major
	}
	if v.Minor != o.Minor {
		return v.Minor < o.Minor
	}
	if v.Patch != o.Patch {
		return v.Patch < o.Patch
	}
	switch {
	case v.Pre == "" && o.Pre == "":
		return false
	case v.Pre == "":
		return false
	case o.Pre == "":
		return true
	}
	return v.Pre < o.Pre
}

// ParseVersion parses 1, 1.2, 1.2.3, v1.2.3, go1.2.3 and -pre /
// +build-suffixed forms; missing components default to zero. It
// reports false when s is not a version.
func ParseVersion(s string) (Version, bool) {
	v := Version{}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "go")
	s = strings.TrimPrefix(s, "v")
	s, _, _ = strings.Cut(s, "+") // build metadata is ignored
	if i := strings.IndexByte(s, '-'); i >= 0 {
		v.Pre = s[i+1:]
		s = s[:i]
	}
	if s == "" {
		return Version{}, false
	}
	for i, comp := range strings.SplitN(s, ".", 3) {
		if comp == "" || strings.IndexFunc(comp, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
			return Version{}, false
		}
		n, err := strconv.Atoi(comp) // Atoi semantics: 007 is 7
		if err != nil {
			return Version{}, false
		}
		switch i {
		case 0:
			v.Major = n
		case 1:
			v.Minor = n
		case 2:
			v.Patch = n
		}
	}
	return v, true
}

// DeriveTags returns the tags for the workflow ref: "dev" on
// master/main, the version list on semver tags (with "latest" first
// when latest — the repository's highest stable version, possibly
// empty — shares the tag's major), and nothing for any other ref.
func DeriveTags(ref, latest string) []string {
	switch ref {
	case "refs/heads/master", "refs/heads/main":
		return []string{"dev"}
	}
	tag, ok := strings.CutPrefix(ref, "refs/tags/")
	if !ok {
		return nil
	}
	v, ok := ParseVersion(tag)
	if !ok {
		return nil
	}
	tags := []string{v.Numeric()}
	if lv, ok := ParseVersion(latest); ok && lv.Major == v.Major {
		tags = append([]string{"latest"}, tags...)
	}
	return tags
}

// RefVersion returns the "version" of the build: the parsed M.m.p for
// tag refs, the SHA on master/main, "unknown" otherwise.
func RefVersion(ref, sha string) string {
	switch ref {
	case "refs/heads/master", "refs/heads/main":
		return sha
	}
	if tag, ok := strings.CutPrefix(ref, "refs/tags/"); ok {
		if v, ok := ParseVersion(tag); ok {
			return v.Numeric()
		}
	}
	return "unknown"
}

// LatestStableMajor returns the major of the highest stable version
// among the tag names, or "" when there is none.
func LatestStableMajor(names []string) string {
	var top Version
	found := false
	for _, name := range names {
		v, ok := ParseVersion(name)
		if !ok || v.Pre != "" {
			continue
		}
		if !found || top.Less(v) {
			top = v
			found = true
		}
	}
	if !found {
		return ""
	}
	return strconv.Itoa(top.Major)
}

// ListRepoTags lists the repository's tag names, paging through the
// instance API (GitHub and Forgejo/Gitea compatible). A failing page
// is not fatal — the "latest" rule only degrades to absent — so it
// logs a warning and returns whatever was collected before the
// failure.
func ListRepoTags(ctx context.Context, log *slog.Logger, apiURL, repository, token string) ([]string, error) {
	if apiURL == "" {
		return nil, fmt.Errorf("no API URL known; cannot list repository tags")
	}
	var names []string
	for page := 1; page <= 100; page++ {
		pageTags, err := listTagPage(ctx, apiURL, repository, token, page)
		if err != nil {
			if log != nil {
				log.Warn("could not list repository tags; \"latest\" cannot be determined", "error", err)
			}
			return names, nil
		}
		if log != nil && len(pageTags) > 0 {
			log.Info(fmt.Sprintf("listed %d repository tags (page %d)", len(pageTags), page))
		}
		names = append(names, pageTags...)
		if len(pageTags) < 100 {
			return names, nil
		}
	}
	return names, nil
}

func listTagPage(ctx context.Context, apiURL, repository, token string, page int) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		apiURL+"/repos/"+repository+"/tags?page="+strconv.Itoa(page)+"&per_page=100", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("listing tags: %s", resp.Status)
	}
	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, fmt.Errorf("tag list was not valid JSON: %w", err)
	}
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names, nil
}
