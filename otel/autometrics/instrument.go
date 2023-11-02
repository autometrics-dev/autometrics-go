package autometrics // import "github.com/autometrics-dev/autometrics-go/otel/autometrics"

import (
	"context"
	"strconv"
	"time"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Instrument called in a defer statement wraps the body of a function
// with automatic instrumentation.
//
// The first argument SHOULD be a call to PreInstrument so that
// the "concurrent calls" gauge is correctly setup.
func Instrument(ctx context.Context, err *error) {
	if amCtx.Err() != nil {
		return
	}

	result := "ok"

	if err != nil && *err != nil {
		result = "error"
	}

	var sloName, latencyTarget, latencyObjective, successObjective string

	callInfo := am.GetCallInfo(ctx)
	buildInfo := am.GetBuildInfo(ctx)
	slo := am.GetAlertConfiguration(ctx)

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
		metric.WithAttributes([]attribute.KeyValue{
			attribute.Key(FunctionLabel).String(callInfo.Current.Function),
			attribute.Key(ModuleLabel).String(callInfo.Current.Module),
			attribute.Key(CallerFunctionLabel).String(callInfo.Parent.Function),
			attribute.Key(CallerModuleLabel).String(callInfo.Parent.Module),
			attribute.Key(ResultLabel).String(result),
			attribute.Key(TargetSuccessRateLabel).String(successObjective),
			attribute.Key(SloNameLabel).String(sloName),
			attribute.Key(CommitLabel).String(buildInfo.Commit),
			attribute.Key(VersionLabel).String(buildInfo.Version),
			attribute.Key(BranchLabel).String(buildInfo.Branch),
			attribute.Key(ServiceNameLabel).String(buildInfo.Service),
			attribute.Key(JobNameLabel).String(am.GetPushJobName()),
		}...))
	functionCallsDuration.Record(ctx, time.Since(am.GetStartTime(ctx)).Seconds(),
		metric.WithAttributes([]attribute.KeyValue{
			attribute.Key(FunctionLabel).String(callInfo.Current.Function),
			attribute.Key(ModuleLabel).String(callInfo.Current.Module),
			attribute.Key(CallerFunctionLabel).String(callInfo.Parent.Function),
			attribute.Key(CallerModuleLabel).String(callInfo.Parent.Module),
			attribute.Key(TargetLatencyLabel).String(latencyTarget),
			attribute.Key(TargetSuccessRateLabel).String(latencyObjective),
			attribute.Key(SloNameLabel).String(sloName),
			attribute.Key(CommitLabel).String(buildInfo.Commit),
			attribute.Key(VersionLabel).String(buildInfo.Version),
			attribute.Key(BranchLabel).String(buildInfo.Branch),
			attribute.Key(ServiceNameLabel).String(buildInfo.Service),
			attribute.Key(JobNameLabel).String(am.GetPushJobName()),
		}...))

	if am.GetTrackConcurrentCalls(ctx) {
		functionCallsConcurrent.Add(ctx, -1,
			metric.WithAttributes([]attribute.KeyValue{
				attribute.Key(FunctionLabel).String(callInfo.Current.Function),
				attribute.Key(ModuleLabel).String(callInfo.Current.Module),
				attribute.Key(CallerFunctionLabel).String(callInfo.Parent.Function),
				attribute.Key(CallerModuleLabel).String(callInfo.Parent.Module),
				attribute.Key(CommitLabel).String(buildInfo.Commit),
				attribute.Key(VersionLabel).String(buildInfo.Version),
				attribute.Key(BranchLabel).String(buildInfo.Branch),
				attribute.Key(ServiceNameLabel).String(buildInfo.Service),
				attribute.Key(JobNameLabel).String(am.GetPushJobName()),
			}...))
	}

	// NOTE: This call means that goroutines that outlive this function as the caller will not have access to parent
	// caller information, but hopefully by that point we got all the necessary accesses done.
	// If not, it is a convenience we accept to give up to prevent memory usage from exploding
	_ = am.PopFunctionName(ctx)
}

// PreInstrument runs the "before wrappee" part of instrumentation.
//
// It is meant to be called as the first argument to Instrument in a
// defer call.
func PreInstrument(ctx context.Context) context.Context {
	if amCtx.Err() != nil {
		return nil
	}

	ctx = am.FillTracingAndCallerInfo(ctx)
	ctx = am.FillBuildInfo(ctx)

	callInfo := am.GetCallInfo(ctx)

	if am.GetTrackConcurrentCalls(ctx) {
		buildInfo := am.GetBuildInfo(ctx)
		functionCallsConcurrent.Add(ctx, 1,
			metric.WithAttributes([]attribute.KeyValue{
				attribute.Key(FunctionLabel).String(callInfo.Current.Function),
				attribute.Key(ModuleLabel).String(callInfo.Current.Module),
				attribute.Key(CallerFunctionLabel).String(callInfo.Parent.Function),
				attribute.Key(CallerModuleLabel).String(callInfo.Parent.Module),
				attribute.Key(CommitLabel).String(buildInfo.Commit),
				attribute.Key(VersionLabel).String(buildInfo.Version),
				attribute.Key(BranchLabel).String(buildInfo.Branch),
				attribute.Key(ServiceNameLabel).String(buildInfo.Service),
				attribute.Key(JobNameLabel).String(am.GetPushJobName()),
			}...))
	}

	ctx = am.SetStartTime(ctx, time.Now())

	return ctx
}
