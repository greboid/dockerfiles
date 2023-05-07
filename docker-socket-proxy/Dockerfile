# Generated from https://github.com/greboid/dockerfiles/blob/master/docker-socket-proxy/Dockerfile.gotpl
# BOM: {"image:baseroot":"fc134ce385569747e9c93e1f1e117040255d231e6173afc7ef2d3b176063e710","image:haproxy":"5adb587703e80bc29e2734bcabb89021d7b8d1cbdacd5e7f5db510a73825fd97"}

FROM reg.g5d.dev/haproxy@sha256:5adb587703e80bc29e2734bcabb89021d7b8d1cbdacd5e7f5db510a73825fd97 AS haproxy

FROM reg.g5d.dev/baseroot@sha256:fc134ce385569747e9c93e1f1e117040255d231e6173afc7ef2d3b176063e710

STOPSIGNAL SIGUSR1

COPY --from=haproxy /haproxy /haproxy
COPY haproxy.cfg haproxy.cfg.license /usr/local/etc/haproxy/

ENTRYPOINT ["/haproxy", "-W", "-db"]
CMD ["-f", "/usr/local/etc/haproxy/haproxy.cfg"]
