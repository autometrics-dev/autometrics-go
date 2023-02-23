package doc

import (
	"fmt"
	"os"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

type AutometricsLinkCommentGenerator interface {
	GenerateAutometricsComment(funcName, moduleName string) []string
	GeneratedLinks() []string
}

// TransformFile takes a file path and generates the documentation
// for the `//autometrics:doc` functions.
//
// It also replaces the file in place.
func TransformFile(path, moduleName string, generator AutometricsLinkCommentGenerator) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting a working directory: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error reading file information from %s: %w", path, err)
	}

	permissions := info.Mode()

	sourceBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading the source code from %s (cwd: %s): %w", path, cwd, err)
	}

	err = os.WriteFile(fmt.Sprintf("%s.bak", path), sourceBytes, permissions)
	if err != nil {
		return fmt.Errorf("error writing backup file: %w", err)
	}

	sourceCode := string(sourceBytes)

	transformedSource, err := GenerateDocumentation(sourceCode, moduleName, generator)
	if err != nil {
		return fmt.Errorf("error generating documentation: %w", err)
	}

	err = os.WriteFile(path, []byte(transformedSource), permissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// GenerateDocumentation takes the raw source code from a file and generates
// the documentation for the `//autometrics:doc` functions.
//
// It returns the new source code with augmented documentation.
func GenerateDocumentation(sourceCode, moduleName string, generator AutometricsLinkCommentGenerator) (string, error) {
	fileTree, err := decorator.Parse(sourceCode)
	if err != nil {
		return "", fmt.Errorf("error parsing source code: %w", err)
	}

	dst.Inspect(fileTree, func(n dst.Node) bool {
		if funcDeclaration, ok := n.(*dst.FuncDecl); ok {
			docComments := funcDeclaration.Decorations().Start.All()

			// Clean up old autometrics comments
			oldStartCommentIndex := autometricsDocStartDirective(docComments)
			oldEndCommentIndex := autometricsDocEndDirective(docComments)
			// TODO: error out if there is:
			// - a start but no end
			// - an end but no start
			// - an end before a start
			// - multiple starts
			// - multiple ends
			if oldStartCommentIndex >= 0 && oldEndCommentIndex > oldStartCommentIndex {
				// We also remove the header and the footer that are used as block separation
				docComments = append(docComments[:oldStartCommentIndex-1], docComments[oldEndCommentIndex+2:]...)

				// Remove the generated links from former passes
				generatedLinks := generator.GeneratedLinks()
				docComments = filter(docComments, func(input string) bool {
					for _, link := range generatedLinks {
						if strings.Contains(input, fmt.Sprintf("[%s]", link)) {
							return false
						}
					}
					return true
				})
			}

			// Insert new autometrics comment
			listIndex := hasAutometricsDocDirective(docComments)
			if listIndex >= 0 {
				autometricsComment := generateAutometricsComment(funcDeclaration.Name.Name, moduleName, generator)
				funcDeclaration.Decorations().Start.Replace(insertComments(docComments, listIndex, autometricsComment)...)
			}
		}

		return true
	})

	var buf strings.Builder

	err = decorator.Fprint(&buf, fileTree)
	if err != nil {
		return "", fmt.Errorf("error writing the AST to buffer: %w", err)
	}

	return buf.String(), nil
}

func hasAutometricsDocDirective(commentGroup []string) int {
	for i, comment := range commentGroup {
		if comment == "//autometrics:doc" {
			return i
		}
	}

	return -1
}

func autometricsDocStartDirective(commentGroup []string) int {
	for i, comment := range commentGroup {
		if strings.Contains(comment, "autometrics:doc-start") {
			return i
		}
	}

	return -1
}

func autometricsDocEndDirective(commentGroup []string) int {
	for i, comment := range commentGroup {
		if strings.Contains(comment, "autometrics:doc-end") {
			return i
		}
	}

	return -1
}

func generateAutometricsComment(funcName, moduleName string, generator AutometricsLinkCommentGenerator) []string {
	var ret []string
	ret = append(ret, "//")
	ret = append(ret, "//   autometrics:doc-start DO NOT EDIT HERE AND LINE ABOVE")
	ret = append(ret, "//")
	ret = append(ret, "// # Autometrics")
	ret = append(ret, "//")
	ret = append(ret, generator.GenerateAutometricsComment(funcName, moduleName)...)
	ret = append(ret, "//")
	ret = append(ret, "//   autometrics:doc-end DO NOT EDIT HERE AND LINE BELOW")
	ret = append(ret, "//")

	return ret
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

// errorReturnValueName returns the name of the error return value if it exists.
func errorReturnValueName(funcNode *dst.FuncDecl) (string, error) {
	returnValues := funcNode.Type.Results
	if returnValues == nil || returnValues.List == nil {
		return "", nil
	}

	for _, field := range returnValues.List {
		fieldType := field.Type
		if spec, ok := fieldType.(*dst.Ident); ok {
			if spec.Name == "error" {
				// Assuming that the `error` type has 0 or 1 name before it.
				if field.Names == nil {
					return "", nil
				} else if len(field.Names) > 1 {
					return "", fmt.Errorf("expecting a single named `error` return value, got %d instead.", len(field.Names))
				}
				return field.Names[0].Name, nil
			}
		}
	}

	return "", nil
}

func filter(ss []string, test func(string) bool) (ret []string) {
    for _, s := range ss {
        if test(s) {
            ret = append(ret, s)
        }
    }
    return
}
