# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"6d131c89e231e9a653da086b59d83f269f52924449a9ab8e7cbd6e86fe35d74f","image:haproxy":"477a750415d205b5351ddc39dfdff64d77fd514143167bb96abfffc120457a39"}

FROM reg.g5d.dev/haproxy@sha256:477a750415d205b5351ddc39dfdff64d77fd514143167bb96abfffc120457a39 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:6d131c89e231e9a653da086b59d83f269f52924449a9ab8e7cbd6e86fe35d74f

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
