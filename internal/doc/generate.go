package doc

import (
	"github.com/autometrics-dev/autometrics-go/internal/ctx"
)

type AutometricsLinkCommentGenerator interface {
	GenerateAutometricsComment(ctx ctx.AutometricsGeneratorContext, funcName, moduleName string) []string
	GeneratedLinks() []string
}

