#!/bin/bash

# Start PostgreSQL service
sudo -u postgres /usr/lib/postgresql/$PGSQL_VER/bin/postgres -D /data &

while ! sudo -u postgres psql -q -c "select true;"; do sleep 1; done

# Create dabatase
echo "Creating databases..."
sudo -u postgres psql -q <<-EOF
CREATE DATABASE coreroller;
CREATE DATABASE coreroller_tests;
EOF

echo "Installing extensions..."
sudo -u postgres psql -d coreroller -q <<-EOF
CREATE EXTENSION "uuid-ossp";
CREATE EXTENSION semver;
EOF
sudo -u postgres psql -d coreroller_tests -q <<-EOF
CREATE EXTENSION "uuid-ossp";
CREATE EXTENSION semver;
EOF

echo "Installing the databse with the initial schema..."
$COREROLLER_DIR/backend/bin/initdb

rm -f /.firstrun