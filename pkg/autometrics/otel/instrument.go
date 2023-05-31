package otel // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel"

import (
	"context"
	"fmt"
	"strconv"
	"time"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"go.opentelemetry.io/otel/attribute"
)

// Instrument called in a defer statement wraps the body of a function
// with automatic instrumentation.
//
// The first argument SHOULD be a call to PreInstrument so that
// the "concurrent calls" gauge is correctly setup.
func Instrument(ctx context.Context, err *error) {
	result := "ok"

	if err != nil && *err != nil {
		result = "error"
	}

	var callerLabel, sloName, latencyTarget, latencyObjective, successObjective string

	callInfo := am.GetCallInfo(ctx)
	buildInfo := am.GetBuildInfo(ctx)
	slo := am.GetAlertConfiguration(ctx)

	if am.GetTrackCallerName(ctx) {
		callerLabel = fmt.Sprintf("%s.%s", callInfo.ParentModuleName, callInfo.ParentFuncName)
	}

	if slo.ServiceName != "" {
		sloName = slo.ServiceName

		if slo.Latency != nil {
			latencyTarget = strconv.FormatFloat(slo.Latency.Target.Seconds(), 'f', -1, 64)
			latencyObjective = strconv.FormatFloat(slo.Latency.Objective, 'f', -1, 64)
		}

		if slo.Success != nil {
			successObjective = strconv.FormatFloat(slo.Success.Objective, 'f', -1, 64)
		}
	}

	functionCallsCount.Add(ctx, 1,
		[]attribute.KeyValue{
			attribute.Key(FunctionLabel).String(callInfo.FuncName),
			attribute.Key(ModuleLabel).String(callInfo.ModuleName),
			attribute.Key(CallerLabel).String(callerLabel),
			attribute.Key(ResultLabel).String(result),
			attribute.Key(TargetSuccessRateLabel).String(successObjective),
			attribute.Key(SloNameLabel).String(sloName),
			attribute.Key(CommitLabel).String(buildInfo.Commit),
			attribute.Key(VersionLabel).String(buildInfo.Version),
			attribute.Key(BranchLabel).String(buildInfo.Branch),
		}...)
	functionCallsDuration.Record(ctx, time.Since(am.GetStartTime(ctx)).Seconds(),
		[]attribute.KeyValue{
			attribute.Key(FunctionLabel).String(callInfo.FuncName),
			attribute.Key(ModuleLabel).String(callInfo.ModuleName),
			attribute.Key(CallerLabel).String(callerLabel),
			attribute.Key(TargetLatencyLabel).String(latencyTarget),
			attribute.Key(TargetSuccessRateLabel).String(latencyObjective),
			attribute.Key(SloNameLabel).String(sloName),
			attribute.Key(CommitLabel).String(buildInfo.Commit),
			attribute.Key(VersionLabel).String(buildInfo.Version),
			attribute.Key(BranchLabel).String(buildInfo.Branch),
		}...)

	if am.GetTrackConcurrentCalls(ctx) {
		functionCallsConcurrent.Add(ctx, -1,
			[]attribute.KeyValue{
				attribute.Key(FunctionLabel).String(callInfo.FuncName),
				attribute.Key(ModuleLabel).String(callInfo.ModuleName),
				attribute.Key(CallerLabel).String(callerLabel),
				attribute.Key(CommitLabel).String(buildInfo.Commit),
				attribute.Key(VersionLabel).String(buildInfo.Version),
				attribute.Key(BranchLabel).String(buildInfo.Branch),
			}...)
	}
}

// PreInstrument runs the "before wrappee" part of instrumentation.
//
// It is meant to be called as the first argument to Instrument in a
// defer call.
func PreInstrument(ctx context.Context) context.Context {
	callInfo := am.CallerInfo()
	ctx = am.SetCallInfo(ctx, callInfo)
	ctx = am.FillBuildInfo(ctx)
	ctx = am.FillTracingInfo(ctx)

	var callerLabel string
	if am.GetTrackCallerName(ctx) {
		callerLabel = fmt.Sprintf("%s.%s", callInfo.ParentModuleName, callInfo.ParentFuncName)
	}

	if am.GetTrackConcurrentCalls(ctx) {
		buildInfo := am.GetBuildInfo(ctx)
		functionCallsConcurrent.Add(ctx, 1,
			[]attribute.KeyValue{
				attribute.Key(FunctionLabel).String(callInfo.FuncName),
				attribute.Key(ModuleLabel).String(callInfo.ModuleName),
				attribute.Key(CallerLabel).String(callerLabel),
				attribute.Key(CommitLabel).String(buildInfo.Commit),
				attribute.Key(VersionLabel).String(buildInfo.Version),
				attribute.Key(BranchLabel).String(buildInfo.Branch),
			}...)
	}

	ctx = am.SetStartTime(ctx, time.Now())

	return ctx
}
