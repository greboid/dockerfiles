# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"ff5c544832e13e4fd5a43c7edcfb640915db4f6b023714179ab5b46a21d4d629","image:haproxy":"337f6db2cf4698c51199a5cd9f3908e52233705212a33234ccf6b14ee6b42306"}

FROM reg.g5d.dev/haproxy@sha256:337f6db2cf4698c51199a5cd9f3908e52233705212a33234ccf6b14ee6b42306 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:ff5c544832e13e4fd5a43c7edcfb640915db4f6b023714179ab5b46a21d4d629

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
