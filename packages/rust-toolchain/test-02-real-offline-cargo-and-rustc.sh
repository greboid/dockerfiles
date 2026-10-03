#!/bin/sh
export PATH="/usr/local/cargo/bin:$PATH"
work=$(mktemp -d)
cd "$work"
printf 'fn main() { println!("rust-musl-ok"); }\n' > hello.rs
rustc --target x86_64-unknown-linux-musl hello.rs -o hello
test "$(./hello)" = rust-musl-ok
cargo new --bin --vcs none smoke
cp hello.rs smoke/src/main.rs
cargo build --offline --manifest-path smoke/Cargo.toml --target x86_64-unknown-linux-musl
test "$(./smoke/target/x86_64-unknown-linux-musl/debug/smoke)" = rust-musl-ok
rm -rf "$work"
