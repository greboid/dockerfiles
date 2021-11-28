# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"cb2f777a8f1be69c385424a3748112aa34e562abfb51ea4ff74a616071660d14","image:haproxy":"1a93153553a0efcb00f698b0607d6f045748eb2045dc329232a1b723001a5929"}

FROM reg.g5d.dev/haproxy@sha256:1a93153553a0efcb00f698b0607d6f045748eb2045dc329232a1b723001a5929 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:cb2f777a8f1be69c385424a3748112aa34e562abfb51ea4ff74a616071660d14

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
