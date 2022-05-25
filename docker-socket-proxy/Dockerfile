# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"fae0414d67593373dbb0cba1d6e3b740837e3b8873f44fd91d5639ab3acdf952","image:haproxy":"d531cd282b16f0fe4301606b3d9ed5c291d2275273aeb00e0e007ddca83d006d"}

FROM reg.g5d.dev/haproxy@sha256:d531cd282b16f0fe4301606b3d9ed5c291d2275273aeb00e0e007ddca83d006d AS haproxy

FROM reg.g5d.dev/baseroot@sha256:fae0414d67593373dbb0cba1d6e3b740837e3b8873f44fd91d5639ab3acdf952

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
