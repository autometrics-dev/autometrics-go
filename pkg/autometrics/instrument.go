package autometrics

import (
	"context"
	"reflect"
	"runtime"
	"strings"
)

// CallerInfo returns the (method name, module name) of the function that called the function that called this function.
//
// It also returns the information about its autometricized grandparent.
//
// The module name and the parent module names are cropped to their last part, because the generator we use
// only has access to the last "package" name in `GOPACKAGE` environment variable.
//
// If there is a way to obtain programmatically the fully qualified package name in go-generate arguments,
// then we can lift this artificial limitation here and use the full "module name" from the caller information.
// Currently this compromise is the only way to have the documentation links generator creating correct
// queries.
func callerInfo(ctx context.Context) (callInfo CallInfo) {
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
		callInfo.Current.Function = functionName
	} else {
		callInfo.Current.Module = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
			functionName[:index],
			"(", ""),
			")", ""),
			"*", "")

		callInfo.Current.Function = functionName[index+1:]
	}

	parent, err := ParentFunctionName(ctx)

	if err != nil {
		// We try to fallback to the parent in the call stack if we don't have the info
		if !hasParent {
			return
		}

		parentFrame, _ := frames.Next()
		parentFrameFunctionName := parentFrame.Function
		index = strings.LastIndex(parentFrameFunctionName, ".")

		if index == -1 {
			callInfo.Parent.Function = parentFrameFunctionName
		} else {
			callInfo.Parent.Module = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
				parentFrameFunctionName[:index],
				"(", ""),
				")", ""),
				"*", "")

			callInfo.Parent.Function = functionName[index+1:]
		}

		return
	}

	callInfo.Parent = parent

	return
}

// ReflectFunctionModuleName takes any function and returns it's name and module split.
//
// There is no `caller` in this context (we just use reflection to extract the information
// from the function pointer), therefore the caller-related fields in the return value are
// empty.
func ReflectFunctionModuleName(f interface{}) (callInfo CallInfo) {
	functionName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	index := strings.LastIndex(functionName, ".")
	if index == -1 {
		callInfo.Current.Function = functionName
	} else {
		callInfo.Current.Module = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
			functionName[:index],
			"(", ""),
			")", ""),
			"*", "")

		callInfo.Current.Function = functionName[index+1:]
	}

	return callInfo
}
