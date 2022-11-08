# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"8ccee4a219c0e0c7b135814c4c423efd774ba75103482c83e5bc9c1ed9f127d8","image:haproxy":"841fc227ae10161bc8051b4d838119d4404b0b0420af4250c08ff46339cdc9e8"}

FROM reg.g5d.dev/haproxy@sha256:841fc227ae10161bc8051b4d838119d4404b0b0420af4250c08ff46339cdc9e8 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:8ccee4a219c0e0c7b135814c4c423efd774ba75103482c83e5bc9c1ed9f127d8

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
