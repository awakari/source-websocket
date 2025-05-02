FROM golang:1.24.2-alpine3.21 AS builder
WORKDIR /go/src/source-websocket
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM alpine:3.21.3
RUN apk --no-cache add ca-certificates \
    && update-ca-certificates
COPY --from=builder /go/src/source-websocket/source-websocket /bin/source-websocket
ENTRYPOINT ["/bin/source-websocket"]
