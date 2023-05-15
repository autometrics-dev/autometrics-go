package prometheus // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/prometheus/client_golang/prometheus"
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

	info := exemplars(ctx)

	functionCallsCount.With(prometheus.Labels{
		FunctionLabel:          ctx.CallInfo.FuncName,
		ModuleLabel:            ctx.CallInfo.ModuleName,
		CallerLabel:            callerLabel,
		ResultLabel:            result,
		TargetSuccessRateLabel: successObjective,
		SloNameLabel:           sloName,
		BranchLabel:            ctx.BuildInfo.Branch,
		CommitLabel:            ctx.BuildInfo.Commit,
		VersionLabel:           ctx.BuildInfo.Version,
	}).(prometheus.ExemplarAdder).AddWithExemplar(1, info)
	functionCallsDuration.With(prometheus.Labels{
		FunctionLabel:          ctx.CallInfo.FuncName,
		ModuleLabel:            ctx.CallInfo.ModuleName,
		CallerLabel:            callerLabel,
		TargetLatencyLabel:     latencyTarget,
		TargetSuccessRateLabel: latencyObjective,
		SloNameLabel:           sloName,
		BranchLabel:            ctx.BuildInfo.Branch,
		CommitLabel:            ctx.BuildInfo.Commit,
		VersionLabel:           ctx.BuildInfo.Version,
	}).(prometheus.ExemplarObserver).ObserveWithExemplar(time.Since(ctx.StartTime).Seconds(), info)

	if ctx.TrackConcurrentCalls {
		functionCallsConcurrent.With(prometheus.Labels{
			FunctionLabel: ctx.CallInfo.FuncName,
			ModuleLabel:   ctx.CallInfo.ModuleName,
			CallerLabel:   callerLabel,
			BranchLabel:   ctx.BuildInfo.Branch,
			CommitLabel:   ctx.BuildInfo.Commit,
			VersionLabel:  ctx.BuildInfo.Version,
		}).Add(-1)
	}
}

// PreInstrument runs the "before wrappee" part of instrumentation.
//
// It is meant to be called as the first argument to Instrument in a
// defer call.
func PreInstrument(ctx *autometrics.Context) *autometrics.Context {
	ctx.CallInfo = autometrics.CallerInfo()
	ctx.FillBuildInfo()
	ctx.FillTracingInfo()

	var callerLabel string
	if ctx.TrackCallerName {
		callerLabel = fmt.Sprintf("%s.%s", ctx.CallInfo.ParentModuleName, ctx.CallInfo.ParentFuncName)
	}

	if ctx.TrackConcurrentCalls {
		functionCallsConcurrent.With(prometheus.Labels{
			FunctionLabel: ctx.CallInfo.FuncName,
			ModuleLabel:   ctx.CallInfo.ModuleName,
			CallerLabel:   callerLabel,
			BranchLabel:   ctx.BuildInfo.Branch,
			CommitLabel:   ctx.BuildInfo.Commit,
			VersionLabel:  ctx.BuildInfo.Version,
		}).Add(1)
	}

	ctx.StartTime = time.Now()

	return ctx
}

// Extract exemplars to add to metrics from the context
func exemplars(ctx *autometrics.Context) prometheus.Labels {
	labels := make(prometheus.Labels)

	if tid, ok := ctx.GetTraceID(); ok {
		labels[traceIdExemplar] = hex.EncodeToString(tid[:])
	}

	if sid, ok := ctx.GetSpanID(); ok {
		labels[spanIdExemplar] = hex.EncodeToString(sid[:])
	}

	if psid, ok := ctx.GetParentSpanID(); ok {
		labels[parentSpanIdExemplar] = hex.EncodeToString(psid[:])
	}

	return labels
}
