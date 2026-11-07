#!/bin/sh
# Per-machine state for the sealed /etc.
#
# /etc is part of the composefs image (read-only, identical on every
# machine, replaced wholesale by upgrades). The few files that must
# differ per machine live in /var/lib/osimage/etc (persistent, shared
# across deployments) and are bind-mounted read-only over their /etc
# counterparts here, before PID 1 starts:
#
#   machine-id  written by the installer (osimage-install); generated
#               here when missing or empty (installs made without it,
#               or after deleting the file to reseed)
#   hostname    written by the installer; seeded from
#               /usr/lib/osimage/etc-seed/hostname when missing, else
#               the image's /etc/hostname. Edit the /var copy and reboot
#   shadow      written by the installer; seeded from the image's
#               /etc/shadow when missing. Changed with osimage-passwd,
#               which edits the /var copy in place
#
# The binds are read-only: the files change only through their /var
# copies, never through /etc. Each file is only bound when the image
# ships the /etc target, so images without the layout are unaffected.
# When /var is not writable (unexpected: bootc binds it rw, live media
# run on an overlay), the image's files stay visible; PID 1 then falls
# back to a transient machine-id for that boot instead of failing.
command -v getarg > /dev/null || . /lib/dracut-lib.sh

STATE="$NEWROOT/var/lib/osimage/etc"
SEED="$NEWROOT/usr/lib/osimage/etc-seed"

if ! mkdir -p "$STATE" 2> /dev/null; then
    warn "osimage-state: cannot create $STATE; using the image's /etc files"
    return 0
fi
chmod 0755 "$STATE"

# A random machine ID in systemd's format: 32 lowercase hex digits.
new_machine_id() {
    IFS=- read -r _a _b _c _d _e < /proc/sys/kernel/random/uuid || return 1
    _id="$_a$_b$_c$_d$_e"
    [ ${#_id} -eq 32 ] || return 1
    printf '%s\n' "$_id"
}

for f in machine-id hostname shadow; do
    [ -f "$NEWROOT/etc/$f" ] || continue
    # an empty machine-id (seeded by an older initramfs that left it to
    # PID 1) cannot be filled in through a read-only bind: regenerate
    if [ "$f" = machine-id ] && [ -f "$STATE/$f" ] && [ ! -s "$STATE/$f" ]; then
        rm -f "$STATE/$f"
    fi
    if [ ! -e "$STATE/$f" ]; then
        if [ "$f" = machine-id ]; then
            new_machine_id > "$STATE/$f.new" && chmod 0444 "$STATE/$f.new"
        elif [ -f "$SEED/$f" ]; then
            cp -p "$SEED/$f" "$STATE/$f.new"
        else
            cp -p "$NEWROOT/etc/$f" "$STATE/$f.new"
        fi || {
            rm -f "$STATE/$f.new"
            warn "osimage-state: cannot seed $STATE/$f"
            continue
        }
        [ "$f" = shadow ] && chmod 0640 "$STATE/$f.new"
        mv "$STATE/$f.new" "$STATE/$f" || continue
        info "osimage-state: seeded $STATE/$f"
    fi
    if ! mount --bind "$STATE/$f" "$NEWROOT/etc/$f"; then
        warn "osimage-state: cannot bind $STATE/$f over /etc/$f"
        continue
    fi
    mount -o remount,bind,ro "$NEWROOT/etc/$f" ||
        warn "osimage-state: cannot make /etc/$f read-only"
done
return 0
