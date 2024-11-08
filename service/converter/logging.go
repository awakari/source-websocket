package converter

import (
	"fmt"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"log/slog"
)

type logging struct {
	svc Service
	log *slog.Logger
}

func NewLogging(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) Convert(src string, raw map[string]any) (evt *pb.CloudEvent, err error) {
	evt, err = l.svc.Convert(src, raw)
	switch err {
	case nil:
		l.log.Debug(fmt.Sprintf("converter.Convert(%s): evt.Id=%s", src, evt.Id))
	default:
		l.log.Warn(fmt.Sprintf("converter.Convert(%s, %+v): %s", src, raw, err))
	}
	return
}
