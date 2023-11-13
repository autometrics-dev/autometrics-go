package autometrics // import "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"

import (
	"context"
	"encoding/hex"
	"log"
	"strconv"
	"time"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
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

	info := exemplars(ctx)

	functionCallsCount.With(prometheus.Labels{
		FunctionLabel:          callInfo.Current.Function,
		ModuleLabel:            callInfo.Current.Module,
		CallerFunctionLabel:    callInfo.Parent.Function,
		CallerModuleLabel:      callInfo.Parent.Module,
		ResultLabel:            result,
		TargetSuccessRateLabel: successObjective,
		SloNameLabel:           sloName,
		BranchLabel:            buildInfo.Branch,
		CommitLabel:            buildInfo.Commit,
		VersionLabel:           buildInfo.Version,
		ServiceNameLabel:       buildInfo.Service,
	}).(prometheus.ExemplarAdder).AddWithExemplar(1, info)

	functionCallsDuration.With(prometheus.Labels{
		FunctionLabel:          callInfo.Current.Function,
		ModuleLabel:            callInfo.Current.Module,
		CallerFunctionLabel:    callInfo.Parent.Function,
		CallerModuleLabel:      callInfo.Parent.Module,
		TargetLatencyLabel:     latencyTarget,
		TargetSuccessRateLabel: latencyObjective,
		SloNameLabel:           sloName,
		BranchLabel:            buildInfo.Branch,
		CommitLabel:            buildInfo.Commit,
		VersionLabel:           buildInfo.Version,
		ServiceNameLabel:       buildInfo.Service,
	}).(prometheus.ExemplarObserver).ObserveWithExemplar(time.Since(am.GetStartTime(ctx)).Seconds(), info)

	if am.GetTrackConcurrentCalls(ctx) {
		functionCallsConcurrent.With(prometheus.Labels{
			FunctionLabel:       callInfo.Current.Function,
			ModuleLabel:         callInfo.Current.Module,
			CallerFunctionLabel: callInfo.Parent.Function,
			CallerModuleLabel:   callInfo.Parent.Module,
			BranchLabel:         buildInfo.Branch,
			CommitLabel:         buildInfo.Commit,
			VersionLabel:        buildInfo.Version,
			ServiceNameLabel:    buildInfo.Service,
		}).Add(-1)
	}

	if pusher != nil {
		go func(parentCtx context.Context) {
			ctx, cancel := context.WithCancel(parentCtx)
			defer cancel()
			// PERF: This might induce way too much contention and a growing number of goroutines
			if pusherLock.TryLock() {
				defer pusherLock.Unlock()
				localPusher := push.
					New(am.GetPushJobURL(), am.GetPushJobName()).
					Format(expfmt.FmtText).
					Collector(functionCallsCount).
					Collector(functionCallsDuration).
					Collector(functionCallsConcurrent)
				if err := localPusher.
					AddContext(ctx); err != nil {
					log.Printf("failed to push metrics to gateway: %s", err)
				}
			}
		}(amCtx)
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
	buildInfo := am.GetBuildInfo(ctx)
	callInfo := am.GetCallInfo(ctx)

	if am.GetTrackConcurrentCalls(ctx) {
		functionCallsConcurrent.With(prometheus.Labels{
			FunctionLabel:       callInfo.Current.Function,
			ModuleLabel:         callInfo.Current.Module,
			CallerFunctionLabel: callInfo.Parent.Function,
			CallerModuleLabel:   callInfo.Parent.Module,
			BranchLabel:         buildInfo.Branch,
			CommitLabel:         buildInfo.Commit,
			VersionLabel:        buildInfo.Version,
			ServiceNameLabel:    buildInfo.Service,
		}).Add(1)
	}

	if pusher != nil {
		go func(parentCtx context.Context) {
			ctx, cancel := context.WithCancel(parentCtx)
			defer cancel()
			// PERF: Using Lock might induce way too much contention and a growing number of goroutines
			if pusherLock.TryLock() {
				defer pusherLock.Unlock()
				localPusher := push.
					New(am.GetPushJobURL(), am.GetPushJobName()).
					Format(expfmt.FmtText).
					Collector(functionCallsConcurrent)
				if err := localPusher.AddContext(ctx); err != nil {
					log.Printf("failed to push metrics to gateway: %s", err)
				}
			}
		}(amCtx)
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
