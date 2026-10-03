#!/bin/sh
mirror="$(pwd)/rust-dist/dist"
mkdir -p "$mirror/2026-09-03"
cp ./*-${PACKAGE_VERSION}-x86_64-unknown-linux-musl.tar.xz "$mirror/2026-09-03/"
# Manifest component URLs are absolute, so redirect them to the local
# mirror too. The original manifest and every component were verified
# above; no mutable upstream checksum is trusted during installation.
sed "s|https://static.rust-lang.org/dist|file://$mirror|g" \
  channel-rust-${PACKAGE_VERSION}.toml > "$mirror/channel-rust-${PACKAGE_VERSION}.toml"
sha256sum "$mirror/channel-rust-${PACKAGE_VERSION}.toml" \
  > "$mirror/channel-rust-${PACKAGE_VERSION}.toml.sha256"
export RUSTUP_HOME="${DESTDIR}/usr/local/rustup"
export CARGO_HOME="${DESTDIR}/usr/local/cargo"
export RUSTUP_DIST_SERVER="file://$(pwd)/rust-dist"
export RUSTUP_INIT_SKIP_PATH_CHECK=yes
chmod +x rustup-init
./rustup-init -y --no-modify-path --profile minimal \
  --default-toolchain ${PACKAGE_VERSION} --default-host x86_64-unknown-linux-musl
# Do not leave build-root paths in the installed shell helper or stored
# manifest: downstream rustup operations must continue to use upstream.
sed -i "s|${DESTDIR}||g" "$CARGO_HOME/env"
find "$RUSTUP_HOME" -name multirust-channel-manifest.toml -exec \
  sed -i "s|file://$mirror|https://static.rust-lang.org/dist|g" {} \;
# Rustup stores the first 20 characters of the channel manifest hash.
# Restore the upstream value alongside the upstream manifest URLs.
printf '%s' "$(sha256sum channel-rust-${PACKAGE_VERSION}.toml | cut -c1-20)" \
  > "$RUSTUP_HOME/update-hashes/${PACKAGE_VERSION}-x86_64-unknown-linux-musl"
find "$RUSTUP_HOME" -name components -exec sort {} -o {} \;
rm -rf "$RUSTUP_HOME/downloads" "$RUSTUP_HOME/tmp"
chmod -R a+w "$RUSTUP_HOME" "$CARGO_HOME"
mkdir -p "${DESTDIR}/usr/local/bin" "${DESTDIR}/etc"
ln -s /usr/local/cargo/bin/cargo "${DESTDIR}/usr/local/bin/cargo"
# Inherited from the legacy Alpine base, not the Rust Containerfile.
printf '[advice]\n    detachedHead = false\n' > "${DESTDIR}/etc/gitconfig"
install -d "${DESTDIR}/usr/share/licenses/rust-toolchain"
# busybox tar cannot parse these GNU-format archives, so install and
# use Alpine's GNU tar explicitly (build-time dependency only; its apk
# replaces busybox's /bin/tar symlink for the build).
for license in LICENSE-MIT LICENSE-APACHE COPYRIGHT; do
  /bin/tar -xJOf rustc-${PACKAGE_VERSION}-x86_64-unknown-linux-musl.tar.xz \
    rustc-${PACKAGE_VERSION}-x86_64-unknown-linux-musl/$license \
    > "${DESTDIR}/usr/share/licenses/rust-toolchain/$license"
  test -s "${DESTDIR}/usr/share/licenses/rust-toolchain/$license"
done
