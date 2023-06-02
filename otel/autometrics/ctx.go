package autometrics // import "github.com/autometrics-dev/autometrics-go/otel/autometrics"

import (
	"context"
	"time"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

type ValidHttpRange = autometrics.InclusiveIntRange

func NewContext(ctx context.Context, opts ...autometrics.Option) context.Context {
	return autometrics.NewContextWithOpts(ctx, opts...)
}

func WithTraceID(tid []byte) autometrics.Option {
	return autometrics.WithTraceID(tid)
}

func WithSpanID(sid []byte) autometrics.Option {
	return autometrics.WithSpanID(sid)
}

func WithAlertLatency(target time.Duration, objective float64) autometrics.Option {
	return autometrics.WithAlertLatency(target, objective)
}

func WithAlertSuccess(objective float64) autometrics.Option {
	return autometrics.WithAlertSuccess(objective)
}

func WithSloName(name string) autometrics.Option {
	return autometrics.WithSloName(name)
}

func WithConcurrentCalls(enabled bool) autometrics.Option {
	return autometrics.WithConcurrentCalls(enabled)
}

func WithCallerName(enabled bool) autometrics.Option {
	return autometrics.WithCallerName(enabled)
}

func WithValidHttpCodes(ranges []ValidHttpRange) autometrics.Option {
	return autometrics.WithValidHttpCodes(ranges)
}
