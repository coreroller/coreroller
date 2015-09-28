FROM postgres:9.4

# adds the following extensions on top of the official image
#   * semver

WORKDIR /tmp

RUN apt-get -y update && \
    apt-get install -y wget \
    	unzip  \
        build-essential postgresql-server-dev-$PG_MAJOR && \

	# Install semver 
	wget --quiet --no-check-certificate -O pg-semver-master.zip https://github.com/theory/pg-semver/archive/master.zip && \
	unzip pg-semver-master.zip && \
	cd pg-semver-master && \
	make && \
	make install && \
	cd - && \
	rm -rf pg-semver-master pg-semver-master.zip && \

    # Cleanup
    apt-get remove -y wget unzip build-essential && \
    apt-get autoremove -y && apt-get clean

# Load custom commands
ADD init.sh /docker-entrypoint-initdb.d/init.sh

# Adjust PostgreSQL configuration so that remote connections are possible. 
RUN echo "host  all  all  0.0.0.0/0  trust" >> /var/lib/postgresql/data/pg_hba.conf
RUN echo "listen_addresses = '*'" >> /var/lib/postgresql/data/postgresql.conf

# Mount share volume (after we change the config)
VOLUME ["/var/lib/postgresql/data"]
