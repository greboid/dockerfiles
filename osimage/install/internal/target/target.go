// Package target seeds per-machine state onto a freshly installed
// osimage root partition.
//
// The layout is bootc's composefs backend (bootc 1.17.3,
// composefs_consts.rs / bootc_composefs/state.rs):
//
//	state/os/default/var       the /var shared by every deployment
//	state/deploy/<id>/etc/     a pristine copy of the image's /etc
//	state/deploy/<id>/var  ->  ../../os/default/var
//
// The osimage initramfs (90osimage-state in osimage-kernel) seeds
// /var/lib/osimage/etc/{machine-id,hostname} only when they do
// not exist yet, then binds them read-only over /etc. Writing them here
// before the first boot therefore replaces the image defaults; the
// machine-id in particular is generated here, so the read-only bind
// never needs PID 1 to fill it in.
package target

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
)

const (
	sharedVar = "state/os/default/var"
	deployDir = "state/deploy"
	stateDir  = "lib/osimage/etc"
	// The initramfs binds this over /etc/systemd/network when it exists.
	networkDir = "lib/osimage/etc/network"
	ipv6DropIn = "80-dhcp.network.d/50-static-ipv6.conf"
)

// Root is a mounted, freshly installed root partition.
type Root struct {
	Path string
}

// Var returns the deployment-shared /var directory.
func (r Root) Var() (string, error) {
	p := filepath.Join(r.Path, sharedVar)
	if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("%s not found on the installed root; not a bootc composefs install?", sharedVar)
	}
	return p, nil
}

// ValidateHostname accepts an RFC 1123 host name of at most 64 bytes
// (HOST_NAME_MAX), optionally dotted.
func ValidateHostname(name string) error {
	if name == "" || len(name) > 64 {
		return fmt.Errorf("hostname %q must be 1-64 characters", name)
	}
	for label := range strings.SplitSeq(name, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("hostname %q has an invalid label %q", name, label)
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return fmt.Errorf("hostname %q contains %q (letters, digits and '-' only)", name, c)
			}
		}
	}
	return nil
}

func ensureStateDir(varDir string) (string, error) {
	dir := filepath.Join(varDir, stateDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	// Same mode the initramfs gives it.
	return dir, os.Chmod(dir, 0o755)
}

// WriteHostname seeds /var/lib/osimage/etc/hostname.
func WriteHostname(varDir, name string) error {
	if err := ValidateHostname(name); err != nil {
		return err
	}
	dir, err := ensureStateDir(varDir)
	if err != nil {
		return err
	}
	// Owner left as-is (-1): the installer runs as root.
	return writeFile(filepath.Join(dir, "hostname"), []byte(name+"\n"), 0o644, -1, -1)
}

// StaticIPv6 is a fixed IPv6 address on the DHCP-configured interface.
type StaticIPv6 struct {
	Address netip.Prefix
	Gateway netip.Addr
}

// ParseStaticIPv6 parses an address with prefix length (2001:db8::1/64)
// and a gateway address.
func ParseStaticIPv6(address, gateway string) (*StaticIPv6, error) {
	a, err := netip.ParsePrefix(address)
	if err != nil {
		return nil, fmt.Errorf("ipv6 address %q: want address/prefix-length", address)
	}
	if !a.Addr().Is6() || a.Addr().Is4In6() || a.Addr().IsLinkLocalUnicast() || a.Addr().IsUnspecified() {
		return nil, fmt.Errorf("ipv6 address %q is not a usable IPv6 address", address)
	}
	g, err := netip.ParseAddr(gateway)
	if err != nil || !g.Is6() || g.Is4In6() || g.Zone() != "" {
		return nil, fmt.Errorf("ipv6 gateway %q is not an IPv6 address", gateway)
	}
	return &StaticIPv6{Address: a, Gateway: g}, nil
}

func (s StaticIPv6) String() string {
	return fmt.Sprintf("%s via %s", s.Address, s.Gateway)
}

// WriteStaticIPv6 seeds a drop-in for osimage-config's 80-dhcp.network
// under /var/lib/osimage/etc/network.
func WriteStaticIPv6(varDir string, s StaticIPv6) error {
	path := filepath.Join(varDir, networkDir, ipv6DropIn)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	conf := fmt.Sprintf("[Network]\nAddress=%s\nGateway=%s\n", s.Address, s.Gateway)
	return writeFile(path, []byte(conf), 0o644, -1, -1)
}

// NewMachineID returns a random machine ID in systemd's format: 32
// lowercase hex digits.
func NewMachineID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating machine-id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// WriteMachineID seeds /var/lib/osimage/etc/machine-id (mode 0444,
// like systemd's own) with id.
func WriteMachineID(varDir, id string) error {
	if len(id) != 32 || strings.Trim(id, "0123456789abcdef") != "" {
		return fmt.Errorf("invalid machine-id %q", id)
	}
	dir, err := ensureStateDir(varDir)
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(dir, "machine-id"), []byte(id+"\n"), 0o444, -1, -1)
}

// writeFile writes atomically: nothing binds these paths yet, so a
// rename is safe here (unlike on the running system). uid/gid -1
// keep the creating process's identity.
func writeFile(path string, data []byte, mode os.FileMode, uid, gid int) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if (uid >= 0 && uid != os.Getuid()) || (gid >= 0 && gid != os.Getgid()) {
		if err := tmp.Chown(uid, gid); err != nil {
			tmp.Close()
			return err
		}
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
