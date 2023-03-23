package autometrics

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Instrument called in a defer statement wraps the body of a function
// with automatic instrumentation.
//
// The first argument SHOULD be a call to PreInstrument so that
// the "concurrent calls" gauge is correctly setup.
func Instrument(ctx *Context, err *error) {
	result := "ok"

	if err != nil && *err != nil {
		result = "error"
	}

	var callerLabel, sloName, latencyTarget, latencyObjective, successObjective string
	if ctx.TrackCallerName {
		callerLabel = fmt.Sprintf("%s.%s", ctx.callInfo.ParentModuleName, ctx.callInfo.ParentFuncName)
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

	FunctionCallsCount.With(prometheus.Labels{
		FunctionLabel:          ctx.callInfo.FuncName,
		ModuleLabel:            ctx.callInfo.ModuleName,
		CallerLabel:            callerLabel,
		ResultLabel:            result,
		TargetSuccessRateLabel: successObjective,
		SloNameLabel:           sloName,
	}).Inc()
	FunctionCallsDuration.With(prometheus.Labels{
		FunctionLabel:          ctx.callInfo.FuncName,
		ModuleLabel:            ctx.callInfo.ModuleName,
		CallerLabel:            callerLabel,
		TargetLatencyLabel:     latencyTarget,
		TargetSuccessRateLabel: latencyObjective,
		SloNameLabel:           sloName,
	}).Observe(time.Since(ctx.startTime).Seconds())
	if ctx.TrackConcurrentCalls {
		FunctionCallsConcurrent.With(prometheus.Labels{
			FunctionLabel: ctx.callInfo.FuncName,
			ModuleLabel:   ctx.callInfo.ModuleName,
			CallerLabel:   callerLabel,
		}).Dec()
	}
}

// PreInstrument runs the "before wrappee" part of instrumentation.
//
// It is meant to be called as the first argument to Instrument in a
// defer call.
func PreInstrument(ctx *Context) *Context {
	ctx.callInfo = callerInfo()

	var callerLabel string
	if ctx.TrackCallerName {
		callerLabel = fmt.Sprintf("%s.%s", ctx.callInfo.ParentModuleName, ctx.callInfo.ParentFuncName)
	}

	FunctionCallsConcurrent.With(prometheus.Labels{
		FunctionLabel: ctx.callInfo.FuncName,
		ModuleLabel:   ctx.callInfo.ModuleName,
		CallerLabel:   callerLabel,
	}).Inc()

	ctx.startTime = time.Now()

	return ctx
}

// callerInfo returns the (method name, module name) of the function that called the function that called this function.
//
// It also returns the information about its grandparent.
//
// The module name and the parent module names are cropped to their last part, because the generator we use
// only has access to the last "package" name in `GOPACKAGE` environment variable.
//
// If there is a way to obtain programmatically the fully qualified package name in go-generate arguments,
// then we can lift this artificial limitation here and use the full "module name" from the caller information.
// Currently this compromise is the only way to have the documentation links generator creating correct
// queries.
func callerInfo() (callInfo CallInfo) {
	programCounters := make([]uintptr, 15)

	// skip 3 frames to start with:
	// frame 0: internal function called by `runtime.Callers`
	// frame 1: us calling `runtime.Callers` (this function)
	// frame 2: Instrument() calling this function -- we don't really care about our own library code
	entries := runtime.Callers(3, programCounters)

	frames := runtime.CallersFrames(programCounters[:entries])
	frame, hasParent := frames.Next()

	functionName := frame.Function
	index := strings.LastIndex(functionName, ".")

	if index == -1 {
		callInfo.FuncName = frame.Func.Name()
	} else {
		moduleIndex := strings.LastIndex(functionName[:index], ".")
		if moduleIndex == -1 {
			callInfo.ModuleName = functionName[:index]
		} else {
			callInfo.ModuleName = functionName[moduleIndex+1 : index]
		}

		callInfo.FuncName = functionName[index+1:]
	}

	if !hasParent {
		return
	}

	// Do the same with the parent
	parentFrame, _ := frames.Next()

	parentFunctionName := parentFrame.Function
	index = strings.LastIndex(parentFunctionName, ".")

	if index == -1 {
		callInfo.ParentFuncName = parentFrame.Func.Name()
	} else {
		moduleIndex := strings.LastIndex(parentFunctionName[:index], ".")
		if moduleIndex == -1 {
			callInfo.ParentModuleName = parentFunctionName[:index]
		} else {
			callInfo.ParentModuleName = parentFunctionName[moduleIndex+1 : index]
		}

		callInfo.ParentFuncName = parentFunctionName[index+1:]
	}

	return
}
