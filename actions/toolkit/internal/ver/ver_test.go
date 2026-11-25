package ver

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in           string
		major, minor int
		patch        int
		pre, build   string
	}{
		{"1", 1, 0, 0, "", ""},
		{"1.2", 1, 2, 0, "", ""},
		{"1.2.3", 1, 2, 3, "", ""},
		{"v1.2.3", 1, 2, 3, "", ""},
		{"go1.2.3", 1, 2, 3, "", ""},
		{"v1.2.3-rc.1", 1, 2, 3, "rc.1", ""},
		{"1.2.3-rc.1+meta", 1, 2, 3, "rc.1", "meta"},
	}
	for _, c := range cases {
		v, err := Parse(c.in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", c.in, err)
		}
		if v.Major != c.major || v.Minor != c.minor || v.Patch != c.patch || v.Pre != c.pre || v.Build != c.build {
			t.Errorf("Parse(%q) = %+v", c.in, v)
		}
	}
	if _, err := Parse(""); err == nil {
		t.Error("empty version should fail")
	}
	if _, err := Parse("1.2.3.4"); err == nil {
		t.Error("four components should fail")
	}
}

func TestCompare(t *testing.T) {
	ordered := []string{
		"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-beta", "1.0.0", "1.0.1", "1.1.0", "2.0.0",
	}
	for i := 1; i < len(ordered); i++ {
		a, _ := Parse(ordered[i-1])
		b, _ := Parse(ordered[i])
		if Compare(a, b) >= 0 {
			t.Errorf("%s should sort before %s", ordered[i-1], ordered[i])
		}
		if Compare(b, a) <= 0 {
			t.Errorf("%s should sort after %s", ordered[i], ordered[i-1])
		}
	}
	// build metadata is ignored
	a, _ := Parse("1.2.3+abc")
	b, _ := Parse("1.2.3+def")
	if Compare(a, b) != 0 {
		t.Error("build metadata should not affect ordering")
	}
	// numeric identifiers sort below alphanumeric
	x, _ := Parse("1.0.0-1")
	y, _ := Parse("1.0.0-alpha")
	if Compare(x, y) >= 0 {
		t.Error("numeric pre-release should sort below alphanumeric")
	}
}

func TestString(t *testing.T) {
	v, _ := Parse("v1.2.3-rc.1+meta")
	if got := v.String(); got != "1.2.3-rc.1" {
		t.Errorf("got %q", got)
	}
}
