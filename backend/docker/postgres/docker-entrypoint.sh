#!/bin/sh

if [ ! -f "$PGDATA/postgresql.conf" ]; then
    mkdir -p $PGDATA
    chown -R postgres:postgres "$PGDATA"

    gosu postgres initdb
    sed -ri "s/^#(listen_addresses\s*=\s*)\S+/\1'*'/" "$PGDATA"/postgresql.conf

    # WARNING! This is not secure as it allows anyone to connect to your database
    # remotely without asking for any password. This is just an example, please
    # consider adding a user to the database, use md5 instead of trust and, if 
    # needed, restrict the network range that should be allowed to connect.
    { echo; echo "host all all 0.0.0.0/0 trust"; } >> "$PGDATA/pg_hba.conf"

    gosu postgres pg_ctl -w start

    gosu postgres psql -c "CREATE DATABASE coreroller;"
    gosu postgres psql -c "ALTER DATABASE coreroller SET TIMEZONE = 'UTC';"
    gosu postgres psql -d coreroller -c "CREATE EXTENSION \"uuid-ossp\";"

    gosu postgres psql -c "CREATE DATABASE coreroller_tests;"
    gosu postgres psql -c "ALTER DATABASE coreroller_tests SET TIMEZONE = 'UTC';"
    gosu postgres psql -d coreroller_tests -c "CREATE EXTENSION \"uuid-ossp\";"

    gosu postgres pg_ctl -m fast -w stop
fi

exec gosu postgres postgres 
