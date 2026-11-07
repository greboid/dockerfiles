#!/bin/sh
# Per-machine state for the sealed /etc.
#
# /etc is part of the composefs image (read-only, identical on every
# machine, replaced wholesale by upgrades). The few files that must
# differ per machine live in /var/lib/osimage/etc (persistent, shared
# across deployments) and are bind-mounted over their /etc counterparts
# here, before PID 1 starts:
#
#   machine-id  seeded empty: PID 1 finds an empty, writable
#               /etc/machine-id and writes a freshly generated ID
#               through the bind on the first boot (not a "first boot"
#               in systemd's sense: that needs "uninitialized")
#   hostname    seeded from /usr/lib/osimage/etc-seed/hostname, else
#               the image's /etc/hostname; edit the /var copy (or
#               `echo name > /etc/hostname`) and reboot
#   shadow      seeded from the image's /etc/shadow (apko's merged
#               accounts); changed with osimage-passwd, which edits the
#               /var copy in place (passwd(1) would write a temporary
#               file next to /etc/shadow and fails on the sealed /etc)
#
# Each file is only bound when the image ships the /etc target, so
# images without the layout are unaffected. When /var is not writable
# (unexpected: bootc binds it rw, live media run on an overlay), the
# image's files stay visible; PID 1 then falls back to a transient
# machine-id for that boot instead of failing.
command -v getarg > /dev/null || . /lib/dracut-lib.sh

STATE="$NEWROOT/var/lib/osimage/etc"
SEED="$NEWROOT/usr/lib/osimage/etc-seed"

if ! mkdir -p "$STATE" 2> /dev/null; then
    warn "osimage-state: cannot create $STATE; using the image's /etc files"
    return 0
fi
chmod 0755 "$STATE"

for f in machine-id hostname shadow; do
    [ -f "$NEWROOT/etc/$f" ] || continue
    if [ ! -e "$STATE/$f" ]; then
        if [ "$f" = machine-id ]; then
            : > "$STATE/$f.new" && chmod 0444 "$STATE/$f.new"
        elif [ -f "$SEED/$f" ]; then
            cp -p "$SEED/$f" "$STATE/$f.new"
        else
            cp -p "$NEWROOT/etc/$f" "$STATE/$f.new"
        fi || {
            warn "osimage-state: cannot seed $STATE/$f"
            continue
        }
        [ "$f" = shadow ] && chmod 0640 "$STATE/$f.new"
        mv "$STATE/$f.new" "$STATE/$f" || continue
        info "osimage-state: seeded $STATE/$f"
    fi
    mount --bind "$STATE/$f" "$NEWROOT/etc/$f" ||
        warn "osimage-state: cannot bind $STATE/$f over /etc/$f"
done
return 0
