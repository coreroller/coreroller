FROM alpine:3.3

RUN set -x \
    && apk update \
    && apk add --virtual .build-deps git curl postgresql-dev gcc make musl-dev \
    && apk add postgresql postgresql-contrib \
    && curl -o /usr/local/bin/gosu -sSL "https://github.com/tianon/gosu/releases/download/1.9/gosu-amd64" \
    && chmod +x /usr/local/bin/gosu \
    && cd /tmp \
    && git clone https://github.com/theory/pg-semver \
    && cd pg-semver \
    && git checkout v0.12.0 \
    && make \
    && make install \
    && apk del .build-deps \
    && rm -rf /var/cache/apk /tmp/pg-semver

ADD docker-entrypoint.sh /

ENV PGDATA /coreroller/data
EXPOSE 5432
CMD ["/docker-entrypoint.sh"]
