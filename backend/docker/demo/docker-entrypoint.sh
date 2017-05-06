#!/bin/sh

if [ ! -f "$PGDATA/postgresql.conf" ]; then
    mkdir -p $PGDATA
    chown -R postgres:postgres "$PGDATA"

    gosu postgres initdb
    gosu postgres pg_ctl -w start
    gosu postgres psql -c "CREATE DATABASE coreroller;"
    gosu postgres psql -c "ALTER DATABASE coreroller SET TIMEZONE = 'UTC';"
    gosu postgres psql -d coreroller -c "CREATE EXTENSION \"uuid-ossp\";"
else
    gosu postgres pg_ctl -w start
fi

LOGXI=*=WRN LOGXI_FORMAT=happy exec /coreroller/rollerd -http-log -http-static-dir=/coreroller/static -enable-syncer=false
