# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"fd51da47fdc6d0f778cc71f985d36559355d00bfbefdb6b92718ca47a594760f","image:haproxy":"e621533d3d3d4f0c6a556357435aad9ea81d19860d528778e2846e07a6ce9e4e"}

FROM reg.g5d.dev/haproxy@sha256:e621533d3d3d4f0c6a556357435aad9ea81d19860d528778e2846e07a6ce9e4e AS haproxy

FROM reg.g5d.dev/baseroot@sha256:fd51da47fdc6d0f778cc71f985d36559355d00bfbefdb6b92718ca47a594760f

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
