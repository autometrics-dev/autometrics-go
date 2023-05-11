package autometrics // import "github.com/autometrics-dev/autometrics-go/internal/autometrics"

import (
	"fmt"
	"net/url"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

// GeneratorContext contains the complete command-line and environment context from the generator invocation.
//
// This context contains all the information necessary to properly process the `autometrics` directives over
// each instrumented function in the file.
type GeneratorContext struct {
	// RuntimeCtx holds the runtime context to build from in the generated code.
	RuntimeCtx autometrics.Context
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
}

type GeneratorFunctionContext struct {
	CommentIndex   int
	FunctionName   string
	ModuleName     string
	ImplImportName string
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

func NewGeneratorContext(implementation autometrics.Implementation, prometheusUrl string, allowCustomLatencies, disableDocGeneration bool) (GeneratorContext, error) {
	ctx := GeneratorContext{
		Implementation:       implementation,
		AllowCustomLatencies: allowCustomLatencies,
		DisableDocGeneration: disableDocGeneration,
		RuntimeCtx:           autometrics.NewContext(),
		FuncCtx:              GeneratorFunctionContext{},
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
