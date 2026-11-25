// Package packagelatest looks up packages in a Forgejo/Gitea instance's
// package registry and reports the latest version: the highest version
// among the registry entries matching an optional package type, name
// and version prefix. It also reports the next version to release —
// the latest with its patch component incremented, or the prefix with
// a 0 appended when nothing matched — which scheduled image builds use
// to pick the day's build number. Listing public packages needs no
// credentials; the token input authenticates private ones.
package packagelatest

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.mouse-lake.ts.net/actions/toolkit/internal/gha"
	"git.mouse-lake.ts.net/actions/toolkit/internal/ver"
)

// pageSize is how many packages to request per page; the registry API
// caps it at 50 (a var so tests can lower it and exercise pagination).
// maxPages bounds the walk so a huge registry cannot spin the action
// forever (20 pages × 50 = 1000 versions).
var pageSize = 50

const maxPages = 20

// Inputs carries the action inputs; the token arrives through the
// environment (TOKEN) so it never shows in argv.
type Inputs struct {
	ServerURL string
	Owner     string
	Type      string
	Package   string
	Prefix    string
	Token     string
}

// regPackage is one entry of the registry's package listing: for
// container packages the version is the image tag.
type regPackage struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	CreatedAt string `json:"created_at"`
}

// Run lists the owner's packages and reports the latest matching
// version as the version output, and the next version to release as
// the next output.
func Run(c *gha.Context, log *slog.Logger, in Inputs) error {
	server := strings.TrimSpace(in.ServerURL)
	if server == "" {
		server = c.ServerURL
	}
	if server == "" {
		return fmt.Errorf("no server URL: set the server-url input or run inside a workflow")
	}
	api := strings.TrimSuffix(server, "/") + "/api/v1"

	owner := strings.TrimSpace(in.Owner)
	if owner == "" {
		owner = c.Owner()
	}
	if owner == "" {
		return fmt.Errorf("no package owner: set the owner input or run inside a workflow")
	}

	if in.Token != "" {
		gha.Mask(in.Token)
	}

	pkgs, err := listPackages(api, owner, in.Type, in.Package, in.Token)
	if err != nil {
		return err
	}

	latest, found, skipped := pickLatest(pkgs, in.Prefix)
	if skipped > 0 {
		log.Debug(fmt.Sprintf("skipped %d version(s) that are not semver", skipped))
	}

	var version string
	if found {
		version = latest.Version
	}
	next := nextVersion(in.Prefix, latest.Version, found)

	if err := c.Output("version", version); err != nil {
		return err
	}
	if err := c.Output("next", next); err != nil {
		return err
	}
	if found {
		log.Info(fmt.Sprintf("latest version of %s/%s is %s; next %s", owner, in.Package, version, next))
	} else {
		log.Info(fmt.Sprintf("no versions of %s/%s in the registry%s; next %s", owner, in.Package, prefixNote(in.Prefix), next))
	}
	return nil
}

func prefixNote(prefix string) string {
	if prefix == "" {
		return ""
	}
	return fmt.Sprintf(" matching prefix %q", prefix)
}

// pickLatest returns the highest version among the packages matching
// the prefix, and how many entries were skipped because their version
// is not semver. Registry entries with unparsable versions (the
// evergreen latest tag) can never be the latest and are ignored.
func pickLatest(pkgs []regPackage, prefix string) (latest regPackage, found bool, skipped int) {
	var latestVer ver.Version
	for _, p := range pkgs {
		if prefix != "" && !strings.HasPrefix(p.Version, prefix) {
			continue
		}
		v, err := ver.Parse(p.Version)
		if err != nil {
			skipped++
			continue
		}
		if !found || ver.Compare(v, latestVer) > 0 {
			latest, latestVer, found = p, v, true
		}
	}
	return latest, found, skipped
}

// nextVersion renders the version a new release should use: the latest
// with its patch component incremented; the prefix with a 0 appended
// when the registry had no matching version (write the prefix with its
// trailing separator, e.g. 1.20261007.); 0.0.0 without a prefix.
func nextVersion(prefix, version string, found bool) string {
	if found {
		if v, err := ver.Parse(version); err == nil {
			v.Patch++
			return v.String()
		}
	}
	if prefix != "" {
		return prefix + "0"
	}
	return ver.Version{}.String()
}

// listPackages walks the owner's package listing page by page. The q
// parameter narrows by package name (substring match), type by package
// type (container, generic, ...).
func listPackages(api, owner, typ, name, token string) ([]regPackage, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	var all []regPackage
	for page := 1; page <= maxPages; page++ {
		q := url.Values{
			"limit": {strconv.Itoa(pageSize)},
			"page":  {strconv.Itoa(page)},
		}
		if typ != "" {
			q.Set("type", typ)
		}
		if name != "" {
			q.Set("q", name)
		}
		req, err := http.NewRequest(http.MethodGet, api+"/packages/"+url.PathEscape(owner)+"?"+q.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("listing packages of %s: %w", owner, err)
		}
		pkgs, err := decodePackages(resp)
		if err != nil {
			return nil, err
		}
		all = append(all, pkgs...)
		if len(pkgs) < pageSize {
			return all, nil
		}
	}
	return all, nil
}

// decodePackages parses the listing body and turns non-2xx responses
// into errors naming the status and whatever body the API returned.
func decodePackages(resp *http.Response) ([]regPackage, error) {
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var buf [512]byte
		n, _ := resp.Body.Read(buf[:])
		return nil, fmt.Errorf("listing packages: %s: %s", resp.Status, strings.TrimSpace(string(buf[:n])))
	}
	var pkgs []regPackage
	if err := json.NewDecoder(resp.Body).Decode(&pkgs); err != nil {
		return nil, fmt.Errorf("parsing package listing: %w", err)
	}
	return pkgs, nil
}
