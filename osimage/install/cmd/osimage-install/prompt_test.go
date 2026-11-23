package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/greboid/osimage-install/internal/installer"
	"github.com/greboid/osimage-install/internal/target"
)

func fakePrompter(lines string) (*prompter, *bytes.Buffer) {
	var out bytes.Buffer
	return &prompter{in: bufio.NewReader(strings.NewReader(lines)), out: &out}, &out
}

func TestHostname(t *testing.T) {
	p, out := fakePrompter("\n")
	if got, err := p.hostname("box1"); err != nil || got != "box1" {
		t.Errorf("empty answer = %q, %v; want the default", got, err)
	}
	if !strings.Contains(out.String(), "Hostname [box1]: ") {
		t.Errorf("prompt = %q", out.String())
	}

	p, out = fakePrompter("bad_name\nbox2\n")
	if got, err := p.hostname("osimage"); err != nil || got != "box2" {
		t.Errorf("hostname() = %q, %v; want re-prompt then box2", got, err)
	}
	if !strings.Contains(out.String(), "contains '_'") {
		t.Errorf("no validation message in %q", out.String())
	}

	p, _ = fakePrompter("")
	if _, err := p.hostname("x"); err == nil {
		t.Error("EOF: want error")
	}
}

func TestYesNo(t *testing.T) {
	for _, tt := range []struct {
		in   string
		def  bool
		want bool
	}{
		{"\n", true, true}, {"\n", false, false}, {"y\n", false, true},
		{"No\n", true, false}, {"maybe\nyes\n", false, true},
	} {
		p, _ := fakePrompter(tt.in)
		if got, err := p.yesNo("Go?", tt.def); err != nil || got != tt.want {
			t.Errorf("yesNo(%q, %v) = %v, %v; want %v", tt.in, tt.def, got, err, tt.want)
		}
	}
	p, out := fakePrompter("\n")
	p.yesNo("Go?", true)
	if out.String() != "Go? [Y/n] " {
		t.Errorf("prompt = %q", out.String())
	}
}

func TestPickDisk(t *testing.T) {
	cands := []installer.Candidate{
		{Path: "/dev/sda", Size: 120 << 30, Ident: "Samsung SSD", Content: "2 partitions (vfat, ext4)"},
		{Path: "/dev/vda", Size: 20 << 30, Ident: "QEMU HARDDISK"},
		{Path: "/dev/zram0", Size: 16 << 30, Reason: "/dev/zram0 is RAM-backed, not a fixed disk"},
	}

	// Enter picks the only usable disk when there is just one.
	cfg := &installer.Config{}
	p, out := fakePrompter("\n")
	if err := pickDisk(p, cfg, []installer.Candidate{cands[1]}); err != nil || cfg.Disk != "/dev/vda" {
		t.Errorf("single usable: pickDisk() = %v, disk %q", err, cfg.Disk)
	}
	if !strings.Contains(out.String(), " 1) /dev/vda") || !strings.Contains(out.String(), "Install to which disk? [1]: ") {
		t.Errorf("single usable output: %q", out.String())
	}

	// The list numbers the usable disks and explains the refused ones.
	cfg = &installer.Config{}
	p, out = fakePrompter("2\n")
	if err := pickDisk(p, cfg, cands); err != nil || cfg.Disk != "/dev/vda" {
		t.Errorf("pickDisk() = %v, disk %q", err, cfg.Disk)
	}
	for _, want := range []string{
		" 1) /dev/sda", "120 GiB", "Samsung SSD, 2 partitions (vfat, ext4)",
		" 2) /dev/vda", "20 GiB", "QEMU HARDDISK",
		"not usable: /dev/zram0 is RAM-backed",
		"Install to which disk? [1-2]: ",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q: %q", want, out.String())
		}
	}

	// Empty answers and junk re-ask; a /dev path is taken as-is.
	cfg = &installer.Config{}
	p, out = fakePrompter("\nx\n/dev/nvme7n1\n")
	if err := pickDisk(p, cfg, cands); err != nil || cfg.Disk != "/dev/nvme7n1" {
		t.Errorf("pickDisk() = %v, disk %q", err, cfg.Disk)
	}
	if !strings.Contains(out.String(), "answer 1-2 or a /dev path") {
		t.Errorf("no re-ask message: %q", out.String())
	}

	// Nothing usable is an error.
	cfg = &installer.Config{}
	p, _ = fakePrompter("\n")
	if err := pickDisk(p, cfg, cands[2:2]); err == nil || cfg.Disk != "" {
		t.Errorf("nothing usable: pickDisk() = %v, disk %q; want error", err, cfg.Disk)
	}

	// EOF while asking is an error.
	p, _ = fakePrompter("")
	if err := pickDisk(p, &installer.Config{}, cands); err == nil {
		t.Error("EOF: want error")
	}
}

func TestFormatSize(t *testing.T) {
	for b, want := range map[uint64]string{
		20 << 30:              "20 GiB",
		5 << 30:               "5 GiB",
		uint64(1649267441664): "1.5 TiB",
	} {
		if got := formatSize(b); got != want {
			t.Errorf("formatSize(%d) = %q, want %q", b, got, want)
		}
	}
}

func TestIPv6Prompt(t *testing.T) {
	def, _ := target.ParseStaticIPv6("2001:db8::1/64", "fe80::1")
	p, out := fakePrompter("bad\n2001:db8::2/64\n")
	got, err := p.ipv6(nil, "fe80::1")
	if err != nil || got == nil || got.String() != "2001:db8::2/64 via fe80::1" {
		t.Fatalf("ipv6() = %v, %v", got, err)
	}
	if !strings.Contains(out.String(), "[none]: ") {
		t.Errorf("output missing default: %q", out.String())
	}
	p, _ = fakePrompter("\n")
	if got, _ := p.ipv6(def, "fe80::1"); got != def {
		t.Errorf("Enter: got %v, want default", got)
	}
	p, _ = fakePrompter("none\n")
	if got, _ := p.ipv6(def, "fe80::1"); got != nil {
		t.Errorf("none: got %v", got)
	}
}

func TestIPv4Prompt(t *testing.T) {
	p, out := fakePrompter("bad\n192.0.2.10/24\n\nfe80::1\n192.0.2.10/24\n192.0.2.1\n")
	got, err := p.ipv4(nil)
	if err != nil || got == nil || got.String() != "192.0.2.10/24 via 192.0.2.1" {
		t.Fatalf("ipv4() = %v, %v", got, err)
	}
	if !strings.Contains(out.String(), "[none]: ") || !strings.Contains(out.String(), `ipv4 address "bad"`) ||
		!strings.Contains(out.String(), `gateway "fe80::1" is not`) {
		t.Errorf("output: %q", out.String())
	}
	def := got
	p, _ = fakePrompter("\n")
	if got, _ := p.ipv4(def); got != def {
		t.Errorf("Enter: got %v, want default", got)
	}
	p, _ = fakePrompter("192.0.2.20/24\n\n")
	if got, _ := p.ipv4(def); got == nil || got.String() != "192.0.2.20/24 via 192.0.2.1" {
		t.Errorf("new address, default gateway: got %v", got)
	}
	p, _ = fakePrompter("none\n")
	if got, _ := p.ipv4(def); got != nil {
		t.Errorf("none: got %v", got)
	}
}

func TestDNSPrompt(t *testing.T) {
	p, out := fakePrompter("1.1.1.1 x\n1.1.1.1, 2606:4700::1111\nmaybe\nNo\n")
	got, err := p.dns(nil)
	if err != nil || got == nil || got.String() != "servers 1.1.1.1,2606:4700::1111, dnssec no" {
		t.Fatalf("dns() = %v, %v", got, err)
	}
	if !strings.Contains(out.String(), "[none]: ") || !strings.Contains(out.String(), "[default]: ") {
		t.Errorf("output: %q", out.String())
	}
	def := got
	p, _ = fakePrompter("\n\n")
	if got, _ := p.dns(def); got.String() != def.String() {
		t.Errorf("Enter: got %v, want default", got)
	}
	p, _ = fakePrompter("none\ndefault\n")
	if got, _ := p.dns(def); got != nil {
		t.Errorf("none/default: got %v", got)
	}
}
