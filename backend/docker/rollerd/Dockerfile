FROM golang:alpine as builder

RUN set -x \
    && apk update \
    && apk add --virtual .build-deps git go nodejs musl-dev\
    && apk add ca-certificates tzdata \
    && mkdir -p /build /coreroller/static \
    && git clone --branch v0.4.4 https://github.com/constabulary/gb /go/src/github.com/constabulary/gb \
    && GOPATH=/go go get github.com/constabulary/gb/... \
    && cd /build \
    && git clone https://github.com/coreroller/coreroller \
    && cd /build/coreroller/backend \
    && /go/bin/gb build \
    && cd /build/coreroller/frontend \
    && npm install \
    && npm run build 

FROM alpine:3.7
RUN apk --no-cache add ca-certificates tzdata \
    && mkdir -p /coreroller/static
COPY --from=builder /build/coreroller/backend/bin/rollerd /coreroller/
COPY --from=builder /build/coreroller/frontend/built/* /coreroller/static/
ENV COREROLLER_DB_URL "postgres://postgres@postgresqld.local:5432/coreroller?sslmode=disable"
EXPOSE 8000
CMD ["/coreroller/rollerd", "-http-static-dir=/coreroller/static"]
