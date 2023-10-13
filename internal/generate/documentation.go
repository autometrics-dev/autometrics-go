package generate // import "github.com/autometrics-dev/autometrics-go/internal/generate"

import (
	"fmt"
	"math"
	"strings"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"

	"github.com/dave/dst"
)

func cleanUpAutometricsComments(ctx internal.GeneratorContext, funcDeclaration *dst.FuncDecl) ([]string, error) {
	docComments := funcDeclaration.Decorations().Start.All()
	oldStartCommentIndices := autometricsDocStartDirectives(docComments)
	oldEndCommentIndices := autometricsDocEndDirectives(docComments)

	if len(oldStartCommentIndices) > 0 && len(oldEndCommentIndices) == 0 {
		return nil, fmt.Errorf("Found an autometrics:doc-start cookie for function %s, but no matching :doc-end cookie", funcDeclaration.Name.Name)
	}

	if len(oldStartCommentIndices) == 0 && len(oldEndCommentIndices) > 0 {
		return nil, fmt.Errorf("Found an autometrics:doc-end cookie for function %s, but no matching :doc-start cookie", funcDeclaration.Name.Name)
	}

	if len(oldStartCommentIndices) > 1 {
		return nil, fmt.Errorf("Found more than 1 autometrics:doc-start cookie for function %s", funcDeclaration.Name.Name)
	}

	if len(oldEndCommentIndices) > 1 {
		return nil, fmt.Errorf("Found more than 1 autometrics:doc-end cookie for function %s", funcDeclaration.Name.Name)
	}

	if len(oldStartCommentIndices) == 1 && len(oldEndCommentIndices) == 1 {
		oldStartCommentIndex := oldStartCommentIndices[0]
		oldEndCommentIndex := oldEndCommentIndices[0]

		if oldStartCommentIndex >= 0 && oldEndCommentIndex <= oldStartCommentIndex {
			return nil, fmt.Errorf("Found an autometrics cookies for function %s, but the end one is after the start one", funcDeclaration.Name.Name)
		}

		if oldStartCommentIndex >= 0 && oldEndCommentIndex > oldStartCommentIndex {
			// We also remove the header and the footer that are used as block separation
			amCommentSectionStart := int(math.Max(0, float64(oldStartCommentIndex-1)))
			amCommentSectionEnd := int(math.Min(float64(len(docComments)), float64(oldEndCommentIndex+2)))
			docComments = append(docComments[:amCommentSectionStart], docComments[amCommentSectionEnd:]...)

			// Remove the generated links from former passes
			if ctx.DocumentationGenerator != nil {
				generatedLinks := ctx.DocumentationGenerator.GeneratedLinks()
				docComments = filter(docComments, func(input string) bool {
					for _, link := range generatedLinks {
						if strings.Contains(input, fmt.Sprintf("[%s]", link)) {
							return false
						}
					}
					return true
				})
			}
		}
	}

	return docComments, nil
}

// autometricsDocStartDirectives return the list of indices in the array where line is a comment start directive.
func autometricsDocStartDirectives(commentGroup []string) []int {
	var lines []int
	for i, comment := range commentGroup {
		if strings.Contains(comment, "autometrics:doc-start") {
			lines = append(lines, i)
		}
	}

	return lines
}

// autometricsDocStartDirectives return the list of indices in the array where line is a comment end directive.
func autometricsDocEndDirectives(commentGroup []string) []int {
	var lines []int
	for i, comment := range commentGroup {
		if strings.Contains(comment, "autometrics:doc-end") {
			lines = append(lines, i)
		}
	}

	return lines
}

func generateAutometricsComment(ctx internal.GeneratorContext) (commentLines []string) {
	if ctx.DocumentationGenerator == nil {
		return
	}

	l := ctx.DocumentationGenerator.GenerateAutometricsComment(
		ctx,
		ctx.FuncCtx.FunctionName,
		ctx.FuncCtx.ModuleName,
	)
	commentLines = append(commentLines, "//")
	commentLines = append(commentLines, "//\tautometrics:doc-start Generated documentation by Autometrics.")
	commentLines = append(commentLines, "//")
	commentLines = append(commentLines, "// # Autometrics")
	commentLines = append(commentLines, "//")
	commentLines = append(commentLines, l...)
	commentLines = append(commentLines, "//")
	commentLines = append(commentLines, "//\tautometrics:doc-end Generated documentation by Autometrics.")
	commentLines = append(commentLines, "//")

	return
}

func insertComments(inputArray []string, index int, values []string) []string {
	if len(inputArray) == index { // nil or empty slice or after last element
		return append(inputArray, values...)
	}

	beginning := inputArray[:index]
	// Maybe the deep copy is not necessary, wasn't able to
	// specify the semantics properly here.
	end := make([]string, len(inputArray[index:]))
	copy(end, inputArray[index:])

	inputArray = append(beginning, values...)
	inputArray = append(inputArray, end...)

	return inputArray
}
