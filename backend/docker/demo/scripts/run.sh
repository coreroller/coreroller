#!/bin/bash

# Initialize data directory
DATA_DIR=/data
if [ ! -f $DATA_DIR/postgresql.conf ]; then
    mkdir -p $DATA_DIR
    chown postgres:postgres /data

    sudo -u postgres /usr/lib/postgresql/$PGSQL_VER/bin/initdb -E utf8 --locale en_US.UTF-8 -D $DATA_DIR
    mkdir -p $DATA_DIR/pg_log
fi
chown -R postgres:postgres /data
chmod -R 700 /data

# Initialize first run
if [[ -e /.firstrun ]]; then
    /scripts/first_run.sh
fi

# Start rollerd
cd $COREROLLER_DIR/backend && LOGXI=* LOGXI_FORMAT=text $COREROLLER_DIR/backend/bin/rollerd
