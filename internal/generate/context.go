package generate // import "github.com/autometrics-dev/autometrics-go/internal/generate"

import (
	"fmt"
	"go/token"
	"log"
	"reflect"
	"strings"

	"golang.org/x/exp/slices"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"
	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

const (
	vanillaContext       = "context"
	gin                  = "github.com/gin-gonic/gin"
	buffalo              = "github.com/gobuffalo/buffalo"
	echoV4               = "github.com/labstack/echo/v4"
	netHttp              = "net/http"
	amDefaultContextName = "amCtx"
	contextDecoration    = "//autometrics:shadow-ctx"
)

// injectContextStatement add all the necessary information into context to produce the correct context shadowing statement.
func injectContextStatement(ctx *internal.GeneratorContext, funcDeclaration *dst.FuncDecl) (identName string, err error) {
	err = detectContext(ctx, funcDeclaration)
	if err != nil {
		return "", fmt.Errorf("failed to get context for tracing: %w", err)
	}

	autometricsContextStatement, err := buildAutometricsContextStatement(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to build the defer statement for instrumentation: %w", err)
	}

	funcDeclaration.Body.List = append([]dst.Stmt{&autometricsContextStatement}, funcDeclaration.Body.List...)

	// The hard cast is okay here because we just built ourselves the statement.
	return autometricsContextStatement.Lhs[0].(*dst.Ident).Name, nil
}

// removeContextStatement removes, if detected, a previously injected context shadowing statement.
func removeContextStatement(ctx *internal.GeneratorContext, funcDeclaration *dst.FuncDecl) error {
	for index, statement := range funcDeclaration.Body.List {
		if assignStatement, ok := statement.(*dst.AssignStmt); ok {
			decorations := assignStatement.Decorations().End
			if slices.Contains(decorations.All(), contextDecoration) {
				funcDeclaration.Body.List = append(funcDeclaration.Body.List[:index], funcDeclaration.Body.List[index+1:]...)
				return nil
			}
		}
	}
	return nil
}

// findContextStatement finds a previously injected context shadowing statement.
// It also updates the passed context so that the ContextVariableName is accurate.
func findContextStatement(ctx *internal.GeneratorContext, funcDeclaration *dst.FuncDecl) (index int, found bool) {
	for index, statement := range funcDeclaration.Body.List {
		if assignStatement, ok := statement.(*dst.AssignStmt); ok {
			decorations := assignStatement.Decorations().End
			if slices.Contains(decorations.All(), contextDecoration) {
				ctx.RuntimeCtx.ContextVariableName = assignStatement.Lhs[0].(*dst.Ident).Name
				return index, true
			}
		}
	}
	return 0, false
}

// buildAutometricsContextStatement builds the AST node for the context shadowing instrumentation statement to be inserted.
func buildAutometricsContextStatement(ctx *internal.GeneratorContext) (dst.AssignStmt, error) {
	preInstrumentArg, contextShadowName, err := buildAutometricsContextNode(ctx)
	if err != nil {
		return dst.AssignStmt{}, fmt.Errorf("could not generate the runtime context value: %w", err)
	}

	// Crude detection of "should the assignment be `:=` or `=`"
	assignmentToken := token.ASSIGN
	if contextShadowName == amDefaultContextName {
		assignmentToken = token.DEFINE
	}

	// TODO: deal with cases where the context variable name does not exist.
	statement := dst.AssignStmt{
		Lhs: []dst.Expr{
			dst.NewIdent(contextShadowName),
		},
		Tok: assignmentToken,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun: dst.NewIdent(fmt.Sprintf("%vPreInstrument", autometricsNamespacePrefix(ctx))),
				Args: []dst.Expr{
					preInstrumentArg,
				},
			},
		},
	}

	statement.Decs.Before = dst.NewLine
	statement.Decs.End = []string{contextDecoration}
	statement.Decs.After = dst.NewLine

	return statement, nil
}

// buildAutometricsContextNode creates an AST node representing the runtime context to inject in the instrumented code.
//
// This AST node is later used to create the defer statement responsible for instrumenting the code.
func buildAutometricsContextNode(agc *internal.GeneratorContext) (newContextCall *dst.CallExpr, contextShadowName string, err error) {
	// Using https://github.com/dave/dst/issues/73 workaround

	var options []string

	if agc.RuntimeCtx.TraceIDGetter != "" {
		options = append(options, fmt.Sprintf("%vWithTraceID(%v)", autometricsNamespacePrefix(agc), agc.RuntimeCtx.TraceIDGetter))
	}
	if agc.RuntimeCtx.SpanIDGetter != "" {
		options = append(options, fmt.Sprintf("%vWithSpanID(%v)", autometricsNamespacePrefix(agc), agc.RuntimeCtx.SpanIDGetter))
	}

	options = append(options,
		fmt.Sprintf("%vWithConcurrentCalls(%#v)", autometricsNamespacePrefix(agc), agc.RuntimeCtx.TrackConcurrentCalls),
		fmt.Sprintf("%vWithCallerName(%#v)", autometricsNamespacePrefix(agc), agc.RuntimeCtx.TrackCallerName),
	)

	if agc.RuntimeCtx.AlertConf != nil {
		options = append(options, fmt.Sprintf("%vWithSloName(%#v)",
			autometricsNamespacePrefix(agc),
			agc.RuntimeCtx.AlertConf.ServiceName,
		))
		if agc.RuntimeCtx.AlertConf.Latency != nil {
			options = append(options, fmt.Sprintf("%vWithAlertLatency(%#v * time.Nanosecond, %#v)",
				autometricsNamespacePrefix(agc),
				agc.RuntimeCtx.AlertConf.Latency.Target,
				agc.RuntimeCtx.AlertConf.Latency.Objective,
			))
		}
		if agc.RuntimeCtx.AlertConf.Success != nil {
			options = append(options, fmt.Sprintf("%vWithAlertSuccess(%#v)",
				autometricsNamespacePrefix(agc),
				agc.RuntimeCtx.AlertConf.Success.Objective))
		}
	}

	contextShadowName = agc.RuntimeCtx.ContextVariableName
	if contextShadowName == "nil" {
		contextShadowName = amDefaultContextName
	}

	var buf strings.Builder
	_, err = fmt.Fprintf(
		&buf,
		`
package main

var dummy = %vNewContext(
	%s,
`,
		autometricsNamespacePrefix(agc),
		agc.RuntimeCtx.ContextVariableName,
	)
	if err != nil {
		return nil, contextShadowName, fmt.Errorf("could not write string builder to build dummy source code: %w", err)
	}

	for _, o := range options {
		_, err = fmt.Fprintf(&buf, "\t%s,\n", o)
		if err != nil {
			return nil, contextShadowName, fmt.Errorf("could not write string builder to build dummy source code: %w", err)
		}
	}

	_, err = fmt.Fprint(&buf, ")\n")
	if err != nil {
		return nil, contextShadowName, fmt.Errorf("could not write string builder to build dummy source code: %w", err)
	}

	sourceCode := buf.String()
	sourceAst, err := decorator.Parse(sourceCode)
	if err != nil {
		return nil, contextShadowName, fmt.Errorf(
			"could not parse dummy code\n%s\n: %w",
			sourceCode,
			err,
		)
	}

	genDeclNode, ok := sourceAst.Decls[0].(*dst.GenDecl)
	if !ok {
		return nil, contextShadowName, fmt.Errorf("unexpected node in the dummy code (expected dst.GenDecl): %w", err)
	}

	specNode, ok := genDeclNode.Specs[0].(*dst.ValueSpec)
	if !ok {
		return nil, contextShadowName, fmt.Errorf("unexpected node in the dummy code (expected dst.ValueSpec): %w", err)
	}

	callExpr, ok := specNode.Values[0].(*dst.CallExpr)
	if !ok {
		return nil, contextShadowName, fmt.Errorf("unexpected node in the dummy code (expected dst.CallExpr): %w", err)
	}

	return callExpr, contextShadowName, nil
}

// detectContextIdentImpl is a Context detection logic helper for arguments whose type is an identifier
//
// The function returns true when it found enough information to ask for iteration to stop.
func detectContextIdentImpl(ctx *internal.GeneratorContext, argName string, ident *dst.Ident) (bool, error) {
	typeName := ident.Name
	// If argType is just a dst.Ident when parsing, that means
	// it is a single identifier ('Context', _not_ 'context.Context').
	// Therefore we can solely check imports that got imported as '.'
	for alias, canonical := range ctx.ImportsMap {
		if alias != "." {
			continue
		}

		if canonical == vanillaContext && typeName == "Context" {
			ctx.RuntimeCtx.ContextVariableName = argName
			ctx.RuntimeCtx.SpanIDGetter = ""
			ctx.RuntimeCtx.TraceIDGetter = ""
			return true, nil
		}

		if canonical == netHttp && typeName == "Request" {
			if argName == "_" {
				log.Println("Warning: an unnamed net/http.Request has been detected. To make Autometrics reuse its context for tracing purposes, please name it, and run 'go generate' again")
				ctx.RuntimeCtx.ContextVariableName = "nil"
			} else {
				ctx.RuntimeCtx.ContextVariableName = fmt.Sprintf("%s.Context()", argName)
			}
			ctx.RuntimeCtx.SpanIDGetter = ""
			ctx.RuntimeCtx.TraceIDGetter = ""
			return true, nil
		}

		if canonical == gin && typeName == "Context" {
			ctx.RuntimeCtx.SpanIDGetter = fmt.Sprintf("%s.DecodeString(%s.GetString(%#v))", ctx.FuncCtx.ImplImportName, argName, am.MiddlewareSpanIDKey)
			ctx.RuntimeCtx.TraceIDGetter = fmt.Sprintf("%s.DecodeString(%s.GetString(%#v))", ctx.FuncCtx.ImplImportName, argName, am.MiddlewareTraceIDKey)
			return true, nil
		}

		// Buffalo context embeds a context.Context so it can work like vanilla
		if canonical == buffalo && typeName == "Context" {
			ctx.RuntimeCtx.ContextVariableName = argName
			ctx.RuntimeCtx.SpanIDGetter = ""
			ctx.RuntimeCtx.TraceIDGetter = ""
			return true, nil
		}

		if canonical == echoV4 && typeName == "Context" {
			ctx.RuntimeCtx.SpanIDGetter = fmt.Sprintf("%s.DecodeString(%s.Get(%#v))", ctx.FuncCtx.ImplImportName, argName, am.MiddlewareSpanIDKey)
			ctx.RuntimeCtx.TraceIDGetter = fmt.Sprintf("%s.DecodeString(%s.Get(%#v))", ctx.FuncCtx.ImplImportName, argName, am.MiddlewareTraceIDKey)
			return true, nil
		}
	}

	return false, nil
}

// detectContextIdentImpl is a Context detection logic helper for arguments whose type is a selector expression.
//
// The function returns true when it found enough information to ask for iteration to stop.
func detectContextSelectorImpl(ctx *internal.GeneratorContext, argName string, selector *dst.SelectorExpr) (bool, error) {
	typeName := selector.Sel.Name
	if parent, p_ok := selector.X.(*dst.Ident); p_ok {
		parentName := parent.Name
		for alias, canonical := range ctx.ImportsMap {
			if canonical == vanillaContext && parentName == alias && typeName == "Context" {
				ctx.RuntimeCtx.ContextVariableName = argName
				ctx.RuntimeCtx.SpanIDGetter = ""
				ctx.RuntimeCtx.TraceIDGetter = ""
				return true, nil
			}

			if canonical == netHttp && parentName == alias && typeName == "Request" {
				ctx.RuntimeCtx.ContextVariableName = fmt.Sprintf("%s.Context()", argName)
				if argName == "_" {
					log.Println("Warning: an unnamed net/http.Request has been detected. To make Autometrics reuse its context for tracing purposes, please name it, and run 'go generate' again")
					ctx.RuntimeCtx.ContextVariableName = "nil"
				} else {
					ctx.RuntimeCtx.ContextVariableName = fmt.Sprintf("%s.Context()", argName)
				}
				ctx.RuntimeCtx.SpanIDGetter = ""
				ctx.RuntimeCtx.TraceIDGetter = ""
				return true, nil
			}

			if canonical == gin && parentName == alias && typeName == "Context" {
				ctx.RuntimeCtx.SpanIDGetter = fmt.Sprintf("%s.DecodeString(%s.GetString(%#v))", ctx.FuncCtx.ImplImportName, argName, am.MiddlewareSpanIDKey)
				ctx.RuntimeCtx.TraceIDGetter = fmt.Sprintf("%s.DecodeString(%s.GetString(%#v))", ctx.FuncCtx.ImplImportName, argName, am.MiddlewareTraceIDKey)
				return true, nil
			}

			// Buffalo context embeds a context.Context so it can work like vanilla
			if canonical == buffalo && parentName == alias && typeName == "Context" {
				ctx.RuntimeCtx.ContextVariableName = argName
				ctx.RuntimeCtx.SpanIDGetter = ""
				ctx.RuntimeCtx.TraceIDGetter = ""
				return true, nil
			}

			if canonical == echoV4 && typeName == "Context" && (parentName == alias || parentName == "echo") {
				ctx.RuntimeCtx.SpanIDGetter = fmt.Sprintf("%s.DecodeString(%s.Get(%#v))", ctx.FuncCtx.ImplImportName, argName, am.MiddlewareSpanIDKey)
				ctx.RuntimeCtx.TraceIDGetter = fmt.Sprintf("%s.DecodeString(%s.Get(%#v))", ctx.FuncCtx.ImplImportName, argName, am.MiddlewareTraceIDKey)
				return true, nil
			}
		}
	} else {
		// TODO: log that autometrics cannot detect multi-nested contexts instead of errorring
		// continue
		return true, fmt.Errorf("expecting parent to be an identifier, got %s instead", reflect.TypeOf(selector.X).String())
	}
	return false, nil
}

// detectContext modifies a RuntimeCtxInfo to inject context when detected in the function signature.
func detectContext(ctx *internal.GeneratorContext, funcDeclaration *dst.FuncDecl) error {
	arguments := funcDeclaration.Type.Params.List
	for _, argGroup := range arguments {
		if len(argGroup.Names) > 1 {
			continue
		}
		argName := argGroup.Names[0].Name
		if argGroup.Type == nil {
			continue
		}

		if argType, ok := argGroup.Type.(*dst.Ident); ok {
			if found, err := detectContextIdentImpl(ctx, argName, argType); found {
				return err
			}
		} else if argType, ok := argGroup.Type.(*dst.SelectorExpr); ok {
			if found, err := detectContextSelectorImpl(ctx, argName, argType); found {
				return err
			}
		} else if argType, ok := argGroup.Type.(*dst.StarExpr); ok {
			if ident, ok := argType.X.(*dst.Ident); ok {
				if found, err := detectContextIdentImpl(ctx, argName, ident); found {
					return err
				}
			} else if selector, ok := argType.X.(*dst.SelectorExpr); ok {
				if found, err := detectContextSelectorImpl(ctx, argName, selector); found {
					return err
				}
			} else {
				return fmt.Errorf("expecting the type being pointed to to be an identifier, got %s instead", reflect.TypeOf(argType.X).String())
			}
		} else {
			return fmt.Errorf("expecting the type of argGroup to be an identifier, got %s instead", reflect.TypeOf(argGroup.Type).String())
		}
	}

	ctx.RuntimeCtx.ContextVariableName = "nil"
	return nil
}
