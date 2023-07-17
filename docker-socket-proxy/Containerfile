# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Containerfile.gotpl
# BOM: {"image:baseroot":"0ff388c461add6a09dde64f3bfe3cb62e3ca09f1fb4b69ca93a2f0aa0ae68137","image:haproxy":"b8b6d5e89175b095b8233b6517f6c0d30ff4f386c3b5b4081c5e3f894470bf34"}

FROM reg.g5d.dev/haproxy@sha256:b8b6d5e89175b095b8233b6517f6c0d30ff4f386c3b5b4081c5e3f894470bf34 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:0ff388c461add6a09dde64f3bfe3cb62e3ca09f1fb4b69ca93a2f0aa0ae68137

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
