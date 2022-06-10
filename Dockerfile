# defaults which may be overridden from the build command
ARG GO_VERSION=1.18
ARG ALPINE_VERSION=3.16

# build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

COPY . /go/src/github.com/jabbors/sparsnas_mqtt_consumer
WORKDIR /go/src/github.com/jabbors/sparsnas_mqtt_consumer
ARG APP_VERSION=0.0
RUN go install -ldflags="-X \"main.version=${APP_VERSION}\""

# final stage
FROM alpine:${ALPINE_VERSION}

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=builder /go/bin/sparsnas_mqtt_consumer /usr/bin/sparsnas_mqtt_consumer
USER nobody:nobody
ENTRYPOINT [ "/usr/bin/sparsnas_mqtt_consumer" ]
