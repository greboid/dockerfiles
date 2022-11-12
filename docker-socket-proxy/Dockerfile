# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"97f890f7e595379178a16d7ce11556a7513e39a04b0f6e1feefdf6d5d4b7e95d","image:haproxy":"7cbb457e987fb0b1ab0e626df35e18d165852d2fa028be56d9b167942c1e43c1"}

FROM reg.g5d.dev/haproxy@sha256:7cbb457e987fb0b1ab0e626df35e18d165852d2fa028be56d9b167942c1e43c1 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:97f890f7e595379178a16d7ce11556a7513e39a04b0f6e1feefdf6d5d4b7e95d

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
