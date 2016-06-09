FROM alpine:3.3

RUN set -x \
    && apk update \
    && apk add --virtual .build-deps git go nodejs \
    && apk add ca-certificates \
    && mkdir -p /build /coreroller/static \
    && GOPATH=/go go get github.com/constabulary/gb/... \
    && cd /build \
    && git clone https://github.com/coreroller/coreroller \
    && cd /build/coreroller/backend \
    && /go/bin/gb build \
    && cp /build/coreroller/backend/bin/rollerd /coreroller \
    && cd /build/coreroller/frontend \
    && npm install \
    && npm run build \
    && cp -R built/* /coreroller/static \
    && apk del .build-deps \
    && rm -rf /var/cache/apk /go /build /root/.npm /tmp/npm*

ENV COREROLLER_DB_URL "postgres://postgres@postgresqld.local:5432/coreroller?sslmode=disable"
EXPOSE 8000
CMD ["/coreroller/rollerd", "-http-static-dir=/coreroller/static"]
