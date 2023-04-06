package otel // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel"

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"go.opentelemetry.io/otel/attribute"
)

// Instrument called in a defer statement wraps the body of a function
// with automatic instrumentation.
//
// The first argument SHOULD be a call to PreInstrument so that
// the "concurrent calls" gauge is correctly setup.
func Instrument(ctx *autometrics.Context, err *error) {
	result := "ok"

	if err != nil && *err != nil {
		result = "error"
	}

	var callerLabel, sloName, latencyTarget, latencyObjective, successObjective string

	if ctx.TrackCallerName {
		callerLabel = fmt.Sprintf("%s.%s", ctx.CallInfo.ParentModuleName, ctx.CallInfo.ParentFuncName)
	}

	if ctx.AlertConf != nil {
		sloName = ctx.AlertConf.ServiceName

		if ctx.AlertConf.Latency != nil {
			latencyTarget = strconv.FormatFloat(ctx.AlertConf.Latency.Target.Seconds(), 'f', -1, 64)
			latencyObjective = strconv.FormatFloat(ctx.AlertConf.Latency.Objective, 'f', -1, 64)
		}

		if ctx.AlertConf.Success != nil {
			successObjective = strconv.FormatFloat(ctx.AlertConf.Success.Objective, 'f', -1, 64)
		}
	}

	functionCallsCount.Add(ctx.Context, 1,
		[]attribute.KeyValue{
			attribute.Key(FunctionLabel).String(ctx.CallInfo.FuncName),
			attribute.Key(ModuleLabel).String(ctx.CallInfo.ModuleName),
			attribute.Key(CallerLabel).String(callerLabel),
			attribute.Key(ResultLabel).String(result),
			attribute.Key(TargetSuccessRateLabel).String(successObjective),
			attribute.Key(SloNameLabel).String(sloName),
		}...)
	functionCallsDuration.Record(ctx.Context, time.Since(ctx.StartTime).Seconds(),
		[]attribute.KeyValue{
			attribute.Key(FunctionLabel).String(ctx.CallInfo.FuncName),
			attribute.Key(ModuleLabel).String(ctx.CallInfo.ModuleName),
			attribute.Key(CallerLabel).String(callerLabel),
			attribute.Key(TargetLatencyLabel).String(latencyTarget),
			attribute.Key(TargetSuccessRateLabel).String(latencyObjective),
			attribute.Key(SloNameLabel).String(sloName),
		}...)

	if ctx.TrackConcurrentCalls {
		functionCallsConcurrent.Add(ctx.Context, -1,
			[]attribute.KeyValue{
				attribute.Key(FunctionLabel).String(ctx.CallInfo.FuncName),
				attribute.Key(ModuleLabel).String(ctx.CallInfo.ModuleName),
				attribute.Key(CallerLabel).String(callerLabel),
			}...)
	}
}

// PreInstrument runs the "before wrappee" part of instrumentation.
//
// It is meant to be called as the first argument to Instrument in a
// defer call.
func PreInstrument(ctx *autometrics.Context) *autometrics.Context {
	ctx.CallInfo = autometrics.CallerInfo()
	ctx.Context = context.Background()

	var callerLabel string
	if ctx.TrackCallerName {
		callerLabel = fmt.Sprintf("%s.%s", ctx.CallInfo.ParentModuleName, ctx.CallInfo.ParentFuncName)
	}

	if ctx.TrackConcurrentCalls {
		functionCallsConcurrent.Add(ctx.Context, 1,
			[]attribute.KeyValue{
				attribute.Key(FunctionLabel).String(ctx.CallInfo.FuncName),
				attribute.Key(ModuleLabel).String(ctx.CallInfo.ModuleName),
				attribute.Key(CallerLabel).String(callerLabel),
			}...)
	}

	ctx.StartTime = time.Now()

	return ctx
}
