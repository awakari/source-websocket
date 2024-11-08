package handler

import (
	"context"
	"github.com/awakari/source-websocket/config"
	"github.com/awakari/source-websocket/model"
	"github.com/awakari/source-websocket/service/interceptor"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"io"
)

type Handler interface {
	io.Closer
	Handle(ctx context.Context)
}

type handler struct {
	url          string
	str          model.Stream
	cfgApi       config.ApiConfig
	cfgEvt       config.WebsocketConfig
	interceptors []interceptor.Interceptor

	conn *websocket.Conn
}

type Factory func(url string, str model.Stream) Handler

func NewFactory(cfgApi config.ApiConfig, cfgEvt config.WebsocketConfig, interceptors []interceptor.Interceptor) Factory {
	return func(url string, str model.Stream) Handler {
		return &handler{
			url:          url,
			str:          str,
			cfgApi:       cfgApi,
			cfgEvt:       cfgEvt,
			interceptors: interceptors,
		}
	}
}

func (h *handler) Close() error {
	return h.conn.Close(websocket.StatusNormalClosure, "")
}

func (h *handler) Handle(ctx context.Context) {
	for {
		evtN, err := h.handleStream(ctx)
		if evtN == 0 && err != nil {
			panic(err)
		}
	}
}

func (h *handler) handleStream(ctx context.Context) (evtN uint64, err error) {
	h.conn, _, err = websocket.Dial(ctx, h.url, nil)
	if err == nil {
		defer h.conn.CloseNow()
		for {
			err = h.handleStreamEvent(ctx, h.url)
			if err != nil {
				break
			}
		}
	}
	return
}

func (h *handler) handleStreamEvent(ctx context.Context, url string) (err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, h.cfgEvt.StreamTimeout)
	defer cancel()
	var raw map[string]any
	err = wsjson.Read(ctxWithTimeout, h.conn, &raw)
	if err == nil {
		var matched bool
		for _, i := range h.interceptors {
			if matched, err = i.Handle(ctx, url, raw); matched {
				break
			}
		}
	}
	return
}
