package generate // import "github.com/autometrics-dev/autometrics-go/internal/generate"

import (
	"fmt"

	"golang.org/x/exp/slices"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"

	"github.com/dave/dst"
)

const (
	deferDecoration = "//autometrics:defer"
)

// injectDeferStatement add all the necessary information into context to produce the correct defer instrumentation statement.
//
// injectDeferStatement is _always_ meant to be called after injectContextStatement, so the injection will always try to happen after the
// statemtent marked with the shadowing marker
func injectDeferStatement(ctx *internal.GeneratorContext, funcDeclaration *dst.FuncDecl) error {
	contextAssignmentIndex, found := findContextStatement(ctx, funcDeclaration)
	if !found {
		return fmt.Errorf("failed to get context: the %v statement is missing", contextDecoration)
	}

	variable, err := errorReturnValueName(funcDeclaration)
	if err != nil {
		return fmt.Errorf("failed to get error return value name: %w", err)
	}

	if len(variable) == 0 {
		variable = "nil"
	} else {
		variable = "&" + variable
	}

	autometricsDeferStatement, err := buildAutometricsDeferStatement(ctx, variable)
	if err != nil {
		return fmt.Errorf("failed to build the defer statement for instrumentation: %w", err)
	}

	funcDeclaration.Body.List = insertStatements(funcDeclaration.Body.List, contextAssignmentIndex+1, []dst.Stmt{&autometricsDeferStatement})
	return nil
}

func insertStatements(inputArray []dst.Stmt, index int, values []dst.Stmt) []dst.Stmt {
	if len(inputArray) == index { // nil or empty slice or after last element
		return append(inputArray, values...)
	}

	beginning := inputArray[:index]
	// Maybe the deep copy is not necessary, wasn't able to
	// specify the semantics properly here.
	end := make([]dst.Stmt, len(inputArray[index:]))
	copy(end, inputArray[index:])

	inputArray = append(beginning, values...)
	inputArray = append(inputArray, end...)

	return inputArray
}

// removeDeferStatement removes, if detected, a previously injected defer statement.
func removeDeferStatement(ctx *internal.GeneratorContext, funcDeclaration *dst.FuncDecl) error {
	for index, statement := range funcDeclaration.Body.List {
		if deferStatement, ok := statement.(*dst.DeferStmt); ok {
			decorations := deferStatement.Decorations().End
			if slices.Contains(decorations.All(), deferDecoration) {
				funcDeclaration.Body.List = append(funcDeclaration.Body.List[:index], funcDeclaration.Body.List[index+1:]...)
				return nil
			}
		}
	}
	return nil
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

// buildAutometricsDeferStatement builds the AST node for the defer instrumentation statement to be inserted.
func buildAutometricsDeferStatement(ctx *internal.GeneratorContext, errorPointerVariable string) (dst.DeferStmt, error) {
	_, contextName, err := buildAutometricsContextNode(ctx)
	if err != nil {
		return dst.DeferStmt{}, fmt.Errorf("could not generate the runtime context value: %w", err)
	}
	statement := dst.DeferStmt{
		Call: &dst.CallExpr{
			Fun: dst.NewIdent(fmt.Sprintf("%vInstrument", autometricsNamespacePrefix(ctx))),
			Args: []dst.Expr{
				dst.NewIdent(contextName),
				dst.NewIdent(errorPointerVariable),
			},
		},
	}

	statement.Decs.Before = dst.NewLine
	statement.Decs.End = []string{deferDecoration}
	statement.Decs.After = dst.EmptyLine

	return statement, nil
}

func autometricsNamespacePrefix(ctx *internal.GeneratorContext) string {
	if ctx.FuncCtx.ImplImportName == "_" {
		return ""
	} else {
		return fmt.Sprintf("%v.", ctx.FuncCtx.ImplImportName)
	}
}
