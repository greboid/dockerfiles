# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"4c1fe9d49d484cca683071563fdf374e20757dfac9fa2a7e4bdead93343c1975","image:haproxy":"7bfca46634b6acca8e9827bb93e59566079b31d3ef56b76e8fcd51cb52bdeac1"}

FROM reg.g5d.dev/haproxy@sha256:7bfca46634b6acca8e9827bb93e59566079b31d3ef56b76e8fcd51cb52bdeac1 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:4c1fe9d49d484cca683071563fdf374e20757dfac9fa2a7e4bdead93343c1975

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
