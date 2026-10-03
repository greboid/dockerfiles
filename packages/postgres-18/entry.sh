#!/usr/bin/env bash
set -Eeuo pipefail

# --------------------------------------------------------------------------
# helpers
# --------------------------------------------------------------------------

is_root() { [[ $(id -u) == 0 ]]; }

# --------------------------------------------------------------------------
# setup steps
# --------------------------------------------------------------------------

# create the directories postgres needs; tolerate chmod/chown failures the
# same way upstream does (weird volume setups may still work)
create_directories() {
    mkdir -p "$PGDATA"
    chmod 700 "$PGDATA" || :
    mkdir -p /var/run/postgresql || :
    chmod 775 /var/run/postgresql || :
    if [[ -n "${POSTGRES_INITDB_WALDIR:-}" ]]; then
        mkdir -p "$POSTGRES_INITDB_WALDIR"
        chmod 700 "$POSTGRES_INITDB_WALDIR"
    fi
    if is_root; then
        # allow the container to be started with `--user`
        find "$PGDATA" \! -user postgres -exec chown postgres '{}' +
        find /var/run/postgresql \! -user postgres -exec chown postgres '{}' +
        if [[ -n "${POSTGRES_INITDB_WALDIR:-}" ]]; then
            find "$POSTGRES_INITDB_WALDIR" \! -user postgres -exec chown postgres '{}' +
        fi
    fi
}

# when started as root, re-exec this script as the postgres user and return
# to the top (whatever privilege-drop tool the host provides)
demote_and_reexec() {
    if command -v su-exec >/dev/null 2>&1; then
        exec su-exec postgres "$@"
    elif command -v gosu >/dev/null 2>&1; then
        exec gosu postgres "$@"
    elif command -v setpriv >/dev/null 2>&1; then
        exec setpriv --reuid=postgres --regid=postgres --init-groups "$@"
    elif command -v runuser >/dev/null 2>&1; then
        exec runuser -u postgres -- "$@"
    else
        echo >&2 "error: started as root but found no su-exec/gosu/setpriv/runuser to drop privileges"
        exit 1
    fi
}

# refuse to init a cluster we could not secure; warn on misconfigurations
check_minimum_env() {
    if [[ ${#POSTGRES_PASSWORD} -ge 100 ]]; then
        cat >&2 <<-'EOWARN'

            WARNING: The supplied POSTGRES_PASSWORD is 100+ characters.

              This will not work if used via PGPASSWORD with "psql".

        EOWARN
    fi
    if [[ -z "$POSTGRES_PASSWORD" && "$POSTGRES_HOST_AUTH_METHOD" != 'trust' ]]; then
        cat >&2 <<-'EOE'
            Error: Database is uninitialized and superuser password is not specified.
                   You must specify POSTGRES_PASSWORD to a non-empty value for the
                   superuser. For example, "-e POSTGRES_PASSWORD=password" on "docker run".

                   You may also use "POSTGRES_HOST_AUTH_METHOD=trust" to allow all
                   connections without a password. This is *not* recommended.
        EOE
        exit 1
    fi
    if [[ "$POSTGRES_HOST_AUTH_METHOD" == 'trust' ]]; then
        echo >&2 'WARNING: POSTGRES_HOST_AUTH_METHOD=trust allows unauthenticated access; not recommended.'
    fi
}

# run initdb on an empty PGDATA, creating the POSTGRES_USER superuser with
# POSTGRES_PASSWORD; extra arguments flow to initdb
initialize_cluster() {
    # hand the password to initdb via a file instead of the command line
    local pwfile
    pwfile="$(mktemp)"
    trap 'rm -f "$pwfile"' RETURN
    printf '%s\n' "$POSTGRES_PASSWORD" > "$pwfile"

    local init_args=( --username="$POSTGRES_USER" --pwfile="$pwfile" )
    if [[ -n "${POSTGRES_INITDB_WALDIR:-}" ]]; then
        init_args+=( --waldir "$POSTGRES_INITDB_WALDIR" )
    fi
    local user_args=()
    if [[ -n "${POSTGRES_INITDB_ARGS:-}" ]]; then
        # evaluated so users may pass quoted multi-word options
        eval "user_args=( $POSTGRES_INITDB_ARGS )"
    fi

    initdb "${init_args[@]}" "${user_args[@]}" "$@"
}

# append the "host" auth line to pg_hba.conf; arguments (minus a leading
# `postgres`) are passed to `postgres -C` to learn the configured password
# encryption so the auth method matches it
append_hba_host_line() {
    if [[ "${1:-}" == 'postgres' ]]; then
        shift
    fi
    local auth_method
    auth_method="$(postgres -C password_encryption "$@")"
    : "${POSTGRES_HOST_AUTH_METHOD:=$auth_method}"
    {
        echo
        if [[ "$POSTGRES_HOST_AUTH_METHOD" == 'trust' ]]; then
            echo '# warning trust is enabled for all connections'
        fi
        echo "host all all all $POSTGRES_HOST_AUTH_METHOD"
    } >> "$PGDATA/pg_hba.conf"
}

# start a socket-only server so setup can talk to the cluster with psql;
# arguments (minus a leading `postgres`) are forwarded to the server
temp_server_start() {
    if [[ "${1:-}" == 'postgres' ]]; then
        shift
    fi
    local server_args=( "$@" -c listen_addresses= -p "${PGPORT:-5432}" )
    local opts
    printf -v opts '%q ' "${server_args[@]}"
    PGUSER="${PGUSER:-$POSTGRES_USER}" pg_ctl -D "$PGDATA" -o "$opts" -w start
}

temp_server_stop() {
    PGUSER="${PGUSER:-postgres}" pg_ctl -D "$PGDATA" -m fast -w stop
}

# run psql against the temporary socket-only server; TCP env vars are
# cleared because the temporary server does not listen on TCP
docker_process_sql() {
    local psql_args=( -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --no-password --no-psqlrc )
    if [[ -n "${POSTGRES_DB:-}" ]]; then
        psql_args+=( --dbname "$POSTGRES_DB" )
    fi
    PGHOST= PGHOSTADDR= psql "${psql_args[@]}" "$@"
}

# create POSTGRES_DB unless it already exists
create_initial_database() {
    local exists
    exists="$(
        POSTGRES_DB= docker_process_sql --dbname postgres --set db="$POSTGRES_DB" --tuples-only <<-'EOSQL'
            SELECT 1 FROM pg_database WHERE datname = :'db' ;
        EOSQL
    )"
    if [[ -z "$exists" ]]; then
        POSTGRES_DB= docker_process_sql --dbname postgres <<-'EOSQL'
            CREATE DATABASE :"db" ;
        EOSQL
        echo
    fi
}

# load the POSTGRES_* environment and detect an already-initialized cluster
setup_env() {
    POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"
    POSTGRES_USER="${POSTGRES_USER:-postgres}"
    POSTGRES_DB="${POSTGRES_DB:-$POSTGRES_USER}"
    : "${POSTGRES_HOST_AUTH_METHOD:=}"

    DATABASE_ALREADY_EXISTS=''
    if [[ -s "$PGDATA/PG_VERSION" ]]; then
        DATABASE_ALREADY_EXISTS='true'
    fi
}

# --------------------------------------------------------------------------
# main
# --------------------------------------------------------------------------

# true when the arguments contain an option that makes postgres exit instead
# of serving (-C is deliberately excluded: it always needs a config file)
wants_help() {
    local arg
    for arg in "$@"; do
        case "$arg" in
            -'?'|--help|--describe-config|-V|--version) return 0 ;;
        esac
    done
    return 1
}

main() {
    # empty invocation still means "run the server"
    if (( $# == 0 )); then
        set -- postgres
    fi

    # a leading flag implies the postgres server
    if [[ "${1:0:1}" == '-' ]]; then
        set -- postgres "$@"
    fi

    if [[ "$1" == 'postgres' ]] && ! wants_help "$@"; then
        setup_env
        create_directories
        if is_root; then
            # restart as the unprivileged postgres user
            demote_and_reexec "$BASH_SOURCE" "$@"
        fi

        # only initialize on an empty data directory
        if [[ -z "$DATABASE_ALREADY_EXISTS" ]]; then
            check_minimum_env

            initialize_cluster
            append_hba_host_line "$@"

            # socket auth may still require a password when initdb was
            # told to use e.g. --auth-local=md5; harmless otherwise
            export PGPASSWORD="${PGPASSWORD:-$POSTGRES_PASSWORD}"
            temp_server_start "$@"

            create_initial_database

            temp_server_stop
            unset PGPASSWORD

            echo
            echo 'PostgreSQL init process complete; ready for start up.'
            echo
        else
            echo
            echo 'PostgreSQL Database directory appears to contain a database; Skipping initialization'
            echo
        fi
    fi

    exec "$@"
}

main "$@"
