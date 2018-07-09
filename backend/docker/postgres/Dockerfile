FROM alpine:3.7

RUN set -x \
    && apk update \
    && apk add --virtual .build-deps curl \
    && apk add postgresql postgresql-contrib \
    && curl -o /usr/local/bin/gosu -sSL "https://github.com/tianon/gosu/releases/download/1.10/gosu-amd64" \
    && chmod +x /usr/local/bin/gosu \
    && mkdir -p /run/postgresql/ \
    && chown -R postgres:postgres /run/postgresql \
    && apk del .build-deps \
    && rm -rf /var/cache/apk

ADD docker-entrypoint.sh /

ENV PGDATA /coreroller/data
EXPOSE 5432
CMD ["/docker-entrypoint.sh"]
