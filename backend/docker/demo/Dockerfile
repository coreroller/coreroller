FROM alpine:3.3

RUN set -x \
    && apk update \
    && apk add --virtual .build-deps git curl go nodejs \
    && apk add ca-certificates postgresql postgresql-contrib tzdata \
    && mkdir -p /build /coreroller/static \
    && curl -o /usr/local/bin/gosu -sSL "https://github.com/tianon/gosu/releases/download/1.9/gosu-amd64" \
    && chmod +x /usr/local/bin/gosu \
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

ADD docker-entrypoint.sh /

ENV PGDATA /coreroller/data
EXPOSE 8000
CMD ["/docker-entrypoint.sh"]
