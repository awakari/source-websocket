package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/source-websocket/api/grpc/events"
	"github.com/awakari/source-websocket/api/http/pub"
	"github.com/awakari/source-websocket/config"
	"github.com/awakari/source-websocket/model"
	"github.com/awakari/source-websocket/service/converter"
	"github.com/bytedance/sonic"
	"github.com/cenkalti/backoff/v4"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/segmentio/ksuid"
	"golang.org/x/time/rate"
	"io"
	"log/slog"
	"math"
	"time"
)

type Handler interface {
	io.Closer
	Handle(ctx context.Context)
}

type handler struct {
	url      string
	str      model.Stream
	cfgApi   config.ApiConfig
	conv     converter.Service
	svcPub   pub.Service
	log      *slog.Logger
	wBluesky events.Writer

	conn *websocket.Conn
	rl   *rate.Limiter
}

type Factory func(url string, str model.Stream, wBluesky events.Writer) Handler

const fmtBluesky = "bluesky"

func NewFactory(cfgApi config.ApiConfig, conv converter.Service, svcPub pub.Service, log *slog.Logger) Factory {
	return func(url string, str model.Stream, wBluesky events.Writer) Handler {
		var rl *rate.Limiter
		switch str.RateLimit {
		case nil:
			rl = rate.NewLimiter(rate.Inf, math.MaxInt)
		default:
			rl = rate.NewLimiter(rate.Limit(*str.RateLimit), int(*str.RateLimit))
		}
		return &handler{
			url:      url,
			str:      str,
			cfgApi:   cfgApi,
			conv:     conv,
			svcPub:   svcPub,
			log:      log,
			wBluesky: wBluesky,
			rl:       rl,
		}
	}
}

func (h *handler) Close() error {
	return h.conn.Close(websocket.StatusNormalClosure, "")
}

func (h *handler) Handle(ctx context.Context) {
	b := backoff.NewExponentialBackOff()
	handleFunc := func() error {
		return h.handleStream(ctx)
	}
	notifyErrFunc := func(err error, t time.Duration) {
		h.log.Warn(fmt.Sprintf("failed to handle the stream from %s, cause: %s, retrying in %s", h.url, err, t))
	}
	for {
		if err := backoff.RetryNotify(handleFunc, b, notifyErrFunc); err != nil {
			panic(fmt.Sprintf("failed to handle the stream from %s, cause: %s", h.url, err))
		}
	}
}

func (h *handler) handleStream(ctx context.Context) (err error) {
	h.conn, _, err = websocket.Dial(ctx, h.url, nil)
	if err == nil {
		defer h.conn.CloseNow()
		if h.str.Request != "" {
			var reqParsed map[string]any
			err = sonic.Unmarshal([]byte(h.str.Request), &reqParsed)
			if err == nil {
				err = wsjson.Write(ctx, h.conn, reqParsed)
			}

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
	switch h.rl.Allow() {
	case true:
		switch h.str.Fmt {
		case fmtBluesky:
			var data []byte
			_, data, err = h.conn.Read(ctx)
			if err == nil {
				evt := &pb.CloudEvent{
					Id:          ksuid.New().String(),
					Source:      url,
					SpecVersion: model.CeSpecVersion,
					Type:        h.cfgApi.Events.Type,
					Data: &pb.CloudEvent_BinaryData{
						BinaryData: data,
					},
				}
				var ackCount uint32
				ackCount, err = h.wBluesky.Write(ctx, []*pb.CloudEvent{
					evt,
				})
				if err != nil {
					panic("bluesky handler: failed to write event: " + err.Error())
				}
				if ackCount < 1 {
					panic("bluesky handler: failed to acknowledge event")
				}
			}
		default:
			var raw map[string]any
			err = wsjson.Read(ctx, h.conn, &raw)
			var evt *pb.CloudEvent
			if err == nil && raw != nil {
				evt, err = h.conv.Convert(url, raw)
			}
			if err == nil && evt != nil {
				err = h.svcPub.Publish(ctx, evt, h.cfgApi.GroupId, url)
			}
		}
	default:
		time.Sleep(1 * time.Second)
	}
	return
}
