# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"4c77ca91e924927dc19fdbbb6cb4ca76f3471ce834b97becb2a4757f1bea0092","image:haproxy":"3dbd00927e6e3dd11f7e0e97460af9ef384159c9eaea8e493ea2296630fff718"}

FROM reg.g5d.dev/haproxy@sha256:3dbd00927e6e3dd11f7e0e97460af9ef384159c9eaea8e493ea2296630fff718 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:4c77ca91e924927dc19fdbbb6cb4ca76f3471ce834b97becb2a4757f1bea0092

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
