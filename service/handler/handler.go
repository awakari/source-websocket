package handler

import (
	"context"
	"errors"
	"github.com/awakari/source-websocket/config"
	"github.com/awakari/source-websocket/model"
	"github.com/awakari/source-websocket/service/converter"
	"github.com/awakari/source-websocket/service/writer"
	"github.com/cenkalti/backoff/v4"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"io"
)

type Handler interface {
	io.Closer
	Handle(ctx context.Context)
}

type handler struct {
	url    string
	str    model.Stream
	cfgApi config.ApiConfig
	conv   converter.Service
	w      writer.Service

	conn *websocket.Conn
}

type Factory func(url string, str model.Stream) Handler

func NewFactory(cfgApi config.ApiConfig, conv converter.Service, w writer.Service) Factory {
	return func(url string, str model.Stream) Handler {
		return &handler{
			url:    url,
			str:    str,
			cfgApi: cfgApi,
			conv:   conv,
			w:      w,
		}
	}
}

func (h *handler) Close() error {
	return h.conn.Close(websocket.StatusNormalClosure, "")
}

func (h *handler) Handle(ctx context.Context) {
	b := backoff.NewExponentialBackOff()
	f := func() error {
		return h.handleStream(ctx)
	}
	for {
		if err := backoff.Retry(f, b); err != nil {
			panic(err)
		}
	}
}

func (h *handler) handleStream(ctx context.Context) (err error) {
	h.conn, _, err = websocket.Dial(ctx, h.url, nil)
	if err == nil {
		defer h.conn.CloseNow()
		if h.str.Request != "" {
			err = wsjson.Write(ctx, h.conn, h.str.Request)
		}
		if err == nil {
			for {
				err = h.handleStreamEvent(ctx, h.url)
				if err != nil && !errors.Is(err, converter.ErrConversion) {
					break
				}
			}
		}
	}
	return
}

func (h *handler) handleStreamEvent(ctx context.Context, url string) (err error) {
	var raw map[string]any
	err = wsjson.Read(ctx, h.conn, &raw)
	var evt *pb.CloudEvent
	if err == nil {
		evt, err = h.conv.Convert(url, raw)
	}
	if err == nil {
		err = h.w.Write(ctx, evt, h.cfgApi.GroupId, url)
	}
	return
}
