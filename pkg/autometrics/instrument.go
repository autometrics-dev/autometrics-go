package autometrics

import (
	"runtime"
	"strings"
	"time"
)

func Instrument(startTime time.Time, err *error) {
	method, module := callerInfo()
	result := "ok"

	if err != nil && *err != nil {
		result = "error"
	}

	FunctionCallsCount.WithLabelValues(method, module, result).Inc()
	FunctionCallsDuration.WithLabelValues(method, module).Observe(time.Since(startTime).Seconds())
	FunctionCallsConcurrent.WithLabelValues(method, module).Dec()
}

func PreInstrument() time.Time {
	method, module := callerInfo()

	FunctionCallsConcurrent.WithLabelValues(method, module).Inc()

	return time.Now()
}

// callerInfo returns the (method name, module name) of the function that called the function that called this function
func callerInfo() (string, string) {
	programCounters := make([]uintptr, 15)

	// skip 3 frames to start with:
	// frame 0: internal function called by `runtime.Callers`
	// frame 1: us calling `runtime.Callers` (this function)
	// frame 2: Instrument() calling this function -- we don't really care about our own library code
	entries := runtime.Callers(3, programCounters)

	frames := runtime.CallersFrames(programCounters[:entries])
	frame, _ := frames.Next()

	functionName := frame.Function
	index := strings.LastIndex(functionName, ".")

	if index == -1 {
		return frame.Func.Name(), ""
	}

	return functionName[index+1:], functionName[:index]
}
