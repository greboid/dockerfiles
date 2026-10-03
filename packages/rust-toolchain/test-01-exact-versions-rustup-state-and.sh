#!/bin/sh
export PATH="/usr/local/cargo/bin:$PATH"
rustc --version | grep '^rustc 1.98.1 '
cargo --version | grep '^cargo 1.98.1 '
rustup --version | grep '^rustup 1.29.1 '
rustup show active-toolchain | grep '^1.98.1-x86_64-unknown-linux-musl (default)'
test "$(rustup show profile)" = minimal
rustup target list --installed | grep '^x86_64-unknown-linux-musl$'
for component in cargo rustc rust-std; do
  rustup component list --installed | grep "^$component-x86_64-unknown-linux-musl"
done
test "$(readlink /usr/local/bin/cargo)" = /usr/local/cargo/bin/cargo
test "$(git config --system advice.detachedHead)" = false
test -z "$(find "$RUSTUP_HOME" "$CARGO_HOME" ! -perm -0002 -print)"
! grep -R '/home/build/' "$CARGO_HOME/env" "$RUSTUP_HOME/settings.toml"
