# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"64b1a284c9851528f274df911604a3f42b251b4272a3160fb21e4f8b0b912aa7","image:haproxy":"2cd3948e889e5a83535a889c2eff7cf04e9237a5cc5182037d382a4e10d575d2"}

FROM reg.g5d.dev/haproxy@sha256:2cd3948e889e5a83535a889c2eff7cf04e9237a5cc5182037d382a4e10d575d2 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:64b1a284c9851528f274df911604a3f42b251b4272a3160fb21e4f8b0b912aa7

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
