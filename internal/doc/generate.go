package doc

type AutometricsLinkCommentGenerator interface {
	GenerateAutometricsComment(funcName, moduleName string) []string
	GeneratedLinks() []string
}

