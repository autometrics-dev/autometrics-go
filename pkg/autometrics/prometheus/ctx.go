package prometheus // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"

import (
	"context"
	"time"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

type optionFunc func(*autometrics.Context)

func (fn optionFunc) Apply(ctx *autometrics.Context) {
	fn(ctx)
}

func NewContext(ctx context.Context, opts ...autometrics.Option) *autometrics.Context {
	amCtx := autometrics.NewContext(ctx)

	for _, o := range(opts) {
		o.Apply(&amCtx)
	}

	return &amCtx
}

func WithTraceID(tid []byte) autometrics.Option {
	return optionFunc(func(ctx *autometrics.Context) {
		if tid != nil {
			var truncatedTid autometrics.TraceID
			copy(truncatedTid[:], tid)
			ctx.SetTraceID(truncatedTid)
		}
	})
}

func WithSpanID(sid []byte) autometrics.Option {
	return optionFunc(func(ctx *autometrics.Context) {
		if sid != nil {
			var truncatedSid autometrics.SpanID
			copy(truncatedSid[:], sid)
			ctx.SetSpanID(truncatedSid)
		}
	})
}

func WithAlertLatency(target time.Duration, objective float64) autometrics.Option {
	return optionFunc(func(ctx *autometrics.Context) {
		latencySlo := &autometrics.LatencySlo{
			Target:    target,
			Objective: objective,
		}
		if ctx.AlertConf != nil {
			ctx.AlertConf.Latency = latencySlo
		} else {
			ctx.AlertConf = &autometrics.AlertConfiguration{
				Latency: latencySlo,
			}
		}
	})
}

func WithAlertSuccess(objective float64) autometrics.Option {
	return optionFunc(func(ctx *autometrics.Context) {
		successSlo := &autometrics.SuccessSlo{
			Objective: objective,
		}
		if ctx.AlertConf != nil {
			ctx.AlertConf.Success = successSlo
		} else {
			ctx.AlertConf = &autometrics.AlertConfiguration{
				Success: successSlo,
			}
		}
	})
}

func WithSloName(name string) autometrics.Option {
	return optionFunc(func(ctx *autometrics.Context) {
		if ctx.AlertConf != nil {
			ctx.AlertConf.ServiceName = name
		} else {
			ctx.AlertConf = &autometrics.AlertConfiguration{
				ServiceName: name,
			}
		}
	})
}

func WithConcurrentCalls(enabled bool) autometrics.Option {
	return optionFunc(func(ctx *autometrics.Context) {
		ctx.TrackConcurrentCalls = enabled
	})
}

func WithCallerName(enabled bool) autometrics.Option {
	return optionFunc(func(ctx *autometrics.Context) {
		ctx.TrackCallerName = enabled
	})
}
