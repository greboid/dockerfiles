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

for f in machine-id hostname shadow; do
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
        [ "$f" = shadow ] && chmod 0640 "$STATE/$f.new"
        mv "$STATE/$f.new" "$STATE/$f" || continue
        info "osimage-state: seeded $STATE/$f"
    fi
    if [ "$f" = shadow ] && [ -f "$STATE/$f" ]; then
        added="$STATE/$f.merge-accounts.$$"
        if awk -F: 'NR == FNR { seen[$1] = 1; next } !($1 in seen) { print }' \
            "$STATE/$f" "$NEWROOT/etc/$f" > "$added"; then
            if [ -s "$added" ]; then
                merged="$STATE/$f.new"
                if cat "$STATE/$f" "$added" > "$merged" \
                    && chmod 0640 "$merged" \
                    && mv "$merged" "$STATE/$f"; then
                    info "osimage-state: merged new accounts into $STATE/$f"
                else
                    rm -f "$merged"
                    warn "osimage-state: cannot merge new accounts into $STATE/$f"
                fi
            fi
        else
            warn "osimage-state: cannot diff accounts between /etc/shadow and $STATE/shadow"
        fi
        rm -f "$added"
    fi
    if ! mount --bind "$STATE/$f" "$NEWROOT/etc/$f"; then
        warn "osimage-state: cannot bind $STATE/$f over /etc/$f"
        continue
    fi
    mount -o remount,bind,ro "$NEWROOT/etc/$f" ||
        warn "osimage-state: cannot make /etc/$f read-only"
done
return 0
