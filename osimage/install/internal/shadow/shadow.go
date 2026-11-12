// Package shadow edits password hashes in /etc/shadow content.
package shadow

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrNoUser is returned when the user has no line in the shadow file.
var ErrNoUser = errors.New("no such user")

// ValidateUser rejects names that cannot appear as a shadow login.
func ValidateUser(user string) error {
	if user == "" {
		return errors.New("empty user name")
	}
	if strings.ContainsAny(user, ":\n\r") {
		return fmt.Errorf("invalid user name %q", user)
	}
	return nil
}

// ValidateHash accepts crypt(3) hashes ("$id$...") and the lock
// markers "!" and "*". Plain text and empty values are refused: an
// empty field means "no password", which is the image default anyway.
func ValidateHash(hash string) error {
	switch {
	case hash == "":
		return errors.New("empty password hash")
	case strings.ContainsAny(hash, ":\n\r"):
		return errors.New("password hash contains ':' or a newline")
	case hash == "!" || hash == "*":
		return nil
	case !strings.HasPrefix(hash, "$"):
		return errors.New("password hash is not a crypt(3) hash (expected $id$salt$hash)")
	}
	if parts := strings.Split(hash, "$"); len(parts) < 4 || parts[1] == "" || parts[len(parts)-1] == "" {
		return errors.New("password hash is not a crypt(3) hash (expected $id$salt$hash)")
	}
	return nil
}

// SetHash returns content with user's password field set to hash and
// the last-change field set to day (days since the epoch). Every other
// line, and the file's trailing newline, is kept byte for byte.
func SetHash(content []byte, user, hash string, day int64) ([]byte, error) {
	if err := ValidateUser(user); err != nil {
		return nil, err
	}
	if err := ValidateHash(hash); err != nil {
		return nil, err
	}
	lines := strings.SplitAfter(string(content), "\n")
	found := false
	for i, line := range lines {
		body, nl := strings.CutSuffix(line, "\n")
		fields := strings.Split(body, ":")
		if fields[0] != user {
			continue
		}
		if len(fields) < 3 {
			return nil, fmt.Errorf("malformed shadow line for %s", user)
		}
		fields[1] = hash
		fields[2] = strconv.FormatInt(day, 10)
		lines[i] = strings.Join(fields, ":")
		if nl {
			lines[i] += "\n"
		}
		found = true
	}
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrNoUser, user)
	}
	return []byte(strings.Join(lines, "")), nil
}
