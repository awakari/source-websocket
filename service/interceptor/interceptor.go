package interceptor

import (
	"context"
)

type Interceptor interface {
	Handle(ctx context.Context, src string, raw map[string]any) (matches bool, err error)
}
