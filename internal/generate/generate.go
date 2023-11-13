package generate // import "github.com/autometrics-dev/autometrics-go/internal/generate"

import (
	"errors"
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

	AmPromPackage = "\"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\""
	AmOtelPackage = "\"github.com/autometrics-dev/autometrics-go/otel/autometrics\""
)

type GenerateError struct {
	FunctionName string
	Detail       error
}

type GenerateErrors []GenerateError

func (errs GenerateErrors) Error() string {
	var sb strings.Builder

	for _, err := range errs {
		sb.WriteString(fmt.Sprintf("in %v: %v\n", err.FunctionName, err.Detail.Error()))
	}

	return sb.String()
}

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
		return fmt.Errorf("errors generating instrumentation and documentation: %w", err)
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
		return "", fmt.Errorf("parsing source code: %w", err)
	}

	var inspectErr GenerateErrors
	var foundAmImport bool

	for _, importSpec := range fileTree.Imports {
		foundAmImport = inspectImportSpec(&ctx, importSpec)
		if foundAmImport {
			break
		}
	}

	if !ctx.RemoveEverything && !foundAmImport {
		err = addAutometricsImport(&ctx, fileTree)
		if err != nil {
			return "", fmt.Errorf("adding the autometrics import: %w", err)
		}
	}

	if ctx.FuncCtx.ImplImportName == "" {
		return "", errors.New("assertion error: ctx.FuncCtx.ImplImportName is empty just before filewalking")
	}

	fileWalk := func(node dst.Node) bool {
		if funcDeclaration, ok := node.(*dst.FuncDecl); ok {
			individualError := walkFuncDeclaration(&ctx, funcDeclaration, moduleName)
			if individualError != nil {
				inspectErr = append(inspectErr, *individualError)
			}
		}

		return true
	}

	dst.Inspect(fileTree, fileWalk)

	if inspectErr != nil {
		return "", fmt.Errorf("transforming file in %v: %w", moduleName, inspectErr)
	}

	var buf strings.Builder

	err = decorator.Fprint(&buf, fileTree)
	if err != nil {
		return "", fmt.Errorf("writing the AST to buffer: %w", err)
	}

	return buf.String(), nil
}

// walkFuncDeclaration uses the context to generate documentation and code if necessary for a function declaration in a file.
func walkFuncDeclaration(ctx *internal.GeneratorContext, funcDeclaration *dst.FuncDecl, moduleName string) *GenerateError {
	if !ctx.RemoveEverything && ctx.FuncCtx.ImplImportName == "" {
		if ctx.Implementation == autometrics.PROMETHEUS {
			return &GenerateError{
				FunctionName: funcDeclaration.Name.Name,
				Detail:       fmt.Errorf("the source file is missing a %v import", AmPromPackage),
			}
		} else if ctx.Implementation == autometrics.OTEL {
			return &GenerateError{
				FunctionName: funcDeclaration.Name.Name,
				Detail:       fmt.Errorf("the source file is missing a %v import", AmOtelPackage),
			}
		} else {
			return &GenerateError{
				FunctionName: funcDeclaration.Name.Name,
				Detail:       fmt.Errorf("unknown implementation of metrics has been queried"),
			}
		}
	}

	ctx.FuncCtx.FunctionName = funcDeclaration.Name.Name
	ctx.FuncCtx.ModuleName = moduleName

	defer ctx.ResetFuncCtx()

	// this block gets run for every function in the file
	// Clean up old autometrics comments
	docComments, err := cleanUpAutometricsComments(*ctx, funcDeclaration)
	if err != nil {
		return &GenerateError{
			FunctionName: funcDeclaration.Name.Name,
			Detail:       fmt.Errorf("removing autometrics comment from former pass: %w", err),
		}
	}

	err = removeDeferStatement(ctx, funcDeclaration)
	if err != nil {
		return &GenerateError{
			FunctionName: funcDeclaration.Name.Name,
			Detail: fmt.Errorf(
				"removing an older autometrics defer statement in %v: %w",
				funcDeclaration.Name.Name,
				err),
		}
	}
	err = removeContextStatement(ctx, funcDeclaration)
	if err != nil {
		return &GenerateError{
			FunctionName: funcDeclaration.Name.Name,
			Detail: fmt.Errorf(
				"removing an older autometrics context statement in %v: %w",
				funcDeclaration.Name.Name,
				err),
		}
	}

	// Early exit if we wanted to remove everything
	if ctx.RemoveEverything {
		funcDeclaration.Decorations().Start.Replace(docComments...)
		return nil
	}

	// Detect autometrics directive
	err = parseAutometricsFnContext(ctx, docComments)
	if err != nil {
		return &GenerateError{
			FunctionName: funcDeclaration.Name.Name,
			Detail: fmt.Errorf(
				"parsing //autometrics directive for %v: %w",
				funcDeclaration.Name.Name,
				err),
		}
	}

	// This block only runs on functions that still have the autometrics directive
	listIndex := ctx.FuncCtx.CommentIndex
	if listIndex >= 0 {
		// Insert comments
		if !ctx.DisableDocGeneration && !ctx.FuncCtx.DisableDocGeneration {
			autometricsComment := generateAutometricsComment(*ctx)

			// HACK: gopls will ignore the doc comment completely when requested for documentation if the docComment
			// starts with an indented line. Autometrics currently uses an indented line at the beginning of the
			// generated section to be able to clean up after itself. Therefore gopls will _not_ display anx documentation
			// if the docComment contains only autometrics docs.
			// To fix this issue, if we detect that the autometrics comment would start the function documentation, we
			// artificially insert an extra line that contains only the function name.
			if listIndex == 0 {
				autometricsComment = append([]string{fmt.Sprintf("// %s", funcDeclaration.Name.Name)}, autometricsComment...)
			}

			funcDeclaration.Decorations().Start.Replace(insertComments(docComments, listIndex, autometricsComment)...)
		} else {
			funcDeclaration.Decorations().Start.Replace(docComments...)
		}

		// context statement
		_, err := injectContextStatement(ctx, funcDeclaration)
		if err != nil {
			return &GenerateError{
				FunctionName: funcDeclaration.Name.Name,
				Detail:       fmt.Errorf("injecting context statement: %w", err),
			}
		}

		// defer statement
		err = injectDeferStatement(ctx, funcDeclaration)
		if err != nil {
			return &GenerateError{
				FunctionName: funcDeclaration.Name.Name,
				Detail:       fmt.Errorf("injecting defer statement: %w", err),
			}
		}
	}
	return nil
}

// parseAutometricsFnContext modifies the GeneratorContext according to the arguments put in the directive.
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
				return fmt.Errorf("parsing the directive arguments: %w", err)
			}
			tokenIndex := 0
			for tokenIndex < len(tokens) {
				token := tokens[tokenIndex]
				switch {
				case token == SloNameArgument:
					tokenIndex, err = parseSloName(tokenIndex, tokens, ctx)
					if err != nil {
						return fmt.Errorf("parsing %v argument: %w", SloNameArgument, err)
					}
				case token == SuccessObjArgument:
					tokenIndex, err = parseSuccessObjective(tokenIndex, tokens, ctx)
					if err != nil {
						return fmt.Errorf("parsing %v argument: %w", SuccessObjArgument, err)
					}
				case token == LatencyMsArgument:
					tokenIndex, err = parseLatencyMs(tokenIndex, tokens, ctx)
					if err != nil {
						return fmt.Errorf("parsing %v argument: %w", LatencyMsArgument, err)
					}
				case token == LatencyObjArgument:
					tokenIndex, err = parseLatencyObjective(tokenIndex, tokens, ctx)
					if err != nil {
						return fmt.Errorf("parsing %v argument: %w", LatencyObjArgument, err)
					}
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

	// If the function didn't have any directive, BUT we asked to process all functions, we change the context still
	if ctx.InstrumentEverything {
		ctx.FuncCtx.CommentIndex = len(commentGroup)
		ctx.RuntimeCtx = internal.DefaultRuntimeCtxInfo()
		return nil
	}

	ctx.FuncCtx.CommentIndex = -1
	ctx.RuntimeCtx = internal.DefaultRuntimeCtxInfo()
	return nil
}

func parseLatencyObjective(tokenIndex int, tokens []string, ctx *internal.GeneratorContext) (int, error) {
	if tokenIndex >= len(tokens)-1 {
		return 0, fmt.Errorf("%v argument needs a value", LatencyObjArgument)
	}

	// Read the "value"
	tokenIndex = tokenIndex + 1
	value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
	if err != nil {
		return 0, fmt.Errorf("%v argument must be a float between 0 and 1", LatencyObjArgument)
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
	return tokenIndex, nil
}

func parseLatencyMs(tokenIndex int, tokens []string, ctx *internal.GeneratorContext) (int, error) {
	if tokenIndex >= len(tokens)-1 {
		return 0, fmt.Errorf("%v argument needs a value", LatencyMsArgument)
	}

	// Read the "value"
	tokenIndex = tokenIndex + 1
	value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
	if err != nil {
		return 0, fmt.Errorf("%v argument must be a positive float", LatencyMsArgument)
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
	return tokenIndex, nil
}

func parseSloName(tokenIndex int, tokens []string, ctx *internal.GeneratorContext) (int, error) {
	if tokenIndex >= len(tokens)-1 {
		return 0, fmt.Errorf("%v argument needs a value", SloNameArgument)
	}

	// Read the "value"
	tokenIndex = tokenIndex + 1
	value := tokens[tokenIndex]
	if strings.HasPrefix(value, "--") {
		return 0, fmt.Errorf("%v argument isn't allowed to start with '--'", SloNameArgument)
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
	return tokenIndex, nil
}

func parseSuccessObjective(tokenIndex int, tokens []string, ctx *internal.GeneratorContext) (int, error) {
	if tokenIndex >= len(tokens)-1 {
		return 0, fmt.Errorf("%v argument needs a value", SuccessObjArgument)
	}

	// Read the "value"
	tokenIndex = tokenIndex + 1
	value, err := strconv.ParseFloat(tokens[tokenIndex], 64)
	if err != nil {
		return 0, fmt.Errorf("%v argument must be a float between 0 and 100: %w", SuccessObjArgument, err)
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
	return tokenIndex, nil
}
