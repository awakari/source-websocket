package handler

import (
	"context"
	"github.com/awakari/source-websocket/api/grpc/events"
	"github.com/awakari/source-websocket/model"
)

type mockHandler struct{}

var NewMock Factory = func(url string, str model.Stream, wBluesky events.Writer) Handler {
	return mockHandler{}
}

func (m mockHandler) Close() error {
	return nil
}

func (m mockHandler) Handle(ctx context.Context) {
	return
}
