# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"1c975ebed03f0278f2a9aecd9637ea97b7d5b0efb5543da9f1e56c83b84450f0","image:haproxy":"1f8ed8e3cd0a7d5d76d2f00017cdf0f06a4ed810cbbfe90fd1a1abbe5f35bc5a"}

FROM reg.g5d.dev/haproxy@sha256:1f8ed8e3cd0a7d5d76d2f00017cdf0f06a4ed810cbbfe90fd1a1abbe5f35bc5a AS haproxy

FROM reg.g5d.dev/baseroot@sha256:1c975ebed03f0278f2a9aecd9637ea97b7d5b0efb5543da9f1e56c83b84450f0

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
