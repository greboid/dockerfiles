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
	"io/fs"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	sharedVar = "state/os/default/var"
	deployDir = "state/deploy"
	stateDir  = "lib/osimage/etc"
	// The initramfs binds this over /etc/systemd/network when it exists.
	networkDir = "lib/osimage/etc/network"
	ipv6DropIn = "80-dhcp.network.d/50-static-ipv6.conf"
	ipv4DropIn = "80-dhcp.network.d/50-static-ipv4.conf"
	dnsDropIn  = "80-dhcp.network.d/50-dns.conf"
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
	conf := fmt.Sprintf("[Network]\nAddress=%s\nGateway=%s\n", s.Address, s.Gateway)
	return writeNetworkFile(varDir, ipv6DropIn, conf)
}

// StaticIPv4 is a fixed IPv4 address replacing DHCPv4 on the
// DHCP-configured interface.
type StaticIPv4 struct {
	Address netip.Prefix
	Gateway netip.Addr
}

// ParseStaticIPv4 parses an address with prefix length (192.0.2.10/24)
// and a gateway address.
func ParseStaticIPv4(address, gateway string) (*StaticIPv4, error) {
	a, err := ParseIPv4Address(address)
	if err != nil {
		return nil, err
	}
	g, err := netip.ParseAddr(gateway)
	if err != nil || !g.Is4() {
		return nil, fmt.Errorf("ipv4 gateway %q is not an IPv4 address", gateway)
	}
	return &StaticIPv4{Address: a, Gateway: g}, nil
}

// ParseIPv4Address parses the address/prefix-length half of a StaticIPv4.
func ParseIPv4Address(address string) (netip.Prefix, error) {
	a, err := netip.ParsePrefix(address)
	if err != nil {
		return a, fmt.Errorf("ipv4 address %q: want address/prefix-length", address)
	}
	if !a.Addr().Is4() || a.Addr().IsUnspecified() || a.Addr().IsLoopback() || a.Addr().IsMulticast() {
		return a, fmt.Errorf("ipv4 address %q is not a usable IPv4 address", address)
	}
	return a, nil
}

func (s StaticIPv4) String() string {
	return fmt.Sprintf("%s via %s", s.Address, s.Gateway)
}

// WriteStaticIPv4 seeds a drop-in for osimage-config's 80-dhcp.network
// that turns DHCPv4 off (DHCPv6 stays) and sets the address.
func WriteStaticIPv4(varDir string, s StaticIPv4) error {
	conf := fmt.Sprintf("[Network]\nDHCP=ipv6\nAddress=%s\nGateway=%s\n", s.Address, s.Gateway)
	return writeNetworkFile(varDir, ipv4DropIn, conf)
}

// DNSSECModes are the values systemd-networkd accepts for DNSSEC=.
var DNSSECModes = []string{"yes", "no", "allow-downgrade"}

// DNS is per-link resolver configuration for 80-dhcp.network. An
// empty DNSSEC keeps systemd-resolved's default.
type DNS struct {
	Servers []netip.Addr
	DNSSEC  string
}

// ParseDNS parses comma- or space-separated server addresses and a
// DNSSEC mode (empty for the default).
func ParseDNS(servers, dnssec string) (*DNS, error) {
	d := &DNS{}
	for f := range strings.FieldsFuncSeq(servers, func(r rune) bool { return r == ',' || r == ' ' }) {
		a, err := netip.ParseAddr(f)
		if err != nil || a.IsUnspecified() {
			return nil, fmt.Errorf("dns server %q is not an IP address", f)
		}
		d.Servers = append(d.Servers, a)
	}
	if dnssec != "" && !slices.Contains(DNSSECModes, dnssec) {
		return nil, fmt.Errorf("dnssec %q: want one of %s", dnssec, strings.Join(DNSSECModes, ", "))
	}
	d.DNSSEC = dnssec
	if len(d.Servers) == 0 && d.DNSSEC == "" {
		return nil, nil
	}
	return d, nil
}

// ServerList is the servers comma-separated.
func (d DNS) ServerList() string {
	s := make([]string, len(d.Servers))
	for i, a := range d.Servers {
		s[i] = a.String()
	}
	return strings.Join(s, ",")
}

func (d DNS) String() string {
	var parts []string
	if len(d.Servers) > 0 {
		parts = append(parts, "servers "+d.ServerList())
	}
	if d.DNSSEC != "" {
		parts = append(parts, "dnssec "+d.DNSSEC)
	}
	return strings.Join(parts, ", ")
}

// WriteDNS seeds a drop-in for osimage-config's 80-dhcp.network.
func WriteDNS(varDir string, d DNS) error {
	var b strings.Builder
	b.WriteString("[Network]\n")
	for _, a := range d.Servers {
		fmt.Fprintf(&b, "DNS=%s\n", a)
	}
	if d.DNSSEC != "" {
		fmt.Fprintf(&b, "DNSSEC=%s\n", d.DNSSEC)
	}
	return writeNetworkFile(varDir, dnsDropIn, b.String())
}

// NetworkFiles lists the files and symlinks under src (a
// systemd-networkd config directory), relative to it. A missing src
// has none.
func NetworkFiles(src string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == src && os.IsNotExist(err) {
				return fs.SkipAll
			}
			return err
		}
		if d.Type().IsRegular() || d.Type()&fs.ModeSymlink != 0 {
			rel, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			out = append(out, rel)
		}
		return nil
	})
	return out, err
}

// CopyNetwork copies the files and symlinks (masks to /dev/null, say)
// under src to /var/lib/osimage/etc/network, which the initramfs binds
// over /etc/systemd/network.
func CopyNetwork(src, varDir string) error {
	files, err := NetworkFiles(src)
	if err != nil {
		return err
	}
	for _, rel := range files {
		from := filepath.Join(src, rel)
		to := filepath.Join(varDir, networkDir, rel)
		if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
			return err
		}
		fi, err := os.Lstat(from)
		if err != nil {
			return err
		}
		if fi.Mode()&fs.ModeSymlink != 0 {
			link, err := os.Readlink(from)
			if err != nil {
				return err
			}
			os.Remove(to)
			if err := os.Symlink(link, to); err != nil {
				return err
			}
			continue
		}
		b, err := os.ReadFile(from)
		if err != nil {
			return err
		}
		if err := writeFile(to, b, fi.Mode().Perm(), -1, -1); err != nil {
			return err
		}
	}
	return nil
}

func writeNetworkFile(varDir, rel, conf string) error {
	path := filepath.Join(varDir, networkDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
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
