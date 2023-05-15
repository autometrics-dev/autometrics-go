package generate // import "github.com/autometrics-dev/autometrics-go/internal/generate"

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

const (
	SloNameArgument    = "--slo"
	SuccessObjArgument = "--success-target"
	LatencyMsArgument  = "--latency-ms"
	LatencyObjArgument = "--latency-target"
	NoDocArgument      = "--no-doc"

	AmPromPackage = "\"github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus\""
	AmOtelPackage = "\"github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel\""
)

// TransformFile takes a file path and generates the documentation
// for the `//autometrics:inst` functions.
//
// It also replaces the file in place.
func TransformFile(ctx internal.GeneratorContext, path, moduleName string) error {
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
	transformedSource, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, moduleName)
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
// the documentation for the `//autometrics:inst` functions.
//
// It returns the new source code with augmented documentation.
func GenerateDocumentationAndInstrumentation(ctx internal.GeneratorContext, sourceCode, moduleName string) (string, error) {
	fileTree, err := decorator.Parse(sourceCode)
	if err != nil {
		return "", fmt.Errorf("error parsing source code: %w", err)
	}

	var inspectErr error

	fileWalk := func(n dst.Node) bool {
		if importSpec, ok := n.(*dst.ImportSpec); ok {
			if ctx.Implementation == autometrics.PROMETHEUS {
				if importSpec.Path.Value == AmPromPackage {
					if importSpec.Name != nil {
						ctx.FuncCtx.ImplImportName = importSpec.Name.Name
					} else {
						ctx.FuncCtx.ImplImportName = "prometheus"
					}
				}
			}

			if ctx.Implementation == autometrics.OTEL {
				if importSpec.Path.Value == AmOtelPackage {
					if importSpec.Name != nil {
						ctx.FuncCtx.ImplImportName = importSpec.Name.Name
					} else {
						ctx.FuncCtx.ImplImportName = "otel"
					}
				}
			}

			if importSpec.Name == nil {
				names := strings.Split(importSpec.Path.Value, "/")
				name := strings.Trim(names[len(names)-1], "\"")
				ctx.ImportsMap[name] = strings.Trim(importSpec.Path.Value, "\"")
			} else {
				ctx.ImportsMap[importSpec.Name.Name] = strings.Trim(importSpec.Path.Value, "\"")
			}

			return true
		}

		if funcDeclaration, ok := n.(*dst.FuncDecl); ok {
			if ctx.FuncCtx.ImplImportName == "" {
				if ctx.Implementation == autometrics.PROMETHEUS {
					inspectErr = fmt.Errorf("the source file is missing a %v import", AmPromPackage)
				} else if ctx.Implementation == autometrics.OTEL {
					inspectErr = fmt.Errorf("the source file is missing a %v import", AmOtelPackage)
				} else {
					inspectErr = fmt.Errorf("unknown implementation of metrics has been queried.")
				}
				return false
			}

			ctx.FuncCtx.FunctionName = funcDeclaration.Name.Name
			ctx.FuncCtx.ModuleName = moduleName
			defer ctx.ResetFuncCtx()

			// this block gets run for every function in the file
			// Clean up old autometrics comments
			docComments, err := cleanUpAutometricsComments(ctx, funcDeclaration)
			if err != nil {
				inspectErr = fmt.Errorf("error trying to remove autometrics comment from former pass: %w", err)
				return false
			}

			// TODO: clean up the defer statement inconditionally here as well, if detected

			// Detect autometrics directive
			err = parseAutometricsFnContext(&ctx, docComments)
			if err != nil {
				inspectErr = fmt.Errorf(
					"failed to parse //autometrics directive for %v: %w",
					funcDeclaration.Name.Name,
					err)
				return false
			}

			// This block only runs on functions that still have the autometrics directive
			listIndex := ctx.FuncCtx.CommentIndex
			if listIndex >= 0 {
				// Insert comments
				if !ctx.DisableDocGeneration && !ctx.FuncCtx.DisableDocGeneration {
					autometricsComment := generateAutometricsComment(ctx)
					funcDeclaration.Decorations().Start.Replace(insertComments(docComments, listIndex, autometricsComment)...)
				} else {
					funcDeclaration.Decorations().Start.Replace(docComments...)
				}

				// defer statement
				err := injectDeferStatement(&ctx, funcDeclaration)
				if err != nil {
					inspectErr = fmt.Errorf("failed to inject defer statement: %w", err)
					return false
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

func parseAutometricsFnContext(ctx *internal.GeneratorContext, commentGroup []string) error {
	for i, comment := range commentGroup {
		if args, found := cutPrefix(comment, "//autometrics:"); found {
			if !strings.Contains(comment, "autometrics:doc") && !strings.Contains(comment, "autometrics:inst") {
				return fmt.Errorf("invalid directive comment '%s': only '//autometrics:doc' and '//autometrics:inst' are allowed.", comment)
			}
			ctx.FuncCtx.CommentIndex = i
			ctx.RuntimeCtx = internal.DefaultRuntimeCtxInfo()

			tokens, err := shlex.Split(args)
			if err != nil {
				return fmt.Errorf("could not parse the directive arguments: %w", err)
			}
			tokenIndex := 0
			for tokenIndex < len(tokens) {
				token := tokens[tokenIndex]
				switch {
				case token == SloNameArgument:
					if tokenIndex >= len(tokens)-1 {
						return fmt.Errorf("%v argument needs a value", SloNameArgument)
					}
					// Read the "value"
					tokenIndex = tokenIndex + 1
					value := tokens[tokenIndex]
					if strings.HasPrefix(value, "--") {
						return fmt.Errorf("%v argument isn't allowed to start with '--'", SloNameArgument)
					}

					if ctx.RuntimeCtx.AlertConf != nil {
						ctx.RuntimeCtx.AlertConf.ServiceName = value
					} else {
						ctx.RuntimeCtx.AlertConf = &autometrics.AlertConfiguration{
							ServiceName: value,
							Latency:     nil,
							Success:     nil,
						}
					}
					// Advance past the "value"
					tokenIndex = tokenIndex + 1
				case token == SuccessObjArgument:
					if tokenIndex >= len(tokens)-1 {
						return fmt.Errorf("%v argument needs a value", SuccessObjArgument)
					}
					// Read the "value"
					tokenIndex = tokenIndex + 1
					value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
					if err != nil {
						return fmt.Errorf("%v argument must be a float between 0 and 100: %w", SuccessObjArgument, err)
					}

					if ctx.RuntimeCtx.AlertConf != nil {
						if ctx.RuntimeCtx.AlertConf.Success != nil {
							ctx.RuntimeCtx.AlertConf.Success.Objective = value
						} else {
							ctx.RuntimeCtx.AlertConf.Success = &autometrics.SuccessSlo{Objective: value}
						}
					} else {
						ctx.RuntimeCtx.AlertConf = &autometrics.AlertConfiguration{
							ServiceName: "",
							Latency:     nil,
							Success:     &autometrics.SuccessSlo{Objective: value},
						}
					}
					// Advance past the "value"
					tokenIndex = tokenIndex + 1
				case token == LatencyMsArgument:
					if tokenIndex >= len(tokens)-1 {
						return fmt.Errorf("%v argument needs a value", LatencyMsArgument)
					}
					// Read the "value"
					tokenIndex = tokenIndex + 1
					value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
					if err != nil {
						return fmt.Errorf("%v argument must be a positive float", LatencyMsArgument)
					}
					timeValue := time.Duration(value * float64(time.Millisecond))

					if ctx.RuntimeCtx.AlertConf != nil {
						if ctx.RuntimeCtx.AlertConf.Latency != nil {
							ctx.RuntimeCtx.AlertConf.Latency.Target = timeValue
						} else {
							ctx.RuntimeCtx.AlertConf.Latency = &autometrics.LatencySlo{
								Target:    timeValue,
								Objective: 0,
							}
						}
					} else {
						ctx.RuntimeCtx.AlertConf = &autometrics.AlertConfiguration{
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
						return fmt.Errorf("%v argument needs a value", LatencyObjArgument)
					}
					// Read the "value"
					tokenIndex = tokenIndex + 1
					value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
					if err != nil {
						return fmt.Errorf("%v argument must be a float between 0 and 1", LatencyObjArgument)
					}

					if ctx.RuntimeCtx.AlertConf != nil {
						if ctx.RuntimeCtx.AlertConf.Latency != nil {
							ctx.RuntimeCtx.AlertConf.Latency.Objective = value
						} else {
							ctx.RuntimeCtx.AlertConf.Latency = &autometrics.LatencySlo{
								Target:    0,
								Objective: value,
							}
						}
					} else {
						ctx.RuntimeCtx.AlertConf = &autometrics.AlertConfiguration{
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
				case token == NoDocArgument:
					ctx.FuncCtx.DisableDocGeneration = true
					tokenIndex = tokenIndex + 1
				default:
					// Advance past the "value"
					tokenIndex = tokenIndex + 1
				}
			}
			err = ctx.RuntimeCtx.Validate(ctx.AllowCustomLatencies)
			if err != nil {
				return fmt.Errorf("parsed configuration is invalid: %w", err)
			}

			return nil
		}
	}

	ctx.FuncCtx.CommentIndex = -1
	ctx.RuntimeCtx = internal.DefaultRuntimeCtxInfo()
	return nil
}
