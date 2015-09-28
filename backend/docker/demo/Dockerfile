FROM ubuntu:trusty

# Set locale
RUN locale-gen en_US en_US.UTF-8 && \
    update-locale LANG=en_US.UTF-8 && \
    dpkg-reconfigure locales

# Set Go, HOMEDIR and locale
ENV DEBIAN_FRONTEND=noninteractive \
	GO_VER=1.5.1 \
    GODIR=/go1.5 \
    GOROOT=/go1.5/go \
    GOPATH=/go \
    PGSQL_VER=9.3 \
    COREROLLER_DIR=/coreroller \
    LANGUAGE=en_US.UTF-8 \
    LANG=en_US.UTF-8

# Install basic stuff
RUN apt-get update && \
    apt-get dist-upgrade -y
RUN apt-get install -y \
        wget \
        git \
        unzip \
        build-essential

RUN apt-get install -y nodejs-legacy npm

# Install Go
RUN mkdir -p $GODIR && \
	wget -qO- http://golang.org/dl/go${GO_VER}.linux-amd64.tar.gz | tar -C $GODIR -xzf -
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Install gb tool
RUN go get github.com/constabulary/gb/...

# Install coreroller repo
RUN mkdir $COREROLLER_DIR
RUN git clone https://github.com/coreroller/coreroller.git $COREROLLER_DIR && \
 	cd $COREROLLER_DIR/backend/ && gb build cmd/rollerd && gb build cmd/initdb

# Build coreroller UI
RUN cd $COREROLLER_DIR/frontend && \
	npm install && \
	npm run build

# Install apt key for Postgresql
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys B97B0AFCAA1A47F044F244A07FCC7D46ACCC4CF8
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ trusty-pgdg main" > /etc/apt/sources.list.d/pgdg.list
RUN apt-get update && \
	apt-get -y -q install python-software-properties software-properties-common && \
    apt-get -y -q install postgresql-$PGSQL_VER postgresql-client-$PGSQL_VER postgresql-contrib-$PGSQL_VER postgresql-server-dev-$PGSQL_VER && \
	# Install semver 
	wget --quiet --no-check-certificate -O pg-semver-master.zip https://github.com/theory/pg-semver/archive/master.zip && \
	unzip pg-semver-master.zip && \
	cd pg-semver-master && \
	make && \
	make install && \
	cd - && \
	rm -rf pg-semver-master pg-semver-master.zip && \
    # Cleanup
    apt-get remove -y wget unzip build-essential python-software-properties software-properties-common && \
    apt-get autoremove -y && apt-get clean

# Add scripts
ADD scripts /scripts
RUN chmod +x /scripts/*.sh
RUN touch /.firstrun

# Command to run
ENV COREROLLER_DB_URL="postgres://postgres@127.0.0.1:5432/coreroller?sslmode=disable"
WORKDIR /coreroller/backend
ENTRYPOINT ["/scripts/run.sh"]
