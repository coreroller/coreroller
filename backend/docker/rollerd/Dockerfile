FROM golang:1.5

# Set Go, HOMEDIR and locale
ENV DEBIAN_FRONTEND=noninteractive \
    COREROLLER_DIR=/coreroller \
    LANGUAGE=en_US.UTF-8 \
    LANG=en_US.UTF-8 \
    POSTGRESQLD_PORT_5432_TCP_ADDR=172.17.8.101 \
    POSTGRESQLD_PORT_5432_TCP_PORT=5432

# Install basic stuff
RUN apt-get update && \
    apt-get dist-upgrade -y && \
	apt-get install -y git nodejs-legacy npm

# Install gb tool
RUN go get github.com/constabulary/gb/...

# Install CoreRoller
RUN mkdir $COREROLLER_DIR
RUN git clone https://github.com/coreroller/coreroller.git $COREROLLER_DIR && \
 	cd $COREROLLER_DIR/backend/ && gb build cmd/rollerd && gb build cmd/initdb

# Build coreroller UI
RUN cd $COREROLLER_DIR/frontend && \
	npm install && \
	npm run build

# We could uninstall nodejs/npm at this stage to save some on image size, but at the moment
# we make it easy for development by leaving npm so you can rebuild the UI if needed

ENV COREROLLER_DB_URL="postgres://postgres@$POSTGRESQLD_PORT_5432_TCP_ADDR:$POSTGRESQLD_PORT_5432_TCP_PORT/coreroller?sslmode=disable"
WORKDIR $COREROLLER_DIR/backend
CMD ["bin/rollerd"]
