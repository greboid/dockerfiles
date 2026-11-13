// Package blockdev inspects block devices through lsblk, sysfs and
// mountinfo.
package blockdev

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
// name, or name itself when it is not a partition. Loop devices backed
// by a block device resolve to the disk holding that device.
func (s SysFS) DiskOf(name string) (string, error) {
	for range 8 {
		if s.IsPartition(name) {
			real, err := filepath.EvalSymlinks(s.node(name))
			if err != nil {
				return "", fmt.Errorf("resolving %s in sysfs: %w", name, err)
			}
			name = filepath.Base(filepath.Dir(real))
		}
		backing := s.loopBacking(name)
		if backing == "" {
			return name, nil
		}
		name = backing
	}
	return "", fmt.Errorf("loop device chain too deep at %s", name)
}

// loopBacking returns the block device name behind loop device name, or
// "" when name is not a loop device or is not backed by a block device.
func (s SysFS) loopBacking(name string) string {
	b, err := os.ReadFile(filepath.Join(s.node(name), "loop", "backing_file"))
	if err != nil {
		return ""
	}
	// The path is as seen by whoever attached it, e.g. "/sr0" from systemd-dissect in the initrd.
	path := strings.TrimSpace(string(b))
	if real, err := filepath.EvalSymlinks(path); err == nil {
		path = real
	}
	name = filepath.Base(path)
	if _, err := os.Stat(s.node(name)); err != nil {
		return ""
	}
	return name
}
