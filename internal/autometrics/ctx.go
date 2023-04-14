package autometrics // import "github.com/autometrics-dev/autometrics-go/internal/autometrics"

import (
	"fmt"
	"net/url"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

type GeneratorContext struct {
	RuntimeCtx             autometrics.Context
	FuncCtx                GeneratorFunctionContext
	Implementation         autometrics.Implementation
	DocumentationGenerator AutometricsLinkCommentGenerator
	AllowCustomLatencies   bool
}

type GeneratorFunctionContext struct {
	CommentIndex   int
	FunctionName   string
	ModuleName     string
	ImplImportName string
}

func (c *GeneratorContext) ResetFuncCtx() {
	c.FuncCtx.CommentIndex = -1
	c.FuncCtx.FunctionName = ""
	c.FuncCtx.ModuleName = ""
}

func (c *GeneratorContext) SetCommentIdx(i int) {
	c.FuncCtx.CommentIndex = i
}

func NewGeneratorContext(implementation autometrics.Implementation, prometheusUrl string, allowCustomLatencies bool) (GeneratorContext, error) {
	ctx := GeneratorContext{
		Implementation:       implementation,
		AllowCustomLatencies: allowCustomLatencies,
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
