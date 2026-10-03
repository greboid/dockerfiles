#!/bin/sh
# --no-optional (like upstream) drops sharp's @img prebuilts; sharp
# is then source-built against the system libvips package — its
# prebuilts bundle a libheif without HEVC codecs, so iPhone HEIC
# would fail to decode. SHARP_FORCE_GLOBAL_LIBVIPS switches its
# binding.gyp from the bundled libvips to pkg-config's.
pnpm --filter immich --prod --no-optional deploy /home/build/out/server
# sharp's binding.gyp adds -flto=auto when detect-libc classifies the
# host as glibc — it misdetects inside the build guest, and gcc15 LTO
# plus Alpine's fortify headers fail the link. Build without LTO.
# pnpm deploys sharp as a symlink into .pnpm — patch every copy of
# the gyp file (the file node-gyp reads lives in the real store dir).
find out/server/node_modules -name binding.gyp -path '*sharp*' \
  -exec sed -i 's/-flto=auto//' {} +
# The sharp build compiles+links with node's common.gypi LTO flags
# (enable_lto leaks in from the guest node's build config), and
# gcc15 LTO + Alpine's fortify headers fail the link. Strip the LTO
# flags from the generated makefile between configure and build.
cd out/server/node_modules/sharp
npx node-gyp configure --directory=src
sed -i 's/ -flto=4 -ffat-lto-objects//' src/build/*.mk
npx node-gyp build --directory=src
cd /home/build
# Cross-platform native prebuilds that can never load on musl are
# dead weight: glibc builds and other-OS/arch prebuild dirs.
find /home/build/out/server/node_modules -name '*.glibc.node' -delete
find /home/build/out/server/node_modules -type d \( \
  -path '*/prebuilds/android-*' -o -path '*/prebuilds/ios-*' -o \
  -path '*/prebuilds/darwin-*' -o -path '*/prebuilds/win32-*' -o \
  -path '*bare-fs*/prebuilds' -o -path '*bare-path*/prebuilds' -o \
  -path '*bare-url*/prebuilds' \) -exec rm -rf {} +
# Non-dist assets the server reads at runtime, if present.
if [ -d server/resources ]; then cp -a server/resources /home/build/out/server/; fi
