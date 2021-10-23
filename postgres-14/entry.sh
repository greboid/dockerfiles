#!/bin/sh
initdb -A trust $PGDATA

if [ -z "$@" ]; then
  exec postgres -c config_file="/etc/postgresql.conf"
else
  exec $@
fi