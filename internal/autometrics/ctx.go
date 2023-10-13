package autometrics // import "github.com/autometrics-dev/autometrics-go/internal/autometrics"

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

// GeneratorContext contains the complete command-line and environment context from the generator invocation.
//
// This context contains all the information necessary to properly process the `autometrics` directives over
// each instrumented function in the file.
type GeneratorContext struct {
	// RuntimeCtx holds the information about the runtime context to build from in the generated code.
	RuntimeCtx RuntimeCtxInfo
	// FuncCtx holds the function specific information for the detected autometrics directive.
	//
	// Notably, it contains all the data relative to the parsing of the arguments in the directive.
	FuncCtx GeneratorFunctionContext
	// Implementation is the metrics library we expect to use in the instrumented code.
	Implementation autometrics.Implementation
	// DocumentationGenerator is the generator to use to generate comments.
	DocumentationGenerator AutometricsLinkCommentGenerator
	// Allow the autometrics directive to have latency targets outside the default buckets.
	AllowCustomLatencies bool
	// Flag to disable/remove the documentation links when calling the generator.
	//
	// This can be set in the command for the generator or through the environment.
	DisableDocGeneration bool
	// Flag to ask the generator to process all function declaration even if they
	// do not have any annotation. The flag is overriden by the RemoveEverything one.
	InstrumentEverything bool
	// Flag to ask the generator to only remove all autometrics generated code in the
	// file.
	RemoveEverything bool
	// ImportMap maps the alias to import in the current file, to canonical names associated with that name.
	ImportsMap map[string]string
}

// This is almost a carbon copy of the autometrics.Context structure, except that
// non-literal types are transcribed to strings to make it possible to reason about
// the runtime context within generator logic.
// Also, all the information that only gets filled at runtime is simply ignored, as
// only statically-known information is useful at this point.
type RuntimeCtxInfo struct {
	// Name of the variable to use as context.Context when building the autometrics.Context.
	// The string will be empty if and only if 'nil' must be used as autometrics.NewContext() argument.
	ContextVariableName string
	// Verbatim code to use to fetch the TraceID.
	// For example, if the instrumented function is detected to use Gin, like
	// `func ginHandler(ginVarName *gin.Context)`
	// then the code in that variable should be something like
	// `"autometrics.DecodeString(ginVarName.GetString(\"TraceID\"))"`
	// This getter should return []byte to allow PreInstrument to
	// log warnings and fill the data manually if the getter returns nil.
	TraceIDGetter        string
	SpanIDGetter         string
	TrackConcurrentCalls bool
	TrackCallerName      bool
	AlertConf            *autometrics.AlertConfiguration
}

func DefaultRuntimeCtxInfo() RuntimeCtxInfo {
	return RuntimeCtxInfo{
		TrackConcurrentCalls: true,
		TrackCallerName:      true,
		ContextVariableName:  "nil",
	}
}

func (c RuntimeCtxInfo) Validate(allowCustomLatencies bool) error {
	if c.AlertConf != nil {
		if c.AlertConf.ServiceName == "" {
			return errors.New("Cannot have an AlertConfiguration without a service name")
		}

		if c.AlertConf.Success != nil && c.AlertConf.Success.Objective <= 0 {
			return errors.New("Cannot have a target success rate that is negative")
		}

		if c.AlertConf.Success != nil && c.AlertConf.Success.Objective <= 1 {
			log.Println("Warning: the target success rate is between 0 and 1, which is between 0 and 1%%. '1' is 1%% not 100%%!")
		}

		if c.AlertConf.Success != nil && c.AlertConf.Success.Objective > 100 {
			return errors.New("Cannot have a target success rate that is strictly greater than 100 (more than 100%)")
		}

		if c.AlertConf.Success != nil && !contains(autometrics.DefObjectives, c.AlertConf.Success.Objective) {
			return fmt.Errorf("Cannot have a target success rate that is not one of the predetermined ones by generated rules files (valid targets are %v)", autometrics.DefObjectives)
		}

		if c.AlertConf.Latency != nil {
			if c.AlertConf.Latency.Objective <= 0 {
				return errors.New("Cannot have a target for latency SLO that is negative")
			}
			if c.AlertConf.Latency.Objective <= 1 {
				log.Println("Warning: the latency target success rate is between 0 and 1, which is between 0 and 1%%. '1' is 1%% not 100%%!")
			}
			if c.AlertConf.Latency.Objective > 100 {
				return errors.New("Cannot have a target for latency SLO that is greater than 100 (more than 100%)")
			}
			if !contains(autometrics.DefObjectives, c.AlertConf.Latency.Objective) {
				return fmt.Errorf("Cannot have a target for latency SLO that is not one of the predetermined in the generated rules files (valid targets are %v)", autometrics.DefObjectives)
			}
			if c.AlertConf.Latency.Target <= 0 {
				return errors.New("Cannot have a target latency SLO threshold that is negative (responses expected before the query)")
			}
			if !allowCustomLatencies && !contains(autometrics.DefBuckets, c.AlertConf.Latency.Target.Seconds()) {
				return fmt.Errorf(
					"Cannot have a target latency SLO threshold that does not match a bucket (valid threshold in seconds are %v). If you set custom latencies in your Init call, then you can add the %v flag to the //go:generate invocation to remove this error",
					autometrics.DefBuckets,
					autometrics.AllowCustomLatenciesFlag,
				)
			}
		}
	}

	return nil
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

type GeneratorFunctionContext struct {
	CommentIndex         int
	FunctionName         string
	ModuleName           string
	ImplImportName       string
	DisableDocGeneration bool
}

func (c *GeneratorContext) ResetFuncCtx() {
	c.FuncCtx.CommentIndex = -1
	c.FuncCtx.FunctionName = ""
	c.FuncCtx.ModuleName = ""
}

func (c *GeneratorContext) SetCommentIdx(i int) {
	c.FuncCtx.CommentIndex = i
}

func NewGeneratorContext(implementation autometrics.Implementation, prometheusUrl string, allowCustomLatencies, disableDocGeneration, instrumentEverything, removeEverything bool) (GeneratorContext, error) {
	ctx := GeneratorContext{
		Implementation:       implementation,
		AllowCustomLatencies: allowCustomLatencies,
		DisableDocGeneration: disableDocGeneration,
		InstrumentEverything: instrumentEverything,
		RemoveEverything:     removeEverything,
		RuntimeCtx:           DefaultRuntimeCtxInfo(),
		FuncCtx:              GeneratorFunctionContext{},
		ImportsMap:           make(map[string]string),
	}

	if prometheusUrl != "" {
		promUrl, err := url.Parse(prometheusUrl)
		if err != nil {
			return ctx, fmt.Errorf("failed to parse prometheus URL: %w", err)
		}

		ctx.DocumentationGenerator = NewPrometheusDoc(*promUrl)
	}

	return ctx, nil
}
