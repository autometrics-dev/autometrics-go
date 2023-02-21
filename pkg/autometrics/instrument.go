package autometrics

import (
	"runtime"
	"strings"
)

func Instrument() {
	method, module := callerInfo()

	println(method)
	println(module)
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
