// Package ver implements just enough semver parsing and comparison to
// resolve toolchain and image tag versions.
package ver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a parsed semver. Missing minor/patch components default
// to zero, so Parse("1.2") == Parse("1.2.0") for comparison purposes.
type Version struct {
	Major, Minor, Patch int
	Pre                 string // pre-release, e.g. "rc.1"; empty means stable
	Build               string // build metadata, ignored in comparisons
}

// Parse accepts forms like 1, 1.2, 1.2.3, v1.2.3, go1.2.3,
// 1.2.3-rc.1 and 1.2.3-rc.1+meta.
func Parse(s string) (Version, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "go") && len(s) > 2 && s[2] >= '0' && s[2] <= '9' {
		s = s[2:]
	}
	s = strings.TrimPrefix(s, "v")
	if s == "" {
		return Version{}, fmt.Errorf("empty version")
	}
	var v Version
	if i := strings.IndexByte(s, '+'); i >= 0 {
		v.Build, s = s[i+1:], s[:i]
	}
	if i := strings.IndexByte(s, '-'); i >= 0 {
		v.Pre, s = s[i+1:], s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return Version{}, fmt.Errorf("too many components in %q", s)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("invalid numeric component %q", p)
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
	return v, nil
}

// String renders the normalized form M.m.p[-pre] (build metadata is
// dropped; docker tags cannot contain '+').
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	return s
}

// IsStable reports whether the version has no pre-release part.
func (v Version) IsStable() bool { return v.Pre == "" }

// Compare orders versions per semver: numeric fields, then pre-release
// (a stable version sorts after any pre-release of the same triple;
// numeric identifiers sort numerically and before alphanumeric ones;
// a longer identifier list sorts after a prefix of it). Build metadata
// is ignored.
func Compare(a, b Version) int {
	if a.Major != b.Major {
		return cmp(a.Major, b.Major)
	}
	if a.Minor != b.Minor {
		return cmp(a.Minor, b.Minor)
	}
	if a.Patch != b.Patch {
		return cmp(a.Patch, b.Patch)
	}
	switch {
	case a.Pre == b.Pre:
		return 0
	case a.Pre == "":
		return 1
	case b.Pre == "":
		return -1
	}
	return comparePre(a.Pre, b.Pre)
}

func cmp(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func comparePre(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		x, y := as[i], bs[i]
		xn, xerr := strconv.Atoi(x)
		yn, yerr := strconv.Atoi(y)
		switch {
		case xerr == nil && yerr == nil:
			if xn != yn {
				return cmp(xn, yn)
			}
		case xerr == nil: // numeric identifiers sort lower than alphanumeric
			return -1
		case yerr == nil:
			return 1
		default:
			if x != y {
				return strings.Compare(x, y)
			}
		}
	}
	return cmp(len(as), len(bs))
}
