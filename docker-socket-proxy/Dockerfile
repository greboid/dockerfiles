# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"5100f1fc0d5d449eb7aee39359429531f3b0806c013ce5e09e58f73a74b8ffa1","image:haproxy":"1aba6648182409324d836b2eeeeb5a83ae36cea959af1b1c7e65e0f9465240fa"}

FROM reg.g5d.dev/haproxy@sha256:1aba6648182409324d836b2eeeeb5a83ae36cea959af1b1c7e65e0f9465240fa AS haproxy

FROM reg.g5d.dev/baseroot@sha256:5100f1fc0d5d449eb7aee39359429531f3b0806c013ce5e09e58f73a74b8ffa1

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
