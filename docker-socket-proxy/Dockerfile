# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"sha256:5dea99b6fc9ffe51ff467e5ae1f2ccd6157090f4d20db63d306faa7fe7adac51","image:haproxy":"sha256:e2a8d9cc81fbfb32f694e45091f6999deb978d48e2ac33d482d790ad0f5c03e6"}

FROM reg.g5d.dev/haproxy@sha256:e2a8d9cc81fbfb32f694e45091f6999deb978d48e2ac33d482d790ad0f5c03e6 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:5dea99b6fc9ffe51ff467e5ae1f2ccd6157090f4d20db63d306faa7fe7adac51

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
