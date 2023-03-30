# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"ff5c544832e13e4fd5a43c7edcfb640915db4f6b023714179ab5b46a21d4d629","image:haproxy":"e9e98cb50f6c0b75123c07d9c86e18a5417b9c2368e1aeec0659f1d0fece4a7f"}

FROM reg.g5d.dev/haproxy@sha256:e9e98cb50f6c0b75123c07d9c86e18a5417b9c2368e1aeec0659f1d0fece4a7f AS haproxy

FROM reg.g5d.dev/baseroot@sha256:ff5c544832e13e4fd5a43c7edcfb640915db4f6b023714179ab5b46a21d4d629

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
