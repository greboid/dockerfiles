# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"4c77ca91e924927dc19fdbbb6cb4ca76f3471ce834b97becb2a4757f1bea0092","image:haproxy":"90790e17696e4db23104ce78e631bf586e6cf1400479fa930b27d59e83a9d9f1"}

FROM reg.g5d.dev/haproxy@sha256:90790e17696e4db23104ce78e631bf586e6cf1400479fa930b27d59e83a9d9f1 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:4c77ca91e924927dc19fdbbb6cb4ca76f3471ce834b97becb2a4757f1bea0092

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
