FROM golang:alpine as builder

RUN set -x \
    && apk update \
    && apk add --virtual .build-deps git curl go nodejs musl-dev \
    && apk add ca-certificates postgresql postgresql-contrib tzdata \
    && mkdir -p /build /coreroller/static \
    && git clone --branch v0.4.4 https://github.com/constabulary/gb /go/src/github.com/constabulary/gb \
    && GOPATH=/go go get github.com/constabulary/gb/... \
    && cd /build \
    && git clone https://github.com/coreroller/coreroller \
    && cd /build/coreroller/backend \
    && /go/bin/gb build \
    && cp /build/coreroller/backend/bin/rollerd /coreroller \
    && cd /build/coreroller/frontend \
    && npm install \
    && npm run build 

FROM alpine:3.7

RUN apk --no-cache add ca-certificates tzdata curl \
    && apk --no-cache add postgresql postgresql-contrib \
    && mkdir -p /coreroller/static \
    && mkdir -p /run/postgresql \
    && chown -R postgres:postgres /run/postgresql \
    && curl -o /usr/local/bin/gosu -sSL "https://github.com/tianon/gosu/releases/download/1.10/gosu-amd64" \
    && chmod +x /usr/local/bin/gosu

COPY --from=builder /build/coreroller/backend/bin/rollerd /coreroller/
COPY --from=builder /build/coreroller/frontend/built/* /coreroller/static/
ADD docker-entrypoint.sh /
ENV PGDATA /coreroller/data
EXPOSE 8000
CMD ["/docker-entrypoint.sh"]
