// Package blockdev inspects block devices through lsblk, sysfs and
// mountinfo.
package blockdev

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Device is one node of `lsblk -J` output.
type Device struct {
	Path        string    `json:"path"`
	Type        string    `json:"type"`
	Size        Size      `json:"size"`
	Model       string    `json:"model"`
	Serial      string    `json:"serial"`
	PartType    string    `json:"parttype"`
	FSType      string    `json:"fstype"`
	Mountpoints []*string `json:"mountpoints"`
	Children    []Device  `json:"children"`
}

// Size accepts lsblk's byte size as a JSON number or string (older
// util-linux releases quote it).
type Size uint64

func (s *Size) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	if str == "null" || str == "" {
		*s = 0
		return nil
	}
	v, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return fmt.Errorf("lsblk size %s: %w", b, err)
	}
	*s = Size(v)
	return nil
}

// lsblkColumns are the fields every lsblk call asks for. lsblk only
// nests children when NAME is among the columns; without it the
// output is a flat list.
const lsblkColumns = "NAME,PATH,TYPE,SIZE,MODEL,SERIAL,PARTTYPE,FSTYPE,MOUNTPOINTS"

// Lsblk returns the device tree rooted at path (the device, its
// partitions and their holders).
func Lsblk(ctx context.Context, path string) (*Device, error) {
	out, err := exec.CommandContext(ctx, "lsblk", "-J", "-b", "-p",
		"-o", lsblkColumns, path).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("lsblk %s: %s", path, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("lsblk %s: %w", path, err)
	}
	return ParseLsblk(out)
}

// Disks returns every whole disk lsblk sees on the system (partitions,
// loops, CD-ROMs and similar are left out).
func Disks(ctx context.Context) ([]*Device, error) {
	out, err := exec.CommandContext(ctx, "lsblk", "-J", "-b", "-p",
		"-o", lsblkColumns).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("lsblk: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("lsblk: %w", err)
	}
	return ParseDisks(out)
}

// ParseDisks parses system-wide `lsblk -J` output and keeps the whole
// disks.
func ParseDisks(b []byte) ([]*Device, error) {
	var r struct {
		BlockDevices []*Device `json:"blockdevices"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("parsing lsblk output: %w", err)
	}
	var out []*Device
	for _, d := range r.BlockDevices {
		if d != nil && d.Type == "disk" {
			out = append(out, d)
		}
	}
	return out, nil
}

// ParseLsblk parses `lsblk -J` output for a single device.
func ParseLsblk(b []byte) (*Device, error) {
	var r struct {
		BlockDevices []Device `json:"blockdevices"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("parsing lsblk output: %w", err)
	}
	if len(r.BlockDevices) != 1 {
		return nil, fmt.Errorf("lsblk returned %d devices, want 1", len(r.BlockDevices))
	}
	return &r.BlockDevices[0], nil
}

// MountedAt returns every mount point of d and its descendants.
func (d *Device) MountedAt() []string {
	var out []string
	for _, m := range d.Mountpoints {
		if m != nil && *m != "" {
			out = append(out, *m)
		}
	}
	for i := range d.Children {
		out = append(out, d.Children[i].MountedAt()...)
	}
	return out
}

// PartitionsOfType returns the direct child partitions whose GPT type
// GUID matches guid (case-insensitive).
func (d *Device) PartitionsOfType(guid string) []Device {
	var out []Device
	for _, c := range d.Children {
		if c.Type == "part" && strings.EqualFold(c.PartType, guid) {
			out = append(out, c)
		}
	}
	return out
}

// Ident is a human-readable "model serial" label for the device.
func (d *Device) Ident() string {
	return strings.Join(strings.Fields(d.Model+" "+d.Serial), " ")
}

// Describe summarizes what is already on the disk: its partitions
// with their filesystem types, or a filesystem formatted straight
// onto the disk. Empty for a blank disk.
func (d *Device) Describe() string {
	if d.FSType != "" {
		return "a " + d.FSType + " filesystem"
	}
	parts, fss := 0, []string{}
	for _, c := range d.Children {
		if c.Type != "part" {
			continue
		}
		parts++
		if c.FSType != "" {
			fss = append(fss, c.FSType)
		}
	}
	if parts == 0 {
		return ""
	}
	s := fmt.Sprintf("%d partition", parts)
	if parts > 1 {
		s += "s"
	}
	if len(fss) > 0 {
		s += " (" + strings.Join(fss, ", ") + ")"
	}
	return s
}

// SysFS answers partition questions from /sys/class/block.
type SysFS struct {
	Root string // normally /sys
}

func (s SysFS) node(name string) string {
	return filepath.Join(s.Root, "class", "block", name)
}

// IsPartition reports whether the kernel block device name is a
// partition (only partitions carry the "partition" attribute).
func (s SysFS) IsPartition(name string) bool {
	_, err := os.Stat(filepath.Join(s.node(name), "partition"))
	return err == nil
}

// DiskOf returns the whole-disk name that holds the block device
// name, or name itself when it is not a partition.
func (s SysFS) DiskOf(name string) (string, error) {
	if !s.IsPartition(name) {
		return name, nil
	}
	real, err := filepath.EvalSymlinks(s.node(name))
	if err != nil {
		return "", fmt.Errorf("resolving %s in sysfs: %w", name, err)
	}
	return filepath.Base(filepath.Dir(real)), nil
}

// Mount is one entry of /proc/self/mountinfo.
type Mount struct {
	Target string
	FSType string
	Source string
}

// ParseMountinfo parses the /proc/<pid>/mountinfo format.
func ParseMountinfo(r io.Reader) ([]Mount, error) {
	var out []Mount
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		pre, post, ok := strings.Cut(sc.Text(), " - ")
		if !ok {
			continue
		}
		pf, sf := strings.Fields(pre), strings.Fields(post)
		if len(pf) < 5 || len(sf) < 2 {
			continue
		}
		out = append(out, Mount{Target: unescape(pf[4]), FSType: sf[0], Source: unescape(sf[1])})
	}
	return out, sc.Err()
}

// unescape decodes the octal escapes (\040 etc.) mountinfo uses.
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+4 <= len(s) {
			if v, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(v))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
