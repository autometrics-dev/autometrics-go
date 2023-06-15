package generate // import "github.com/autometrics-dev/autometrics-go/internal/generate"

import (
	"errors"
	"fmt"
	"go/token"
	"strconv"
	"strings"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"

	"github.com/dave/dst"
)

func inspectImportSpec(ctx *internal.GeneratorContext, importSpec *dst.ImportSpec) bool {
	var foundAm bool

	if ctx.Implementation == autometrics.PROMETHEUS {
		if importSpec.Path.Value == AmPromPackage {
			if importSpec.Name != nil {
				ctx.FuncCtx.ImplImportName = importSpec.Name.Name
			} else {
				ctx.FuncCtx.ImplImportName = "autometrics"
			}
			foundAm = true
		}
	}

	if ctx.Implementation == autometrics.OTEL {
		if importSpec.Path.Value == AmOtelPackage {
			if importSpec.Name != nil {
				ctx.FuncCtx.ImplImportName = importSpec.Name.Name
			} else {
				ctx.FuncCtx.ImplImportName = "autometrics"
			}
			foundAm = true
		}
	}

	if importSpec.Name == nil {
		names := strings.Split(importSpec.Path.Value, "/")
		name := strings.Trim(names[len(names)-1], "\"")
		ctx.ImportsMap[name] = strings.Trim(importSpec.Path.Value, "\"")
	} else {
		ctx.ImportsMap[importSpec.Name.Name] = strings.Trim(importSpec.Path.Value, "\"")
	}

	return foundAm
}

func addAutometricsImport(ctx *internal.GeneratorContext, fileTree *dst.File) error {
	// var importSpec dst.ImportSpec
	if ctx.Implementation == autometrics.PROMETHEUS {
		addImport(fileTree, mustUnquote(AmPromPackage))
		ctx.FuncCtx.ImplImportName = "autometrics"
		ctx.ImportsMap["autometrics"] = AmPromPackage

		return nil
	}

	if ctx.Implementation == autometrics.OTEL {
		addImport(fileTree, mustUnquote(AmOtelPackage))
		ctx.FuncCtx.ImplImportName = "autometrics"
		ctx.ImportsMap["autometrics"] = AmOtelPackage

		return nil
	}

	if ctx.FuncCtx.ImplImportName == "" {
		return errors.New("assertion error: ctx.FuncCtx.ImplImportName is empty at the end of addAutometricsImport")
	}

	return errors.New("unrecognized implementation of Autometrics has been queried")
}

// Ref: https://github.com/dave/dst/issues/61#issuecomment-928529830
func addImport(file *dst.File, imp string) {
	// Where to insert our import block within the file's Decl slice
	index := 0

	importSpec := &dst.ImportSpec{
		Path: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", imp)},
	}

	for i, node := range file.Decls {
		n, ok := node.(*dst.GenDecl)
		if !ok {
			continue
		}

		if n.Tok != token.IMPORT {
			continue
		}

		if len(n.Specs) == 1 && mustUnquote(n.Specs[0].(*dst.ImportSpec).Path.Value) == "C" {
			// If we're going to insert, it must be after the "C" import
			index = i + 1
			continue
		}

		// Insert our import into the first non-"C" import block
		for j, spec := range n.Specs {
			path := mustUnquote(spec.(*dst.ImportSpec).Path.Value)
			if !strings.Contains(path, ".") || imp > path {
				continue
			}

			importSpec.Decorations().Before = spec.Decorations().Before
			spec.Decorations().Before = dst.NewLine

			n.Specs = append(n.Specs[:j], append([]dst.Spec{importSpec}, n.Specs[j:]...)...)
			return
		}

		n.Specs = append(n.Specs, importSpec)
		return
	}

	gd := &dst.GenDecl{
		Tok:   token.IMPORT,
		Specs: []dst.Spec{importSpec},
		Decs: dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{Before: dst.EmptyLine, After: dst.EmptyLine},
		},
	}

	file.Decls = append(file.Decls[:index], append([]dst.Decl{gd}, file.Decls[index:]...)...)
}

func mustUnquote(s string) string {
	out, err := strconv.Unquote(s)
	if err != nil {
		panic(err)
	}
	return out
}
