# Generated from https://github.com/greboid/dockerfiles/blob/master/baseroot/Dockerfile.gotpl
# BOM: {"apk:busybox":"1.35.0-r29","apk:busybox-binsh":"1.35.0-r29","apk:ca-certificates":"20220614-r4","apk:libacl":"2.3.1-r1","apk:libcrypto3":"3.0.8-r4","apk:lz4-libs":"1.9.4-r1","apk:musl":"1.2.3-r4","apk:popt":"1.19-r0","apk:rsync":"3.2.7-r0","apk:tzdata":"2023c-r0","apk:zlib":"1.2.13-r0","apk:zstd-libs":"1.5.5-r0","image:alpine":"e05ad3477f38c022bccca78bf22d39f9355bd07a65f4b7d224324c328a84d51b"}

FROM reg.g5d.dev/alpine@sha256:e05ad3477f38c022bccca78bf22d39f9355bd07a65f4b7d224324c328a84d51b AS build

RUN set -eux; \
    # Sort out our dependencies (and sources)
    apk add --no-cache \
        busybox=1.35.0-r29 \
        busybox-binsh=1.35.0-r29 \
        ca-certificates=20220614-r4 \
        libacl=2.3.1-r1 \
        libcrypto3=3.0.8-r4 \
        lz4-libs=1.9.4-r1 \
        musl=1.2.3-r4 \
        popt=1.19-r0 \
        rsync=3.2.7-r0 \
        tzdata=2023c-r0 \
        zlib=1.2.13-r0 \
        zstd-libs=1.5.5-r0 \
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
    cp /etc/ssl/certs/ca-certificates.crt /rootfs/etc/ssl/certs/ca-certificates.crt; \
    echo "[advice]" > /rootfs/etc/gitconfig; \
    echo "    detachedHead = false" >> /rootfs/etc/gitconfig;

FROM scratch

COPY --from=build /rootfs/ /

ENV SSL_CERT_FILE /etc/ssl/certs/ca-certificates.crt
