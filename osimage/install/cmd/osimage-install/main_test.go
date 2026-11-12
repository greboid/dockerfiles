package main

import (
	"context"
	"strings"
	"testing"

	"github.com/greboid/osimage-install/internal/installer"
	"github.com/greboid/osimage-install/internal/sha512crypt"
)

func TestWipeQuestion(t *testing.T) {
	for _, tt := range []struct {
		content string
		want    string
	}{
		{"", "Erase /dev/vda (QEMU QEMU_HARDDISK) and install?"},
		{"2 partitions (vfat, ext4)", "Erase /dev/vda (QEMU QEMU_HARDDISK, currently 2 partitions (vfat, ext4)) and install?"},
	} {
		tgt := &installer.Target{Path: "/dev/vda", Ident: "QEMU QEMU_HARDDISK", Content: tt.content}
		if got := wipeQuestion("/dev/vda", tgt); got != tt.want {
			t.Errorf("wipeQuestion(content %q) = %q, want %q", tt.content, got, tt.want)
		}
	}
}

func TestAskDefaults(t *testing.T) {
	t.Setenv("PATH", "") // no tailscale CLI: the Tailscale question is skipped
	cfg := &installer.Config{
		Hostname:  "box1",
		Passwords: []installer.Password{{User: "root", Hash: "$6$s$fromflag"}},
	}
	// Enter on everything keeps the command line values.
	p, out := fakePrompter("\n", "")
	if err := ask(context.Background(), p, cfg, &flags{hostname: "box1"}); err != nil {
		t.Fatal(err)
	}
	if cfg.Hostname != "box1" || len(cfg.Passwords) != 1 || cfg.Passwords[0].Hash != "$6$s$fromflag" {
		t.Errorf("defaults not kept: %+v", cfg)
	}
	for _, want := range []string{"Hostname [box1]: ", "[Enter keeps the password from the command line]"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q: %q", want, out.String())
		}
	}
}

func TestAskAnswers(t *testing.T) {
	t.Setenv("PATH", "")
	cfg := &installer.Config{}
	p, out := fakePrompter("box2\n", "pw", "pw")
	if err := ask(context.Background(), p, cfg, &flags{}); err != nil {
		t.Fatal(err)
	}
	if cfg.Hostname != "box2" {
		t.Errorf("hostname = %q", cfg.Hostname)
	}
	if !strings.Contains(out.String(), "[Enter keeps no password (image default)]") {
		t.Errorf("prompt = %q", out.String())
	}
	if len(cfg.Passwords) != 1 || cfg.Passwords[0].User != "root" {
		t.Fatalf("passwords = %+v", cfg.Passwords)
	}
	salt := strings.Split(cfg.Passwords[0].Hash, "$")[2]
	if sha512crypt.Crypt([]byte("pw"), []byte(salt), sha512crypt.DefaultRounds) != cfg.Passwords[0].Hash {
		t.Error("root hash is not the typed password")
	}
}

func TestAskImageDefaultHostname(t *testing.T) {
	t.Setenv("PATH", "")
	cfg := &installer.Config{}
	p, _ := fakePrompter("\n", "")
	if err := ask(context.Background(), p, cfg, &flags{}); err != nil {
		t.Fatal(err)
	}
	if cfg.Hostname != "" || len(cfg.Passwords) != 0 {
		t.Errorf("Enter everywhere should seed nothing: %+v", cfg)
	}
}
