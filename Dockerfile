FROM golang:alpine as builder

RUN apk --no-cache add bash make openssl git \
    && mkdir -p /go/src/github.com/phires \
    && cd /go/src/github.com/phires \
    && git clone https://github.com/phires/prometheus-certcheck \
    && cd prometheus-certcheck \
    && go get \
    && go build

FROM alpine:latest

ENV INTERVAL 60
ENV ADDRESS :8080
ENV METRICSPATH /metrics

RUN mkdir -p /conf
COPY --from=builder /go/src/github.com/phires/prometheus-certcheck/prometheus-certcheck /prometheus-certcheck
COPY entrypoint.sh /entrypoint.sh

EXPOSE 8080
ENTRYPOINT [ "/entrypoint.sh" ]