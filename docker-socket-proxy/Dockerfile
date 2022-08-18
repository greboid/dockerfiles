# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"58bb78426f91647d2336b955555c415ba8e06d603a9ac20fe45ebf049c4a3739","image:haproxy":"63a0e4f3e43b64d389432c3cd42a13e8786cab853151db34698031f44c19a866"}

FROM reg.g5d.dev/haproxy@sha256:63a0e4f3e43b64d389432c3cd42a13e8786cab853151db34698031f44c19a866 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:58bb78426f91647d2336b955555c415ba8e06d603a9ac20fe45ebf049c4a3739

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
