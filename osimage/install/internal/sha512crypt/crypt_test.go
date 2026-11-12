package sha512crypt

import (
	"strings"
	"testing"
)

func TestCrypt(t *testing.T) {
	tests := []struct {
		name     string
		password string
		salt     string
		rounds   int
		want     string
	}{
		{
			name:     "spec vector, default rounds",
			password: "Hello world!",
			salt:     "saltstring",
			rounds:   DefaultRounds,
			want:     "$6$saltstring$svn8UoSVapNtMuq1ukKS4tPQd8iKwSMHWjl/O817G3uBnIFNjnQJuesI68u4OTLiBFdcbYEdFCoEOfaS35inz1",
		},
		{
			name:     "spec vector, custom rounds, salt truncated to 16",
			password: "Hello world!",
			salt:     "saltstringsaltstring",
			rounds:   10000,
			want:     "$6$rounds=10000$saltstringsaltst$OW1/O6BYHV6BcXZu8QVeXbDWra3Oeqh0sbHbbMCVNSnCM/UrjmM0Dp8vOuZeHBy/YTBmSK6H9qs/y3RnOaw5v.",
		},
		{
			name:     "password longer than one digest",
			password: strings.Repeat("p", 130),
			salt:     "abcdefgh",
			rounds:   1000,
			want:     "$6$rounds=1000$abcdefgh$2h.yiDdLHlXddXNN/NO39P.M5GrzQebj2cg1UusuLeZxxb5lXcvLOmvsamWFGranGpN901eaufJ5iI8Vi1eCf/",
		},
		{
			name:     "empty password",
			password: "",
			salt:     "a1b2c3d4",
			rounds:   DefaultRounds,
			want:     "$6$a1b2c3d4$aohzwcTpU.NLz1Hz9Ik5INHGaC0f6u5L.u1HsPLRuscT0GtcLKxAOGaWTU/ODP9LQA5bK/E7V94ZBwQJ1llFB/",
		},
		{
			name:     "rounds clamped to minimum",
			password: strings.Repeat("p", 130),
			salt:     "abcdefgh",
			rounds:   10,
			want:     "$6$rounds=1000$abcdefgh$2h.yiDdLHlXddXNN/NO39P.M5GrzQebj2cg1UusuLeZxxb5lXcvLOmvsamWFGranGpN901eaufJ5iI8Vi1eCf/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Crypt([]byte(tt.password), []byte(tt.salt), tt.rounds); got != tt.want {
				t.Errorf("Crypt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	a, err := Generate([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Errorf("two hashes of the same password share a salt: %q", a)
	}
	parts := strings.Split(a, "$")
	if len(parts) != 4 || parts[1] != "6" || len(parts[2]) != saltMax || len(parts[3]) != 86 {
		t.Fatalf("unexpected hash shape %q", a)
	}
	if got := Crypt([]byte("secret"), []byte(parts[2]), DefaultRounds); got != a {
		t.Errorf("re-hashing with the generated salt = %q, want %q", got, a)
	}
}
