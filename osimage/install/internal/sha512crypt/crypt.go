// Package sha512crypt implements the SHA-512 crypt(3) scheme ("$6$")
// as specified by Ulrich Drepper's "Unix crypt using SHA-256 and
// SHA-512", the format glibc, musl and busybox cryptpw produce for
// /etc/shadow.
package sha512crypt

import (
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"strconv"
)

const (
	// DefaultRounds is the round count used when none is given; it is
	// left out of the encoded hash.
	DefaultRounds = 5000
	// MinRounds and MaxRounds bound the round count, as in the spec.
	MinRounds = 1000
	MaxRounds = 999999999

	saltMax  = 16
	alphabet = "./0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// Generate hashes password with a random 16 character salt and the
// default round count.
func Generate(password []byte) (string, error) {
	raw := make([]byte, saltMax)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}
	salt := make([]byte, saltMax)
	for i, b := range raw {
		salt[i] = alphabet[b&0x3f]
	}
	return Crypt(password, salt, DefaultRounds), nil
}

// Crypt returns the encoded "$6$" hash of password. The salt is
// truncated to 16 bytes and rounds is clamped to [MinRounds,
// MaxRounds], as the spec requires.
func Crypt(password, salt []byte, rounds int) string {
	if len(salt) > saltMax {
		salt = salt[:saltMax]
	}
	rounds = max(MinRounds, min(MaxRounds, rounds))

	h := sha512.New()

	// Alternate sum B = H(P S P).
	h.Write(password)
	h.Write(salt)
	h.Write(password)
	alt := h.Sum(nil)

	// A = H(P S B[len(P)] bits(len(P))).
	h.Reset()
	h.Write(password)
	h.Write(salt)
	for n := len(password); n > 0; n -= sha512.Size {
		h.Write(alt[:min(n, sha512.Size)])
	}
	for n := len(password); n > 0; n >>= 1 {
		if n&1 != 0 {
			h.Write(alt)
		} else {
			h.Write(password)
		}
	}
	a := h.Sum(nil)

	// P' = H(P * len(P)) stretched to len(P).
	h.Reset()
	for range len(password) {
		h.Write(password)
	}
	pSeq := stretch(h.Sum(nil), len(password))

	// S' = H(S * (16 + A[0])) stretched to len(S).
	h.Reset()
	for range 16 + int(a[0]) {
		h.Write(salt)
	}
	sSeq := stretch(h.Sum(nil), len(salt))

	c := a
	for i := range rounds {
		h.Reset()
		if i&1 != 0 {
			h.Write(pSeq)
		} else {
			h.Write(c)
		}
		if i%3 != 0 {
			h.Write(sSeq)
		}
		if i%7 != 0 {
			h.Write(pSeq)
		}
		if i&1 != 0 {
			h.Write(c)
		} else {
			h.Write(pSeq)
		}
		c = h.Sum(c[:0])
	}

	out := []byte("$6$")
	if rounds != DefaultRounds {
		out = append(out, "rounds="...)
		out = strconv.AppendInt(out, int64(rounds), 10)
		out = append(out, '$')
	}
	out = append(out, salt...)
	out = append(out, '$')
	return string(encode(out, c))
}

func stretch(digest []byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = digest[i%len(digest)]
	}
	return out
}

// byteOrder is the spec's permutation of the final digest into
// 24-bit groups.
var byteOrder = [21][3]int{
	{0, 21, 42}, {22, 43, 1}, {44, 2, 23}, {3, 24, 45}, {25, 46, 4},
	{47, 5, 26}, {6, 27, 48}, {28, 49, 7}, {50, 8, 29}, {9, 30, 51},
	{31, 52, 10}, {53, 11, 32}, {12, 33, 54}, {34, 55, 13}, {56, 14, 35},
	{15, 36, 57}, {37, 58, 16}, {59, 17, 38}, {18, 39, 60}, {40, 61, 19},
	{62, 20, 41},
}

func encode(out, c []byte) []byte {
	put := func(w uint32, n int) {
		for range n {
			out = append(out, alphabet[w&0x3f])
			w >>= 6
		}
	}
	for _, g := range byteOrder {
		put(uint32(c[g[0]])<<16|uint32(c[g[1]])<<8|uint32(c[g[2]]), 4)
	}
	put(uint32(c[63]), 2)
	return out
}
