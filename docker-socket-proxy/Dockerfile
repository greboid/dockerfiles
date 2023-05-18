# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"c0e2307d35f12546a86a703c9366d018619d61dae48cd16ebc6c10a51d3d7d91","image:haproxy":"b3a658f3e3d402d21660b3a78abdedd431fc67fd95a3e138a1221398145caf3c"}

FROM reg.g5d.dev/haproxy@sha256:b3a658f3e3d402d21660b3a78abdedd431fc67fd95a3e138a1221398145caf3c AS haproxy

FROM reg.g5d.dev/baseroot@sha256:c0e2307d35f12546a86a703c9366d018619d61dae48cd16ebc6c10a51d3d7d91

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
