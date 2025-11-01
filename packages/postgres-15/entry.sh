#!/bin/sh
set -e

if [ ! -s "$PGDATA/PG_VERSION" ]; then
    echo "Initializing database..."
    : "${POSTGRES_USER:=postgres}"
    : "${POSTGRES_DB:=$POSTGRES_USER}"
    : "${POSTGRES_HOST_AUTH_METHOD:=scram-sha-256}"

    if [ -z "$POSTGRES_PASSWORD" ] && [ "$POSTGRES_HOST_AUTH_METHOD" != "trust" ]; then
        echo >&2 "Error: Database is uninitialized and POSTGRES_PASSWORD is not set."
        echo >&2 "Set POSTGRES_PASSWORD or use POSTGRES_HOST_AUTH_METHOD=trust"
        exit 1
    fi

    mkdir -p "$PGDATA"
    chmod 700 "$PGDATA"

    # Use a temporary file for the password since sh doesn't support process substitution
    PWFILE=$(mktemp)
    printf "%s\n" "$POSTGRES_PASSWORD" > "$PWFILE"
    initdb --username="$POSTGRES_USER" --pwfile="$PWFILE" \
        --auth="$POSTGRES_HOST_AUTH_METHOD" --auth-host="$POSTGRES_HOST_AUTH_METHOD"
    rm -f "$PWFILE"

    echo "host all all all $POSTGRES_HOST_AUTH_METHOD" >> "$PGDATA/pg_hba.conf"
    echo "PostgreSQL init process complete"
else
    echo "Database already initialized, skipping initialization"
fi

exec postgres "$@"
