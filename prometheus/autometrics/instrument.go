package autometrics // import "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/prometheus/client_golang/prometheus"
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

	info := exemplars(ctx)

	functionCallsCount.With(prometheus.Labels{
		FunctionLabel:          callInfo.FuncName,
		ModuleLabel:            callInfo.ModuleName,
		CallerFunctionLabel:    callInfo.ParentFuncName,
		CallerModuleLabel:      callInfo.ParentModuleName,
		ResultLabel:            result,
		TargetSuccessRateLabel: successObjective,
		SloNameLabel:           sloName,
		BranchLabel:            buildInfo.Branch,
		CommitLabel:            buildInfo.Commit,
		VersionLabel:           buildInfo.Version,
	}).(prometheus.ExemplarAdder).AddWithExemplar(1, info)

	functionCallsDuration.With(prometheus.Labels{
		FunctionLabel:          callInfo.FuncName,
		ModuleLabel:            callInfo.ModuleName,
		CallerFunctionLabel:    callInfo.ParentFuncName,
		CallerModuleLabel:      callInfo.ParentModuleName,
		TargetLatencyLabel:     latencyTarget,
		TargetSuccessRateLabel: latencyObjective,
		SloNameLabel:           sloName,
		BranchLabel:            buildInfo.Branch,
		CommitLabel:            buildInfo.Commit,
		VersionLabel:           buildInfo.Version,
	}).(prometheus.ExemplarObserver).ObserveWithExemplar(time.Since(am.GetStartTime(ctx)).Seconds(), info)

	if am.GetTrackConcurrentCalls(ctx) {
		functionCallsConcurrent.With(prometheus.Labels{
			FunctionLabel:       callInfo.FuncName,
			ModuleLabel:         callInfo.ModuleName,
			CallerFunctionLabel: callInfo.ParentFuncName,
			CallerModuleLabel:   callInfo.ParentModuleName,
			BranchLabel:         buildInfo.Branch,
			CommitLabel:         buildInfo.Commit,
			VersionLabel:        buildInfo.Version,
		}).Add(-1)
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
	buildInfo := am.GetBuildInfo(ctx)

	if am.GetTrackConcurrentCalls(ctx) {
		functionCallsConcurrent.With(prometheus.Labels{
			FunctionLabel:       callInfo.FuncName,
			ModuleLabel:         callInfo.ModuleName,
			CallerFunctionLabel: callInfo.ParentFuncName,
			CallerModuleLabel:   callInfo.ParentModuleName,
			BranchLabel:         buildInfo.Branch,
			CommitLabel:         buildInfo.Commit,
			VersionLabel:        buildInfo.Version,
		}).Add(1)
	}

	ctx = am.SetStartTime(ctx, time.Now())

	return ctx
}

// Extract exemplars to add to metrics from the context
func exemplars(ctx context.Context) prometheus.Labels {
	labels := make(prometheus.Labels)

	if tid, ok := am.GetTraceID(ctx); ok {
		labels[traceIdExemplar] = hex.EncodeToString(tid[:])
	}

	if sid, ok := am.GetSpanID(ctx); ok {
		labels[spanIdExemplar] = hex.EncodeToString(sid[:])
	}

	if psid, ok := am.GetParentSpanID(ctx); ok {
		labels[parentSpanIdExemplar] = hex.EncodeToString(psid[:])
	}

	return labels
}
