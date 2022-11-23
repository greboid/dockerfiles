#!/bin/sh
if [ -n "$POSTGRES_USER" ]; then
  PGUSER=${POSTGRES_USER}
fi
if [ -n "$POSTGRES_PASSWORD" ]; then
  PGPASSWORD=${POSTGRES_PASSWORD}
fi
if [ -z "$PGDATA" ]; then
  PGDATA="/var/lib/postgres/data"
fi
if [ -z "$PGUSER" ]; then
  PGUSER="postgres"
fi
if [ -z "$PGPASSWORD" ]; then
  PGPASSWORD="password"
fi
if [ -z "$POSTGRES_DB" ]; then
  POSTGRES_DB=$PGUSER
fi
if [ -n "$POSTGRES_HOST_AUTH_METHOD" ]; then
  POSTGRES_HOST_AUTH_METHOD="scram-sha-256"
fi
if [ ! -s "$PGDATA" ]; then
        initdb -U postgres --pwfile=<(echo "$PGPASSWORD") -D $PGDATA --no-instructions --auth-host ${POSTGRES_HOST_AUTH_METHOD} --auth-local trust
        echo "local all all trust" > "$PGDATA/pg_hba.conf"
        echo "host all all 127.0.0.1/32 trust" >> "$PGDATA/pg_hba.conf"
        echo "host all all ::1/128 trust" >> "$PGDATA/pg_hba.conf"
        echo "host all all all ${POSTGRES_HOST_AUTH_METHOD}" >> "$PGDATA/pg_hba.conf"
	pg_ctl -D "$PGDATA" -o "-c listen_addresses='127.0.0.1'" -o "-p 5432" -o "-c unix_socket_directories='/var/run/psql'" -w start
	echo "CREATE USER ${PGUSER} WITH PASSWORD '${PGPASSWORD}';CREATE DATABASE ${POSTGRES_DB} OWNER '${PGUSER}';" | psql -U postgres
	pg_ctl -D "$PGDATA" -m fast -w stop

else
	echo "Data directory exists, not editing"
fi

if [ -z "$@" ]; then
  exec postgres -c config_file="/etc/postgresql.conf"
else
  exec $@
fi
