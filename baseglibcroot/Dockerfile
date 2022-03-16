# Generated from https://github.com/greboid/dockerfiles/blob/master/baseglibcroot/Dockerfile.gotpl
# BOM: {"apk:busybox":"1.34.1-r4","apk:ca-certificates":"20211220-r0","apk:libacl":"2.2.53-r0","apk:libc6-compat":"1.2.2-r7","apk:libcrypto1.1":"1.1.1n-r0","apk:libgcc":"10.3.1_git20211027-r0","apk:libstdc++":"10.3.1_git20211027-r0","apk:lz4-libs":"1.9.3-r1","apk:musl":"1.2.2-r7","apk:popt":"1.18-r0","apk:rsync":"3.2.3-r5","apk:tzdata":"2021e-r0","apk:zlib":"1.2.11-r3","apk:zstd-libs":"1.5.0-r0","image:alpine":"25fb69ad519fd7702759c8dea5971c27527220b259f850e1c16b81571f78245d"}

FROM reg.g5d.dev/alpine@sha256:25fb69ad519fd7702759c8dea5971c27527220b259f850e1c16b81571f78245d AS build

RUN set -eux; \
    # Sort out our dependencies (and sources)
    apk add --no-cache \
        busybox=1.34.1-r4 \
        ca-certificates=20211220-r0 \
        libacl=2.2.53-r0 \
        libcrypto1.1=1.1.1n-r0 \
        lz4-libs=1.9.3-r1 \
        musl=1.2.2-r7 \
        popt=1.18-r0 \
        rsync=3.2.3-r5 \
        zlib=1.2.11-r3 \
        zstd-libs=1.5.0-r0 \
        ; \
    # Set up directories and password file
    mkdir -p /rootfs/etc/ssl/certs /rootfs/tmp /rootfs/home/nonroot /rootfs/root; \
    echo "root:x:0:0:root:/root:/sbin/nologin" >> /rootfs/etc/passwd; \
    echo "nobody:x:65534:65534:nobody:/nonexistent:/sbin/nologin" >> /rootfs/etc/passwd; \
    echo "nonroot:x:65532:65532:nonroot:/home/nonroot:/sbin/nologin" >> /rootfs/etc/passwd; \
    chown 65532:65532 /rootfs/home/nonroot; \
    chmod 01777 /tmp; \
    # Copy packages into rootfs
    apk add --no-cache libc6-compat=1.2.2-r7; \
      apk info -qL libc6-compat | rsync -aq --files-from=- / /rootfs/; \
    apk add --no-cache libgcc=10.3.1_git20211027-r0; \
      apk info -qL libgcc | rsync -aq --files-from=- / /rootfs/; \
    apk add --no-cache libstdc++=10.3.1_git20211027-r0; \
      apk info -qL libstdc++ | rsync -aq --files-from=- / /rootfs/; \
    apk add --no-cache musl=1.2.2-r7; \
      apk info -qL musl | rsync -aq --files-from=- / /rootfs/; \
    apk add --no-cache tzdata=2021e-r0; \
      apk info -qL tzdata | rsync -aq --files-from=- / /rootfs/; \
    # Create special mount points
    mkdir /rootfs/proc /rootfs/dev /rootfs/sys; \
    # Generate thw CA certificates file and copy the result
    /usr/sbin/update-ca-certificates; \
    cp /etc/ssl/certs/ca-certificates.crt /rootfs/etc/ssl/certs/ca-certificates.crt;

FROM scratch

COPY --from=build /rootfs/ /

ENV SSL_CERT_FILE /etc/ssl/certs/ca-certificates.crt
