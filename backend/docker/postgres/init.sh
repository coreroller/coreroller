/usr/lib/postgresql/9.4/bin/psql -U postgres -c "create database coreroller;"
/usr/lib/postgresql/9.4/bin/psql -U postgres -d coreroller -c "create extension \"uuid-ossp\";"
/usr/lib/postgresql/9.4/bin/psql -U postgres -d coreroller -c "create extension semver;"

/usr/lib/postgresql/9.4/bin/psql -U postgres -c "create database coreroller_tests;"
/usr/lib/postgresql/9.4/bin/psql -U postgres -d coreroller_tests -c "create extension \"uuid-ossp\";"
/usr/lib/postgresql/9.4/bin/psql -U postgres -d coreroller_tests -c "create extension semver;"
