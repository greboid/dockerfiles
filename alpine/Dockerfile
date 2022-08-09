# Generated from https://github.com/greboid/dockerfiles/blob/master/alpine/Dockerfile.gotpl
# BOM: {"alpine":"3.16.2"}

FROM reg.g5d.dev/alpine AS verify

ARG ARCHIVE_URL="https://mirrors.melbourne.co.uk/alpine/latest-stable/releases/x86_64/alpine-minirootfs-3.16.2-x86_64.tar.gz"
ARG ARCHIVE_SUM="bea608c9ae0190096f1671dfbef9d6868701681e3870eecad658dde032c5885c"

RUN set -eux; \
    apk add --no-cache curl; \
    curl -sfLo fs.tar.gz $ARCHIVE_URL; \
    echo "$ARCHIVE_SUM *fs.tar.gz" | sha256sum -wc -; \
    mkdir /fs; \
    tar -C /fs -xzf fs.tar.gz;

FROM scratch
COPY --from=verify /fs/ /

ENV LANG=en_US.UTF-8
CMD ["/bin/sh"]
