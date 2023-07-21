# Generated from https://github.com/greboid/dockerfiles/blob/master/baseglibcroot/Containerfile.gotpl
# BOM: {"apk:libacl":"2.3.1-r3","apk:libc6-compat":"1.2.4-r0","apk:libgcc":"12.2.1_git20220924-r10","apk:libstdc++":"12.2.1_git20220924-r10","apk:libxxhash":"0.8.1-r4","apk:lz4-libs":"1.9.4-r4","apk:musl":"1.2.4-r0","apk:popt":"1.19-r2","apk:rsync":"3.2.7-r4","apk:zlib":"1.2.13-r1","apk:zstd-libs":"1.5.5-r4","image:alpine":"347bebd1a359b0cccf62ebd4f65854ddf25331f9458011a2cc06a5e1679e3578","image:baseroot":"0ff388c461add6a09dde64f3bfe3cb62e3ca09f1fb4b69ca93a2f0aa0ae68137"}

FROM reg.g5d.dev/alpine@sha256:347bebd1a359b0cccf62ebd4f65854ddf25331f9458011a2cc06a5e1679e3578 AS build

RUN set -eux; \
    apk add --no-cache libacl=2.3.1-r3; \
    apk add --no-cache libxxhash=0.8.1-r4; \
    apk add --no-cache lz4-libs=1.9.4-r4; \
    apk add --no-cache musl=1.2.4-r0; \
    apk add --no-cache popt=1.19-r2; \
    apk add --no-cache rsync=3.2.7-r4; \
    apk add --no-cache zlib=1.2.13-r1; \
    apk add --no-cache zstd-libs=1.5.5-r4; \
    # Add packages into rootfs
    apk add --no-cache libc6-compat=1.2.4-r0; \
      apk info -qL libc6-compat | rsync -aq --files-from=- / /rootfs/; \
    apk add --no-cache libgcc=12.2.1_git20220924-r10; \
      apk info -qL libgcc | rsync -aq --files-from=- / /rootfs/; \
    apk add --no-cache libstdc++=12.2.1_git20220924-r10; \
      apk info -qL libstdc++ | rsync -aq --files-from=- / /rootfs/; \
    apk add --no-cache musl=1.2.4-r0; \
      apk info -qL musl | rsync -aq --files-from=- / /rootfs/; \
    echo;

FROM reg.g5d.dev/baseroot@sha256:0ff388c461add6a09dde64f3bfe3cb62e3ca09f1fb4b69ca93a2f0aa0ae68137

COPY --from=build /rootfs/ /
