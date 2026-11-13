# common

Shared build, seal and publish scripts and reusable workflows for the
osimage image repositories.

- `keys/`: the org repository signing key (`packages@<sha256>.rsa.pub`). The
  same key signs the packages, which name it `packages.rsa.pub` (a symlink), so
  `build-installer-iso.sh` can verify the live initramfs package.
- `repart.d/`: systemd-repart definitions for the installer ISO (ESP with the
  live UKI, squashfs root partition, El Torito catalog).
