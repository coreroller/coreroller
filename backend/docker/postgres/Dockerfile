FROM alpine:3.8

RUN set -x \
    && apk add --no-cache --virtual .build-deps curl \
    && apk add --no-cache postgresql postgresql-contrib \
    && curl -o /usr/local/bin/gosu -sSL "https://github.com/tianon/gosu/releases/download/1.10/gosu-amd64" \
    && chmod +x /usr/local/bin/gosu \
    && mkdir -p /run/postgresql/ \
    && chown -R postgres:postgres /run/postgresql \
    && apk del .build-deps

ADD docker-entrypoint.sh /

ENV PGDATA /coreroller/data
EXPOSE 5432
CMD ["/docker-entrypoint.sh"]
