#!/bin/sh
command -v getarg > /dev/null || . /lib/dracut-lib.sh

STATE="$NEWROOT/var/lib/osimage/etc"
SEED="$NEWROOT/usr/lib/osimage/etc-seed"

if ! mkdir -p "$STATE" 2> /dev/null; then
    warn "osimage-state: cannot create $STATE; using the image's /etc files"
    return 0
fi
chmod 0755 "$STATE"

new_machine_id() {
    IFS=- read -r _a _b _c _d _e < /proc/sys/kernel/random/uuid || return 1
    _id="$_a$_b$_c$_d$_e"
    [ ${#_id} -eq 32 ] || return 1
    printf '%s\n' "$_id"
}

for f in machine-id hostname; do
    [ -f "$NEWROOT/etc/$f" ] || continue
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
# Passwords are always locked unless the image opts out (the live installer);
# $STATE/shadow is left for rollback to older deployments.
if [ -f "$NEWROOT/etc/shadow" ] && [ ! -e "$NEWROOT/usr/lib/osimage/shadow-unlocked" ]; then
    if awk -F: -v OFS=: '$2 !~ /^[!*]/ { $2 = "!" $2 } { print }' \
        "$NEWROOT/etc/shadow" > "$STATE/shadow.locked.new" \
        && chmod 0640 "$STATE/shadow.locked.new" \
        && mv "$STATE/shadow.locked.new" "$STATE/shadow.locked"; then
        src="$STATE/shadow.locked"
    else
        rm -f "$STATE/shadow.locked.new"
        warn "osimage-state: cannot lock /etc/shadow; hiding it"
        src=/dev/null
    fi
    if mount --bind "$src" "$NEWROOT/etc/shadow"; then
        mount -o remount,bind,ro "$NEWROOT/etc/shadow" ||
            warn "osimage-state: cannot make /etc/shadow read-only"
    else
        warn "osimage-state: cannot bind $src over /etc/shadow"
    fi
fi
return 0
