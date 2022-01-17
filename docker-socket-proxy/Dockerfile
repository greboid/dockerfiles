# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"4c77ca91e924927dc19fdbbb6cb4ca76f3471ce834b97becb2a4757f1bea0092","image:haproxy":"afbbcd2de2a470b04fdd3501efb054d7870ea072631405c2670e8693552dda27"}

FROM reg.g5d.dev/haproxy@sha256:afbbcd2de2a470b04fdd3501efb054d7870ea072631405c2670e8693552dda27 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:4c77ca91e924927dc19fdbbb6cb4ca76f3471ce834b97becb2a4757f1bea0092

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
