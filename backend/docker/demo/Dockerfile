FROM golang:alpine as builder

ENV GOPATH /go
ARG GIT_REMOTE_URL=https://github.com/coreroller/coreroller
ARG GIT_BRANCH=master
WORKDIR /build

RUN set -x \
    && apk --no-cache add git curl \
    ca-certificates postgresql postgresql-contrib tzdata \
    && mkdir -p /build /coreroller/static \
    # Switch to https://pkgs.alpinelinux.org/package/edge/testing/x86_64/gb once this is in stable
    && git clone --branch v0.4.4 https://github.com/constabulary/gb /go/src/github.com/constabulary/gb \
    && go get github.com/constabulary/gb/... \
    && git clone --single-branch --branch ${GIT_BRANCH} ${GIT_REMOTE_URL} \
    && cd coreroller/backend \
    && /go/bin/gb build \
    && cp /build/coreroller/backend/bin/rollerd /coreroller

FROM node:9-alpine as assets
WORKDIR /build
ARG GIT_REMOTE_URL=https://github.com/coreroller/coreroller
ARG GIT_BRANCH=master

RUN apk --no-cache add git \
    && git clone --single-branch --branch ${GIT_BRANCH} ${GIT_REMOTE_URL} \
    && cd /build/coreroller/frontend \
    && npm install \
    && npm run build

FROM alpine:3.8
ENV PGDATA /coreroller/data

RUN apk --no-cache add ca-certificates tzdata curl \
    postgresql postgresql-contrib \
    && mkdir -p /coreroller/static /run/postgresql \
    && chown -R postgres:postgres /run/postgresql \
    && curl -o /usr/local/bin/gosu -sSL "https://github.com/tianon/gosu/releases/download/1.10/gosu-amd64" \
    && chmod +x /usr/local/bin/gosu

COPY --from=builder /build/coreroller/backend/bin/rollerd /coreroller/
COPY --from=assets /build/coreroller/frontend/built/ /coreroller/static/

ADD docker-entrypoint.sh /
EXPOSE 8000
CMD ["/docker-entrypoint.sh"]
