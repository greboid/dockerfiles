# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"019bd3cd845b6b093fd35645aa2e8266834428f2caee98027b425d2b56b7c8e1","image:haproxy":"76bdaff85ef39cd2450497d5b33ed5cc3a44a11de568637c4396b59ca979c8b1"}

FROM reg.g5d.dev/haproxy@sha256:76bdaff85ef39cd2450497d5b33ed5cc3a44a11de568637c4396b59ca979c8b1 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:019bd3cd845b6b093fd35645aa2e8266834428f2caee98027b425d2b56b7c8e1

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
