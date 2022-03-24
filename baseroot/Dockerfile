# Generated from https://github.com/greboid/dockerfiles/blob/master/baseroot/Dockerfile.gotpl
# BOM: {"apk:busybox":"1.34.1-r4","apk:ca-certificates":"20211220-r0","apk:libacl":"2.2.53-r0","apk:libcrypto1.1":"1.1.1n-r0","apk:lz4-libs":"1.9.3-r1","apk:musl":"1.2.2-r7","apk:popt":"1.18-r0","apk:rsync":"3.2.3-r5","apk:tzdata":"2022a-r0","apk:zlib":"1.2.11-r3","apk:zstd-libs":"1.5.0-r0","image:alpine":"a4460332f57a36a82ca308dd52b6565829a025b66580248508b45908f0061bc1"}

FROM reg.g5d.dev/alpine@sha256:a4460332f57a36a82ca308dd52b6565829a025b66580248508b45908f0061bc1 AS build

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
        tzdata=2022a-r0 \
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
    # Copy the tzdata files and musl
    apk info -qL tzdata | rsync -aq --files-from=- / /rootfs/; \
    apk info -qL musl | rsync -aq --files-from=- / /rootfs/; \
    # Create special mount points
    mkdir /rootfs/proc /rootfs/dev /rootfs/sys; \
    # Generate thw CA certificates file and copy the result
    /usr/sbin/update-ca-certificates; \
    cp /etc/ssl/certs/ca-certificates.crt /rootfs/etc/ssl/certs/ca-certificates.crt;

FROM scratch

COPY --from=build /rootfs/ /

ENV SSL_CERT_FILE /etc/ssl/certs/ca-certificates.crt
