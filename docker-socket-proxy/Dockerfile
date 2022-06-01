# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"fae0414d67593373dbb0cba1d6e3b740837e3b8873f44fd91d5639ab3acdf952","image:haproxy":"99f4da35fd0629dc427e29b4f932f39071aa3032dfb746de30dbdc0be7b68d37"}

FROM reg.g5d.dev/haproxy@sha256:99f4da35fd0629dc427e29b4f932f39071aa3032dfb746de30dbdc0be7b68d37 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:fae0414d67593373dbb0cba1d6e3b740837e3b8873f44fd91d5639ab3acdf952

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
