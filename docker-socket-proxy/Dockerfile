# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"e8223af049600b52310ee1879444afc3d633fb3af91a2139ead53d22b4ad156d","image:haproxy":"d17bb48abbea50eddc2d3b2fd371c548db5cd66b7df815c96c316a6e2b49fa5c"}

FROM reg.g5d.dev/haproxy@sha256:d17bb48abbea50eddc2d3b2fd371c548db5cd66b7df815c96c316a6e2b49fa5c AS haproxy

FROM reg.g5d.dev/baseroot@sha256:e8223af049600b52310ee1879444afc3d633fb3af91a2139ead53d22b4ad156d

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
