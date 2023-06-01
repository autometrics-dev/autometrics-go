package autometrics

import (
	"reflect"
	"runtime"
	"strings"
)

// CallerInfo returns the (method name, module name) of the function that called the function that called this function.
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
func CallerInfo() (callInfo CallInfo) {
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
		callInfo.FuncName = functionName
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
		callInfo.ParentFuncName = parentFunctionName
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

// ReflectFunctionModuleName takes any function and returns it's name and module split.
//
// There is no `caller` in this context (we just use reflection to extract the information
// from the function pointer), therefore the caller-related fields in the return value are
// empty.
func ReflectFunctionModuleName(f interface{}) (callInfo CallInfo) {
	functionName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	index := strings.LastIndex(functionName, ".")
	if index == -1 {
		callInfo.FuncName = functionName
	} else {
		moduleIndex := strings.LastIndex(functionName[:index], ".")
		if moduleIndex == -1 {
			callInfo.ModuleName = functionName[:index]
		} else {
			callInfo.ModuleName = functionName[moduleIndex+1 : index]
		}
		callInfo.FuncName = functionName[index+1:]
	}

	return callInfo
}
