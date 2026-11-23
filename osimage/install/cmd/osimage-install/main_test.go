package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/greboid/osimage-install/internal/installer"
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
	cfg := &installer.Config{Hostname: "box1"}
	// Enter on everything keeps the command line values.
	p, out := fakePrompter("\n\n\n\n\n")
	if err := ask(context.Background(), p, cfg, &flags{hostname: "box1"}); err != nil {
		t.Fatal(err)
	}
	if cfg.Hostname != "box1" {
		t.Errorf("defaults not kept: %+v", cfg)
	}
	if !strings.Contains(out.String(), "Hostname [box1]: ") {
		t.Errorf("output missing hostname prompt: %q", out.String())
	}
}

func TestAskAnswers(t *testing.T) {
	t.Setenv("PATH", "")
	cfg := &installer.Config{}
	p, out := fakePrompter("box2\n192.0.2.10/24\n192.0.2.1\n2001:db8::1/64\n192.0.2.53\nno\n")
	if err := ask(context.Background(), p, cfg, &flags{ipv6Gateway: defaultIPv6Gateway}); err != nil {
		t.Fatal(err)
	}
	if cfg.Hostname != "box2" {
		t.Errorf("hostname = %q", cfg.Hostname)
	}
	if cfg.IPv4 == nil || cfg.IPv4.String() != "192.0.2.10/24 via 192.0.2.1" {
		t.Errorf("ipv4 = %v", cfg.IPv4)
	}
	if cfg.IPv6 == nil || cfg.IPv6.String() != "2001:db8::1/64 via fe80::1" {
		t.Errorf("ipv6 = %v", cfg.IPv6)
	}
	if cfg.DNS == nil || cfg.DNS.String() != "servers 192.0.2.53, dnssec no" {
		t.Errorf("dns = %v", cfg.DNS)
	}
	if strings.Contains(out.String(), "assword") {
		t.Errorf("unexpected password prompt: %q", out.String())
	}
}

func TestAskImageDefaultHostname(t *testing.T) {
	t.Setenv("PATH", "")
	cfg := &installer.Config{}
	p, _ := fakePrompter("\n\n\n\n\n")
	if err := ask(context.Background(), p, cfg, &flags{}); err != nil {
		t.Fatal(err)
	}
	if cfg.Hostname != "" || cfg.IPv4 != nil || cfg.IPv6 != nil || cfg.DNS != nil {
		t.Errorf("Enter everywhere should seed nothing: %+v", cfg)
	}
}

func TestAskNetworkCopy(t *testing.T) {
	t.Setenv("PATH", "")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "10-eth0.network"), []byte("[Match]\nName=eth0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &installer.Config{CopyNetwork: true, Paths: installer.Paths{Network: dir}}
	p, out := fakePrompter("\nn\n\n\n\n\n")
	if err := ask(context.Background(), p, cfg, &flags{}); err != nil {
		t.Fatal(err)
	}
	if cfg.CopyNetwork {
		t.Error("answered n: CopyNetwork still set")
	}
	if !strings.Contains(out.String(), "(10-eth0.network)") || !strings.Contains(out.String(), "network configuration to the installed system? [Y/n]") {
		t.Errorf("output: %q", out.String())
	}
}
