package interceptor

import (
	"context"
	"github.com/awakari/source-websocket/service/writer"
)

type defaultInterceptor struct {
	w writer.Service
}

func NewDefault(w writer.Service) Interceptor {
	return defaultInterceptor{
		w: w,
	}
}

func (d defaultInterceptor) Handle(ctx context.Context, url string, raw map[string]any) (matches bool, err error) {
	matches = true
	return
}
