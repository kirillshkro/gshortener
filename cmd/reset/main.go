package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: reset <root-dir>")
		return
	}
	rootDir := os.Args[1]

	pkgs, err := findPackages(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при сканировании пакетов: %v\n", err)
		os.Exit(1)
	}

	for _, pkgPath := range pkgs {
		if err := processPackage(pkgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при обработке пакета %s: %v\n", pkgPath, err)
			os.Exit(1)
		}
	}
}

func findPackages(rootDir string) ([]string, error) {
	var packages []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		if isVendorDir(path) || isGeneratedDir(path) {
			return filepath.SkipDir
		}

		goFiles, err := filepath.Glob(filepath.Join(path, "*.go"))
		if err != nil {
			return nil
		}

		if len(goFiles) > 0 {
			pkgName, err := getPackageName(path)
			if err != nil {
				return nil
			}
			if pkgName != "" && pkgName != "main" {
				packages = append(packages, path)
			}
		}

		return nil
	})

	return packages, err
}

func isVendorDir(path string) bool {
	return strings.Contains(path, "/vendor/") || strings.HasSuffix(path, "/vendor")
}

func isGeneratedDir(path string) bool {
	return strings.Contains(path, "/.gigacode") || strings.Contains(path, "/.opencode") ||
		strings.Contains(path, "/.git")
}

func getPackageName(dir string) (string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	for _, pkg := range pkgs {
		return pkg.Name, nil
	}

	return "", nil
}

func processPackage(pkgPath string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("ошибка парсинга: %w", err)
	}

	var generatedMethods []string

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				decl, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				if _, ok := decl.Type.(*ast.StructType); ok {
					if hasGenerateResetComment(decl.Doc) {
						method := generateResetMethod(decl.Name.Name, file)
						generatedMethods = append(generatedMethods, method)
					}
				}

				return true
			})
		}
	}

	if len(generatedMethods) > 0 {
		return writeResetFile(pkgPath, generatedMethods)
	}

	return nil
}

func hasGenerateResetComment(comments *ast.CommentGroup) bool {
	if comments == nil {
		return false
	}

	for _, comment := range comments.List {
		if strings.TrimSpace(comment.Text) == "// generate:reset;" {
			return true
		}
	}

	return false
}
