package shadow

import (
	"errors"
	"testing"
)

const sample = "root:::0:::::\nmessagebus:!::0:::::\nsystemd-network:!::0:::::\n"

func TestSetHash(t *testing.T) {
	got, err := SetHash([]byte(sample), "root", "$6$salt$hash", 20400)
	if err != nil {
		t.Fatal(err)
	}
	want := "root:$6$salt$hash:20400:0:::::\nmessagebus:!::0:::::\nsystemd-network:!::0:::::\n"
	if string(got) != want {
		t.Errorf("SetHash() =\n%q\nwant\n%q", got, want)
	}
}

func TestSetHashNoTrailingNewline(t *testing.T) {
	got, err := SetHash([]byte("a:x:1:0:::::\nb:x:1:0:::::"), "b", "!", 5)
	if err != nil {
		t.Fatal(err)
	}
	if want := "a:x:1:0:::::\nb:!:5:0:::::"; string(got) != want {
		t.Errorf("SetHash() = %q, want %q", got, want)
	}
}

func TestSetHashErrors(t *testing.T) {
	if _, err := SetHash([]byte(sample), "nobody", "$6$s$h", 1); !errors.Is(err, ErrNoUser) {
		t.Errorf("missing user: err = %v, want ErrNoUser", err)
	}
	if _, err := SetHash([]byte("root:x\n"), "root", "$6$s$h", 1); err == nil {
		t.Error("malformed line: want error")
	}
	if _, err := SetHash([]byte(sample), "root", "plaintext", 1); err == nil {
		t.Error("plaintext hash: want error")
	}
}

func TestValidateHash(t *testing.T) {
	for _, h := range []string{"$6$salt$hash", "$6$rounds=9000$salt$hash", "$y$j9T$salt$hash", "!", "*"} {
		if err := ValidateHash(h); err != nil {
			t.Errorf("ValidateHash(%q) = %v, want nil", h, err)
		}
	}
	for _, h := range []string{"", "hunter2", "$6$", "$6$salt$", "$$salt$hash", "$6$a:b$c", "$6$s$h\n"} {
		if err := ValidateHash(h); err == nil {
			t.Errorf("ValidateHash(%q) = nil, want error", h)
		}
	}
}

func TestValidateUser(t *testing.T) {
	if err := ValidateUser("root"); err != nil {
		t.Error(err)
	}
	for _, u := range []string{"", "a:b", "a\nb"} {
		if err := ValidateUser(u); err == nil {
			t.Errorf("ValidateUser(%q) = nil, want error", u)
		}
	}
}
