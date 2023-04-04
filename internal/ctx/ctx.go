package ctx

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

type AutometricsGeneratorContext struct {
	CommentIndex   int
	Ctx            autometrics.Context
	ImportName     string
	ImplImportName string
}
