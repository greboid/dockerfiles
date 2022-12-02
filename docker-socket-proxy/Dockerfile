# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"1c975ebed03f0278f2a9aecd9637ea97b7d5b0efb5543da9f1e56c83b84450f0","image:haproxy":"0794814c052a66f27b695759db1c5ab15d70c2765d5e202aadbac9ae681c79df"}

FROM reg.g5d.dev/haproxy@sha256:0794814c052a66f27b695759db1c5ab15d70c2765d5e202aadbac9ae681c79df AS haproxy

FROM reg.g5d.dev/baseroot@sha256:1c975ebed03f0278f2a9aecd9637ea97b7d5b0efb5543da9f1e56c83b84450f0

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
