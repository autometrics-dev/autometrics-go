package autometrics

import (
	"fmt"
	"runtime"
	"strings"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Instrument called in a defer statement wraps the body of a function
// with automatic instrumentation.
//
// The first argument SHOULD be a call to PreInstrument so that
// the "concurrent calls" gauge is correctly setup.
func Instrument(ctx Context, startTime time.Time, err *error) {
	method, module, parentMethod, parentModule := callerInfo()
	result := "ok"

	if err != nil && *err != nil {
		result = "error"
	}

	var callerLabel, sloName, latencyTarget, latencyObjective, successObjective string
	if ctx.TrackCallerName {
		callerLabel = fmt.Sprintf("%s.%s", parentModule, parentMethod)
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
		FunctionLabel:          method,
		ModuleLabel:            module,
		CallerLabel:            callerLabel,
		ResultLabel:            result,
		TargetSuccessRateLabel: successObjective,
		SloNameLabel:           sloName,
	}).Inc()
	FunctionCallsDuration.With(prometheus.Labels{
		FunctionLabel:          method,
		ModuleLabel:            module,
		CallerLabel:            callerLabel,
		TargetLatencyLabel:     latencyTarget,
		TargetSuccessRateLabel: latencyObjective,
		SloNameLabel:           sloName,
	}).Observe(time.Since(startTime).Seconds())
	if ctx.TrackConcurrentCalls {
		FunctionCallsConcurrent.With(prometheus.Labels{
			FunctionLabel: method,
			ModuleLabel:   module,
			CallerLabel:   callerLabel,
		}).Dec()
	}
}

// PreInstrument runs the "before wrappee" part of instrumentation.
//
// It is meant to be called as the first argument to Instrument in a
// defer call.
func PreInstrument(ctx Context) time.Time {
	method, module, parentMethod, parentModule := callerInfo()

	var callerLabel string
	if ctx.TrackCallerName {
		callerLabel = fmt.Sprintf("%s.%s", parentModule, parentMethod)
	}

	FunctionCallsConcurrent.With(prometheus.Labels{FunctionLabel: method, ModuleLabel: module, CallerLabel: callerLabel}).Inc()

	return time.Now()
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
func callerInfo() (funcName, moduleName, parentFuncName, parentModuleName string) {
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
		funcName = frame.Func.Name()
	} else {
		moduleIndex := strings.LastIndex(functionName[:index], ".")
		if moduleIndex == -1 {
			moduleName = functionName[:index]
		} else {
			moduleName = functionName[moduleIndex+1 : index]
		}

		funcName = functionName[index+1:]
	}

	if !hasParent {
		return
	}

	// Do the same with the parent
	parentFrame, _ := frames.Next()

	parentFunctionName := parentFrame.Function
	index = strings.LastIndex(parentFunctionName, ".")

	if index == -1 {
		parentFuncName = parentFrame.Func.Name()
	} else {
		moduleIndex := strings.LastIndex(parentFunctionName[:index], ".")
		if moduleIndex == -1 {
			parentModuleName = parentFunctionName[:index]
		} else {
			parentModuleName = parentFunctionName[moduleIndex+1 : index]
		}

		parentFuncName = parentFunctionName[index+1:]
	}

	return
}
