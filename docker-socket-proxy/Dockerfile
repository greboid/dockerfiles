# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"64873d5cd1e5603150a658574295b3b617f2213f97e03caf9530728e92006780","image:haproxy":"c244f81879e24a9dd00d49de193a3de6cf33e557ccc76596cd6163cf61c9032c"}

FROM reg.g5d.dev/haproxy@sha256:c244f81879e24a9dd00d49de193a3de6cf33e557ccc76596cd6163cf61c9032c AS haproxy

FROM reg.g5d.dev/baseroot@sha256:64873d5cd1e5603150a658574295b3b617f2213f97e03caf9530728e92006780

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
