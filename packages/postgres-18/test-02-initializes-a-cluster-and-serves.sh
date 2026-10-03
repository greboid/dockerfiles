#!/bin/sh
# postgres refuses to run as root: drop to the guest's build user.
# Socket-only, under /tmp/pgtest so nothing outside is touched.
mkdir -p /tmp/pgtest && chown build:build /tmp/pgtest
install -m 0644 check.sql /tmp/pgtest/check.sql
chown build:build /tmp/pgtest/check.sql
su build -s /bin/sh -c 'export PATH=/usr/bin:$PATH; \
  initdb -D /tmp/pgtest/data -A trust --no-locale \
    -c unix_socket_directories=/tmp/pgtest'
su build -s /bin/sh -c 'export PATH=/usr/bin:$PATH; \
  pg_ctl -D /tmp/pgtest/data -w -o "-c unix_socket_directories=/tmp/pgtest" start'
su build -s /bin/sh -c '/usr/bin/psql -h /tmp/pgtest -d postgres \
  -f /tmp/pgtest/check.sql' | grep -q 18.6
su build -s /bin/sh -c '/usr/bin/pg_ctl -D /tmp/pgtest/data -m fast -w stop'
