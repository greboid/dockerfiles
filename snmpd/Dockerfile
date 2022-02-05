# Generated from https://github.com/greboid/dockerfiles/blob/master/snmpd/Dockerfile.gotpl
# BOM: {"image:alpine":"ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1","image:base":"e390636a8773e2b84066cdbde990e80b3c97b8e7e25428df9f4b8f097323f861"}

FROM reg.g5d.dev/alpine@sha256:ef95559f475cdf56e55db71c168f1616d497e2215bbd9f6bc63398036ae546a1 as build

RUN set -eux; \
  apk add --update --no-cache git build-base linux-headers libnl3-dev && \
  git clone https://github.com/net-snmp/net-snmp /src/snmp; \
  cd /src/snmp; \
  find . -type f -print0 | xargs -0 sed -i 's/\"\/proc/\"\/host_proc/g'

RUN cd /src/snmp; \ 
  ./configure --with-defaults --build x86_64-pc-linux \ 
              --disable-embedded-perl && \
  make standardall && \
  make install;

FROM reg.g5d.dev/base@sha256:e390636a8773e2b84066cdbde990e80b3c97b8e7e25428df9f4b8f097323f861

EXPOSE 1161 1161/udp
COPY snmpd.conf /etc/snmp
CMD [ "/usr/local/sbin/snmpd", "-f", "-c", "/etc/snmp/snmpd.conf" ]
