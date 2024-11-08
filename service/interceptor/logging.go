package interceptor

import (
	"context"
	"fmt"
	"log/slog"
)

type logging struct {
	i   Interceptor
	log *slog.Logger
	t   string
}

func NewLogging(i Interceptor, log *slog.Logger, t string) Interceptor {
	return logging{
		i:   i,
		log: log,
		t:   t,
	}
}

func (l logging) Handle(ctx context.Context, src string, raw map[string]any) (matches bool, err error) {
	if matches, err = l.i.Handle(ctx, src, raw); matches {
		switch err {
		case nil:
			l.log.Debug(fmt.Sprintf("interceptor(%s).Handle(%s, %+v): ok", l.t, src, raw))
		default:
			l.log.Error(fmt.Sprintf("interceptor(%s).Handle(%s, %+v): %s", l.t, src, raw, err))
		}
	}
	return
}
