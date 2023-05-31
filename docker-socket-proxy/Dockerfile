# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"c0e2307d35f12546a86a703c9366d018619d61dae48cd16ebc6c10a51d3d7d91","image:haproxy":"eb1d5b2d880b6213bbd04cf037212f1c47d5226930fca2ae33a382a113a1dfc3"}

FROM reg.g5d.dev/haproxy@sha256:eb1d5b2d880b6213bbd04cf037212f1c47d5226930fca2ae33a382a113a1dfc3 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:c0e2307d35f12546a86a703c9366d018619d61dae48cd16ebc6c10a51d3d7d91

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
