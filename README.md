# source-websocket
Server-sent events source type

```shell
grpcurl \
  -plaintext \
  -proto api/grpc/service.proto \
  -d @ \
  localhost:50051 \
  awakari.source.websocket.Service/Create
```

```json
{
  "url": "wss://www.seismicportal.eu/standing_order/websocket",
  "groupId": "default"
}
```
