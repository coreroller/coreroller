#!/bin/bash

# Start PostgreSQL service
sudo -u postgres /usr/lib/postgresql/9.3/bin/pg_ctl -D /data -l /tmp/logfile start
sleep 3

# Create dabatase
echo "Creating databases..."
sudo -u postgres psql -q <<-EOF
CREATE DATABASE coreroller;
ALTER DATABASE coreroller SET TIMEZONE = 'UTC';
CREATE DATABASE coreroller_tests;
ALTER DATABASE coreroller_tests SET TIMEZONE = 'UTC';
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

rm -f /.firstrun
