FROM golang:1.23.4-alpine3.20 AS builder
WORKDIR /go/src/source-websocket
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM alpine:3.20
RUN apk --no-cache add ca-certificates \
    && update-ca-certificates
COPY --from=builder /go/src/source-websocket/source-websocket /bin/source-websocket
ENTRYPOINT ["/bin/source-websocket"]
