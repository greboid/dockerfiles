# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"5100f1fc0d5d449eb7aee39359429531f3b0806c013ce5e09e58f73a74b8ffa1","image:haproxy":"cf2a3c3689c95fb7148d04ba5674e6956a37ba5104812991e4a2f3acfb470849"}

FROM reg.g5d.dev/haproxy@sha256:cf2a3c3689c95fb7148d04ba5674e6956a37ba5104812991e4a2f3acfb470849 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:5100f1fc0d5d449eb7aee39359429531f3b0806c013ce5e09e58f73a74b8ffa1

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
