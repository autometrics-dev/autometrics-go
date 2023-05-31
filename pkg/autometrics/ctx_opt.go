package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

import (
	"context"
	"time"
)

type Option interface {
	// Apply the option to the currently created context
	Apply(context.Context) context.Context
}

type optionFunc func(context.Context) context.Context

func (fn optionFunc) Apply(ctx context.Context) context.Context {
	return fn(ctx)
}

func NewContextWithOpts(ctx context.Context, opts ...Option) context.Context {
	amCtx := NewContext(ctx)

	for _, o := range opts {
		amCtx = o.Apply(amCtx)
	}

	return amCtx
}

func WithTraceID(tid []byte) Option {
	return optionFunc(func(ctx context.Context) context.Context {
		if tid != nil {
			var truncatedTid TraceID
			copy(truncatedTid[:], tid)
			return SetTraceID(ctx, truncatedTid)
		}
		return ctx
	})
}

func WithSpanID(sid []byte) Option {
	return optionFunc(func(ctx context.Context) context.Context {
		if sid != nil {
			var truncatedSid SpanID
			copy(truncatedSid[:], sid)
			return SetSpanID(ctx, truncatedSid)
		}
		return ctx
	})
}

func WithAlertLatency(target time.Duration, objective float64) Option {
	return optionFunc(func(ctx context.Context) context.Context {
		latencySlo := &LatencySlo{
			Target:    target,
			Objective: objective,
		}
		slo := GetAlertConfiguration(ctx)
		slo.Latency = latencySlo
		return SetAlertConfiguration(ctx, slo)
	})
}

func WithAlertSuccess(objective float64) Option {
	return optionFunc(func(ctx context.Context) context.Context {
		successSlo := &SuccessSlo{
			Objective: objective,
		}
		slo := GetAlertConfiguration(ctx)
		slo.Success = successSlo
		return SetAlertConfiguration(ctx, slo)
	})
}

func WithSloName(name string) Option {
	return optionFunc(func(ctx context.Context) context.Context {
		slo := GetAlertConfiguration(ctx)
		slo.ServiceName = name
		return SetAlertConfiguration(ctx, slo)
	})
}

func WithConcurrentCalls(enabled bool) Option {
	return optionFunc(func(ctx context.Context) context.Context {
		return SetTrackConcurrentCalls(ctx, enabled)
	})
}

func WithCallerName(enabled bool) Option {
	return optionFunc(func(ctx context.Context) context.Context {
		return SetTrackCallerName(ctx, enabled)
	})
}

func WithValidHttpCodes(ranges []InclusiveIntRange) Option {
	return optionFunc(func(ctx context.Context) context.Context {
		return SetValidHttpCodeRanges(ctx, ranges)
	})
}
