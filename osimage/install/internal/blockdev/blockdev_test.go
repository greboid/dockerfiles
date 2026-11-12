package blockdev

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const lsblkOut = `{
   "blockdevices": [
      {
         "path": "/dev/vda",
         "type": "disk",
         "size": 21474836480,
         "model": "QEMU  HARDDISK",
         "serial": "abc123",
         "parttype": null,
         "mountpoints": [null],
         "children": [
            {
               "path": "/dev/vda1",
               "type": "part",
               "size": "1073741824",
               "model": null,
               "serial": null,
               "parttype": "c12a7328-f81f-11d2-ba4b-00a0c93ec93b",
               "mountpoints": ["/boot/efi"]
            },
            {
               "path": "/dev/vda2",
               "type": "part",
               "size": 20399976448,
               "model": null,
               "serial": null,
               "parttype": "4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709",
               "mountpoints": [null],
               "children": [
                  {"path": "/dev/mapper/x", "type": "crypt", "size": 1, "mountpoints": ["/mnt/a", "/mnt/b"]}
               ]
            }
         ]
      }
   ]
}`

func TestParseLsblk(t *testing.T) {
	d, err := ParseLsblk([]byte(lsblkOut))
	if err != nil {
		t.Fatal(err)
	}
	if d.Path != "/dev/vda" || d.Size != 21474836480 || d.Children[0].Size != 1073741824 {
		t.Errorf("unexpected device %+v", d)
	}
	if got := d.Ident(); got != "QEMU HARDDISK abc123" {
		t.Errorf("Ident() = %q", got)
	}
	if got, want := d.MountedAt(), []string{"/boot/efi", "/mnt/a", "/mnt/b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("MountedAt() = %v, want %v", got, want)
	}
	roots := d.PartitionsOfType("4f68bce3-e8cd-4db1-96e7-fbcaf984b709")
	if len(roots) != 1 || roots[0].Path != "/dev/vda2" {
		t.Errorf("PartitionsOfType() = %+v", roots)
	}
}

func TestParseLsblkErrors(t *testing.T) {
	for _, in := range []string{`not json`, `{"blockdevices": []}`, `{"blockdevices": [{"size": "x"}]}`} {
		if _, err := ParseLsblk([]byte(in)); err == nil {
			t.Errorf("ParseLsblk(%q): want error", in)
		}
	}
}

// TestParseDisks uses system-wide lsblk output: whole disks are kept,
// loops, ROMs and null entries dropped.
func TestParseDisks(t *testing.T) {
	in := `{
  "blockdevices": [
     {"path": "/dev/vda", "type": "disk", "size": 21474836480, "model": "QEMU  HARDDISK", "mountpoints": [null]},
     {"path": "/dev/loop0", "type": "loop", "size": 1, "mountpoints": [null]},
     {"path": "/dev/sr0", "type": "rom", "size": 2, "mountpoints": [null]},
     null
  ]
}`
	disks, err := ParseDisks([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(disks) != 1 || disks[0].Path != "/dev/vda" || disks[0].Type != "disk" {
		t.Errorf("ParseDisks() = %+v, want only /dev/vda", disks)
	}
}

func TestDescribe(t *testing.T) {
	for _, tt := range []struct {
		name string
		d    Device
		want string
	}{
		{"blank", Device{Path: "/dev/vda", Type: "disk", Size: 1 << 31}, ""},
		{"two partitions", Device{
			Path: "/dev/sda", Type: "disk", Size: 1 << 32,
			Children: []Device{
				{Path: "/dev/sda1", Type: "part", FSType: "vfat"},
				{Path: "/dev/sda2", Type: "part", FSType: "ext4"},
			},
		}, "2 partitions (vfat, ext4)"},
		{"bare partition", Device{
			Path: "/dev/sda", Type: "disk", Size: 1 << 32,
			Children: []Device{{Path: "/dev/sda1", Type: "part"}},
		}, "1 partition"},
		{"filesystem on the disk", Device{Path: "/dev/sda", Type: "disk", Size: 1 << 32, FSType: "vfat"}, "a vfat filesystem"},
		{"non-partition children ignored", Device{
			Path: "/dev/sda", Type: "disk", Size: 1 << 32,
			Children: []Device{{Path: "/dev/mapper/x", Type: "crypt", FSType: "ext4"}},
		}, ""},
	} {
		if got := tt.d.Describe(); got != tt.want {
			t.Errorf("%s: Describe() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestSysFS(t *testing.T) {
	root := t.TempDir()
	dev := filepath.Join(root, "devices", "pci0000:00", "block", "sda")
	must(t, os.MkdirAll(filepath.Join(dev, "sda1"), 0o755))
	must(t, os.WriteFile(filepath.Join(dev, "sda1", "partition"), []byte("1\n"), 0o644))
	class := filepath.Join(root, "class", "block")
	must(t, os.MkdirAll(class, 0o755))
	must(t, os.Symlink("../../devices/pci0000:00/block/sda", filepath.Join(class, "sda")))
	must(t, os.Symlink("../../devices/pci0000:00/block/sda/sda1", filepath.Join(class, "sda1")))

	s := SysFS{Root: root}
	if s.IsPartition("sda") || !s.IsPartition("sda1") {
		t.Error("IsPartition: wrong answer")
	}
	for name, want := range map[string]string{"sda": "sda", "sda1": "sda", "loop0": "loop0"} {
		got, err := s.DiskOf(name)
		if err != nil || got != want {
			t.Errorf("DiskOf(%s) = %q, %v; want %q", name, got, err, want)
		}
	}
}

func TestParseMountinfo(t *testing.T) {
	in := strings.Join([]string{
		`22 1 0:21 / / rw,relatime shared:1 - overlay LiveOS_rootfs rw,lowerdir=/run/rootfsbase`,
		`30 22 11:0 / /run/initramfs/live ro,relatime shared:9 - iso9660 /dev/sr0 ro`,
		`31 22 8:17 / /mnt/with\040space rw - ext4 /dev/sdb1 rw`,
		`garbage`,
	}, "\n")
	got, err := ParseMountinfo(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := []Mount{
		{Target: "/", FSType: "overlay", Source: "LiveOS_rootfs"},
		{Target: "/run/initramfs/live", FSType: "iso9660", Source: "/dev/sr0"},
		{Target: "/mnt/with space", FSType: "ext4", Source: "/dev/sdb1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseMountinfo() = %+v, want %+v", got, want)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// TestLsblkLive runs the host's lsblk: each disk must come back as one
// device with its partitions nested, not as a flat list.
func TestLsblkLive(t *testing.T) {
	if _, err := exec.LookPath("lsblk"); err != nil {
		t.Skip("lsblk not installed")
	}
	disks, err := Disks(t.Context())
	if err != nil {
		t.Skip(err)
	}
	for _, d := range disks {
		dev, err := Lsblk(t.Context(), d.Path)
		if err != nil {
			t.Fatal(err)
		}
		if len(dev.Children) != len(d.Children) {
			t.Errorf("%s: Lsblk has %d children, Disks has %d", d.Path, len(dev.Children), len(d.Children))
		}
	}
}
