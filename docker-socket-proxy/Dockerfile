# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"fd51da47fdc6d0f778cc71f985d36559355d00bfbefdb6b92718ca47a594760f","image:haproxy":"03dffb2a1b52b6290c7008602eb2dadd6ecb1e10c54fb09c0d1f08cab4afbe46"}

FROM reg.g5d.dev/haproxy@sha256:03dffb2a1b52b6290c7008602eb2dadd6ecb1e10c54fb09c0d1f08cab4afbe46 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:fd51da47fdc6d0f778cc71f985d36559355d00bfbefdb6b92718ca47a594760f

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
