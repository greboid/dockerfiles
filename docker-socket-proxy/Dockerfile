# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"fd51da47fdc6d0f778cc71f985d36559355d00bfbefdb6b92718ca47a594760f","image:haproxy":"f6f2af1d579f796bd6c92ac0041216c28a4883ba53074e2b031573c0322227c6"}

FROM reg.g5d.dev/haproxy@sha256:f6f2af1d579f796bd6c92ac0041216c28a4883ba53074e2b031573c0322227c6 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:fd51da47fdc6d0f778cc71f985d36559355d00bfbefdb6b92718ca47a594760f

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
