#!/bin/sh
# OPTFLAGS="" drops upstream's -march=native so the package runs on
# any host. PGXS honors DESTDIR, so the extension stages into this
# package's own tree alongside the server's /usr layout.
make -j"$(nproc)" OPTFLAGS="" PG_CONFIG=/usr/bin/pg_config
make install OPTFLAGS="" PG_CONFIG=/usr/bin/pg_config \
  DESTDIR="${DESTDIR}"
