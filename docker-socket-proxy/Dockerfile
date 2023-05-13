# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"e8223af049600b52310ee1879444afc3d633fb3af91a2139ead53d22b4ad156d","image:haproxy":"cfd4dfff788581525850dd20e3ee286c3b1c3593bfa2432429ef60e74401a16c"}

FROM reg.g5d.dev/haproxy@sha256:cfd4dfff788581525850dd20e3ee286c3b1c3593bfa2432429ef60e74401a16c AS haproxy

FROM reg.g5d.dev/baseroot@sha256:e8223af049600b52310ee1879444afc3d633fb3af91a2139ead53d22b4ad156d

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
