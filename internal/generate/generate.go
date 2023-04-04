package generate

import (
	"fmt"
	"go/token"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"

	"golang.org/x/exp/slices"

	"github.com/autometrics-dev/autometrics-go/internal/ctx"
	"github.com/autometrics-dev/autometrics-go/internal/doc"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

const (
	SloNameArgument    = "--slo"
	SuccessObjArgument = "--success-target"
	LatencyMsArgument  = "--latency-ms"
	LatencyObjArgument = "--latency-target"

	AmBasePackage = "\"github.com/autometrics-dev/autometrics-go/pkg/autometrics\""
	AmPromPackage = "\"github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus\""
	AmOtelPackage = "\"github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel\""
)

// TransformFile takes a file path and generates the documentation
// for the `//autometrics:doc` functions.
//
// It also replaces the file in place.
func TransformFile(path, moduleName string, generator doc.AutometricsLinkCommentGenerator, implementation autometrics.Implementation) error {
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

	sourceCode := string(sourceBytes)
	transformedSource, err := GenerateDocumentationAndInstrumentation(sourceCode, moduleName, generator, implementation)
	if err != nil {
		return fmt.Errorf("error generating documentation: %w", err)
	}

	err = os.WriteFile(path, []byte(transformedSource), permissions)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

// GenerateDocumentationAndInstrumentation takes the raw source code from a file and generates
// the documentation for the `//autometrics:doc` functions.
//
// It returns the new source code with augmented documentation.
func GenerateDocumentationAndInstrumentation(sourceCode, moduleName string, generator doc.AutometricsLinkCommentGenerator, implementation autometrics.Implementation) (string, error) {
	fileTree, err := decorator.Parse(sourceCode)
	if err != nil {
		return "", fmt.Errorf("error parsing source code: %w", err)
	}

	var inspectErr error
	var autometricsName string
	var autometricsImplName string

	fileWalk := func(n dst.Node) bool {
		if importSpec, ok := n.(*dst.ImportSpec); ok {
			if importSpec.Path.Value == AmBasePackage {
				if importSpec.Name != nil {
					autometricsName = importSpec.Name.Name
				} else {
					autometricsName = "autometrics"
				}
			}

			if implementation == autometrics.PROMETHEUS {
				if importSpec.Path.Value == AmPromPackage {
					if importSpec.Name != nil {
						autometricsImplName = importSpec.Name.Name
					} else {
						autometricsImplName = "prometheus"
					}
				}
			}

			if implementation == autometrics.OTEL {
				if importSpec.Path.Value == AmOtelPackage {
					if importSpec.Name != nil {
						autometricsImplName = importSpec.Name.Name
					} else {
						autometricsImplName = "otel"
					}
				}
			}

			return true
		}

		if funcDeclaration, ok := n.(*dst.FuncDecl); ok {
			if autometricsImplName == "" {
				if implementation == autometrics.PROMETHEUS {
					inspectErr = fmt.Errorf("the source file is missing a %v import", AmPromPackage)
				} else if implementation == autometrics.OTEL {
					inspectErr = fmt.Errorf("the source file is missing a %v import", AmOtelPackage)
				} else {
					inspectErr = fmt.Errorf("unknown implementation of metrics has been queried.")
				}
				return false
			}

			if autometricsName == "" {
				inspectErr = fmt.Errorf("the source file is missing a %v import", AmBasePackage)
				return false
			}

			// this block gets run for every function in the file
			docComments := funcDeclaration.Decorations().Start.All()

			// Clean up old autometrics comments
			oldStartCommentIndices := autometricsDocStartDirectives(docComments)
			oldEndCommentIndices := autometricsDocEndDirectives(docComments)

			if len(oldStartCommentIndices) > 0 && len(oldEndCommentIndices) == 0 {
				inspectErr = fmt.Errorf("Found an autometrics:doc-start cookie for function %s, but no matching :doc-end cookie", funcDeclaration.Name.Name)
				return false
			}

			if len(oldStartCommentIndices) == 0 && len(oldEndCommentIndices) > 0 {
				inspectErr = fmt.Errorf("Found an autometrics:doc-end cookie for function %s, but no matching :doc-start cookie", funcDeclaration.Name.Name)
				return false
			}

			if len(oldStartCommentIndices) > 1 {
				inspectErr = fmt.Errorf("Found more than 1 autometrics:doc-start cookie for function %s", funcDeclaration.Name.Name)
				return false
			}

			if len(oldEndCommentIndices) > 1 {
				inspectErr = fmt.Errorf("Found more than 1 autometrics:doc-end cookie for function %s", funcDeclaration.Name.Name)
				return false
			}

			if len(oldStartCommentIndices) == 1 && len(oldEndCommentIndices) == 1 {
				oldStartCommentIndex := oldStartCommentIndices[0]
				oldEndCommentIndex := oldEndCommentIndices[0]

				if oldStartCommentIndex >= 0 && oldEndCommentIndex <= oldStartCommentIndex {
					inspectErr = fmt.Errorf("Found an autometrics cookies for function %s, but the end one is after the start one", funcDeclaration.Name.Name)
					return false
				}

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

			}

			// Detect autometrics directive
			generatorCtx, err := parseAutometricsFnContext(docComments)
			generatorCtx.ImportName = autometricsName
			generatorCtx.ImplImportName = autometricsImplName
			if err != nil {
				inspectErr = fmt.Errorf(
					"failed to parse //autometrics directive for %v: %w",
					funcDeclaration.Name.Name,
					err)
				return false
			}
			listIndex := generatorCtx.CommentIndex
			if listIndex >= 0 {
				// Insert comments
				autometricsComment := generateAutometricsComment(generatorCtx, funcDeclaration.Name.Name, moduleName, generator)
				funcDeclaration.Decorations().Start.Replace(insertComments(docComments, listIndex, autometricsComment)...)

				// defer statement
				firstStatement := funcDeclaration.Body.List[0]
				variable, err := errorReturnValueName(funcDeclaration)
				if err != nil {
					inspectErr = fmt.Errorf("failed to get error return value name: %w", err)
					return false
				}

				if len(variable) == 0 {
					variable = "nil"
				} else {
					variable = "&" + variable
				}

				autometricsDeferStatement, err := buildAutometricsDeferStatement(generatorCtx, variable)
				if err != nil {
					inspectErr = fmt.Errorf("failed to build the defer statement for instrumentation: %w", err)
					return false
				}

				if deferStatement, ok := firstStatement.(*dst.DeferStmt); ok {
					decorations := deferStatement.Decorations().End

					if slices.Contains(decorations.All(), "//autometrics:defer") {
						funcDeclaration.Body.List[0] = &autometricsDeferStatement
					} else {
						funcDeclaration.Body.List = append([]dst.Stmt{&autometricsDeferStatement}, funcDeclaration.Body.List...)
					}
				} else {
					funcDeclaration.Body.List = append([]dst.Stmt{&autometricsDeferStatement}, funcDeclaration.Body.List...)
				}
			}
		}

		if inspectErr != nil {
			return false
		}

		return true
	}

	dst.Inspect(fileTree, fileWalk)

	if inspectErr != nil {
		return "", fmt.Errorf("error while transforming file in %v: %w", moduleName, inspectErr)
	}

	var buf strings.Builder

	err = decorator.Fprint(&buf, fileTree)
	if err != nil {
		return "", fmt.Errorf("error writing the AST to buffer: %w", err)
	}

	return buf.String(), nil
}

func buildAutometricsContextNode(agc ctx.AutometricsGeneratorContext) (*dst.CompositeLit, error) {
	// Using https://github.com/dave/dst/issues/73 workaround

	alertConf := "nil"
	alertConfLatency := "nil"
	alertConfSuccess := "nil"

	if agc.Ctx.AlertConf != nil {
		if agc.Ctx.AlertConf.Latency != nil {
			alertConfLatency = fmt.Sprintf("&%v.LatencySlo{ Target: %#v * time.Nanosecond, Objective: %#v }",
				agc.ImportName,
				agc.Ctx.AlertConf.Latency.Target,
				agc.Ctx.AlertConf.Latency.Objective,
			)
		}
		if agc.Ctx.AlertConf.Success != nil {
			alertConfSuccess = fmt.Sprintf("&%v.SuccessSlo{ Objective: %#v }",
				agc.ImportName,
				agc.Ctx.AlertConf.Success.Objective)
		}

		alertConf = fmt.Sprintf("&%v.AlertConfiguration{ ServiceName: %#v, Latency: %v, Success: %v }",
			agc.ImportName,
			agc.Ctx.AlertConf.ServiceName,
			alertConfLatency,
			alertConfSuccess,
		)
	}

	sourceCode := fmt.Sprintf(`
package main

var dummy = %v.Context {
        TrackConcurrentCalls: %#v,
        TrackCallerName: %#v,
        AlertConf: %v,
}
`,
		agc.ImportName,
		agc.Ctx.TrackConcurrentCalls, agc.Ctx.TrackCallerName, alertConf)
	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		return nil, fmt.Errorf("could not parse dummy code: %w", err)
	}

	genDeclNode, ok := sourceAst.Decls[0].(*dst.GenDecl)
	if !ok {
		return nil, fmt.Errorf("unexpected node in the dummy code (expected dst.GenDecl): %w", err)
	}

	specNode, ok := genDeclNode.Specs[0].(*dst.ValueSpec)
	if !ok {
		return nil, fmt.Errorf("unexpected node in the dummy code (expected dst.ValueSpec): %w", err)
	}

	literal, ok := specNode.Values[0].(*dst.CompositeLit)
	if !ok {
		return nil, fmt.Errorf("unexpected node in the dummy code (expected dst.CompositeLit): %w", err)
	}

	return literal, nil
}

// buildAutometricsDeferStatement builds the AST for the defer statement to be inserted.
func buildAutometricsDeferStatement(ctx ctx.AutometricsGeneratorContext, secondVar string) (dst.DeferStmt, error) {
	preInstrumentArg, err := buildAutometricsContextNode(ctx)
	if err != nil {
		return dst.DeferStmt{}, fmt.Errorf("could not generate the runtime context value: %w", err)
	}
	statement := dst.DeferStmt{
		Call: &dst.CallExpr{
			Fun: dst.NewIdent(fmt.Sprintf("%v.Instrument", ctx.ImplImportName)),
			Args: []dst.Expr{
				&dst.CallExpr{
					Fun: dst.NewIdent(fmt.Sprintf("%v.PreInstrument", ctx.ImplImportName)),
					Args: []dst.Expr{
						&dst.UnaryExpr{
							Op:   token.AND,
							X:    preInstrumentArg,
							Decs: dst.UnaryExprDecorations{},
						},
					},
				},
				dst.NewIdent(secondVar),
			},
		},
	}

	statement.Decs.Before = dst.NewLine
	statement.Decs.End = []string{"//autometrics:defer"}
	statement.Decs.After = dst.EmptyLine
	return statement, nil
}

func parseAutometricsFnContext(commentGroup []string) (ctx.AutometricsGeneratorContext, error) {
	for i, comment := range commentGroup {
		if args, found := cutPrefix(comment, "//autometrics:"); found {
			retval := ctx.AutometricsGeneratorContext{
				CommentIndex: i,
				Ctx:          autometrics.NewContext(),
			}
			// TODO: Parse the end of the directive into the autometrics.Context
			tokens, err := shlex.Split(args)
			if err != nil {
				return retval, fmt.Errorf("could not parse the directive arguments: %w", err)
			}
			tokenIndex := 0
			for tokenIndex < len(tokens) {
				token := tokens[tokenIndex]
				switch {
				case token == SloNameArgument:
					if tokenIndex >= len(tokens)-1 {
						return retval, fmt.Errorf("%v argument needs a value", SloNameArgument)
					}
					// Read the "value"
					tokenIndex = tokenIndex + 1
					value := tokens[tokenIndex]
					if strings.HasPrefix(value, "--") {
						return retval, fmt.Errorf("%v argument isn't allowed to start with '--'", SloNameArgument)
					}

					if retval.Ctx.AlertConf != nil {
						retval.Ctx.AlertConf.ServiceName = value
					} else {
						retval.Ctx.AlertConf = &autometrics.AlertConfiguration{
							ServiceName: value,
							Latency:     nil,
							Success:     nil,
						}
					}
					// Advance past the "value"
					tokenIndex = tokenIndex + 1
				case token == SuccessObjArgument:
					if tokenIndex >= len(tokens)-1 {
						return retval, fmt.Errorf("%v argument needs a value", SuccessObjArgument)
					}
					// Read the "value"
					tokenIndex = tokenIndex + 1
					value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
					if err != nil {
						return retval, fmt.Errorf("%v argument must be a float between 0 and 100: %w", SuccessObjArgument, err)
					}

					if retval.Ctx.AlertConf != nil {
						if retval.Ctx.AlertConf.Success != nil {
							retval.Ctx.AlertConf.Success.Objective = value
						} else {
							retval.Ctx.AlertConf.Success = &autometrics.SuccessSlo{Objective: value}
						}
					} else {
						retval.Ctx.AlertConf = &autometrics.AlertConfiguration{
							ServiceName: "",
							Latency:     nil,
							Success:     &autometrics.SuccessSlo{Objective: value},
						}
					}
					// Advance past the "value"
					tokenIndex = tokenIndex + 1
				case token == LatencyMsArgument:
					if tokenIndex >= len(tokens)-1 {
						return retval, fmt.Errorf("%v argument needs a value", LatencyMsArgument)
					}
					// Read the "value"
					tokenIndex = tokenIndex + 1
					value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
					if err != nil {
						return retval, fmt.Errorf("%v argument must be a positive float", LatencyMsArgument)
					}
					timeValue := time.Duration(value * float64(time.Millisecond))

					if retval.Ctx.AlertConf != nil {
						if retval.Ctx.AlertConf.Latency != nil {
							retval.Ctx.AlertConf.Latency.Target = timeValue
						} else {
							retval.Ctx.AlertConf.Latency = &autometrics.LatencySlo{
								Target:    timeValue,
								Objective: 0,
							}
						}
					} else {
						retval.Ctx.AlertConf = &autometrics.AlertConfiguration{
							ServiceName: "",
							Latency: &autometrics.LatencySlo{
								Target:    timeValue,
								Objective: 0,
							},
							Success: nil,
						}
					}
					// Advance past the "value"
					tokenIndex = tokenIndex + 1
				case token == LatencyObjArgument:
					if tokenIndex >= len(tokens)-1 {
						return retval, fmt.Errorf("%v argument needs a value", LatencyObjArgument)
					}
					// Read the "value"
					tokenIndex = tokenIndex + 1
					value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
					if err != nil {
						return retval, fmt.Errorf("%v argument must be a float between 0 and 1", LatencyObjArgument)
					}

					if retval.Ctx.AlertConf != nil {
						if retval.Ctx.AlertConf.Latency != nil {
							retval.Ctx.AlertConf.Latency.Objective = value
						} else {
							retval.Ctx.AlertConf.Latency = &autometrics.LatencySlo{
								Target:    0,
								Objective: value,
							}
						}
					} else {
						retval.Ctx.AlertConf = &autometrics.AlertConfiguration{
							ServiceName: "",
							Latency: &autometrics.LatencySlo{
								Target:    0,
								Objective: value,
							},
							Success: nil,
						}
					}
					// Advance past the "value"
					tokenIndex = tokenIndex + 1
				default:
					// Advance past the "value"
					tokenIndex = tokenIndex + 1
				}
			}
			err = retval.Ctx.Validate()
			if err != nil {
				return retval, fmt.Errorf("parsed configuration is invalid: %w", err)
			}

			return retval, nil
		}
	}

	return ctx.AutometricsGeneratorContext{
		CommentIndex: -1,
		Ctx:          autometrics.NewContext(),
	}, nil
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

func generateAutometricsComment(generatorCtx ctx.AutometricsGeneratorContext, funcName, moduleName string, generator doc.AutometricsLinkCommentGenerator) []string {
	var ret []string
	ret = append(ret, "//")
	ret = append(ret, "//   autometrics:doc-start DO NOT EDIT HERE AND LINE ABOVE")
	ret = append(ret, "//")
	ret = append(ret, "// # Autometrics")
	ret = append(ret, "//")
	ret = append(ret, generator.GenerateAutometricsComment(generatorCtx, funcName, moduleName)...)
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

// Backport of strings.CutPrefix for pre-1.20
func cutPrefix(s, prefix string) (after string, found bool) {
	if !strings.HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}
