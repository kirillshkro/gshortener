// Package analyzers provide custom analyzer
package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const doc = `checker for direct os.Exit calls in main function

This analyzer detects and reports direct calls to os.Exit inside the main function
of the main package. Using os.Exit in main prevents proper cleanup and defers
from being executed. It's recommended to return error codes or use panic instead.

Examples of code that will be flagged:
    package main
    
    import "os"
    
    func main() {
        os.Exit(1) // This will be reported
    }

Examples of allowed code:
    package main
    
    import "os"
    
    func main() {
        if err := doSomething(); err != nil {
            panic(err) // panic is allowed
        }
        // or return error code using os.Exit in a separate function
    }`

// OsExitAnalyzer custom error analyzer fo os.Exit() statement
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() == "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if fDecl.Name.Name == "main" {
				return true
			}

			if fDecl.Body != nil {
				ast.Inspect(fDecl.Body, func(n ast.Node) bool {
					callExpr, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					ident, ok := selExpr.X.(*ast.Ident)
					if ok {
						if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
							pass.Reportf(selExpr.Pos(), "direct call os.Exit in main function")
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}
