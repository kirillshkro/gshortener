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

// isVendorDir проверяет, является ли директория vendor-директорией.
func isVendorDir(path string) bool {
	return strings.Contains(path, "/vendor/") || strings.HasSuffix(path, "/vendor")
}

// isGeneratedDir проверяет, является ли директория сгенерированной.
func isGeneratedDir(path string) bool {
	return strings.Contains(path, "/.git")
}

// getPackageName возвращает имя пакета для указанной директории.
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

// generateResetMethod генерирует тело метода Reset() для указанной структуры.
// Файл file используется для получения информации о полях структуры.
func generateResetMethod(structName string, file *ast.File) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "// Reset%s сбрасывает все поля структуры %s к значениям по умолчанию\n", structName, structName)
	fmt.Fprintf(&sb, "func (s *%s) Reset%s() {\n", structName, structName)

	if file != nil {
		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			if decl.Name.Name != structName {
				return true
			}

			structType, ok := decl.Type.(*ast.StructType)
			if !ok {
				return true
			}

			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue
				}

				fieldName := field.Names[0].Name
				fieldValue := generateResetFieldValue(field.Type)
				fmt.Fprintf(&sb, "\ts.%s = %s\n", fieldName, fieldValue)
			}

			return false
		})
	}

	sb.WriteString("}\n\n")

	return sb.String()
}

// generateResetFieldValue генерирует выражение для сброса значения поля к нулевому значению.
func generateResetFieldValue(node ast.Node) string {
	switch v := node.(type) {
	case *ast.Ident:
		return getZeroValue(v.Name)
	case *ast.StarExpr:
		return fmt.Sprintf("&%s{}", getZeroValueExpr(v.X))
	case *ast.ArrayType:
		elemType := generateResetFieldValue(v.Elt)
		return fmt.Sprintf("%s[:0]", elemType)
	case *ast.MapType:
		keyType := generateResetFieldValue(v.Key)
		valType := generateResetFieldValue(v.Value)
		return fmt.Sprintf("map[%s]%s{}", keyType, valType)
	case *ast.SelectorExpr:
		if ident, ok := v.X.(*ast.Ident); ok {
			return fmt.Sprintf("&%s{}", ident.Name)
		}
		return "nil"
	case *ast.ParenExpr:
		return generateResetFieldValue(v.X)
	default:
		return "nil"
	}
}

// getZeroValueExpr извлекает строковое представление типа для создания нулевого значения.
func getZeroValueExpr(node ast.Node) string {
	switch v := node.(type) {
	case *ast.Ident:
		return getZeroValue(v.Name)
	case *ast.StarExpr:
		return getZeroValueExpr(v.X)
	case *ast.SelectorExpr:
		if ident, ok := v.X.(*ast.Ident); ok {
			return ident.Name
		}
		return ""
	default:
		return ""
	}
}

// getZeroValue возвращает строковое представление нулевого значения для примитивного типа.
func getZeroValue(ty string) string {
	switch ty {
	case "string":
		return `""`
	case "bool":
		return "false"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64",
		"complex64", "complex128":
		return "0"
	default:
		return fmt.Sprintf("&%s{}", ty)
	}
}

// writeResetFile записывает все сгенерированные методы в файл reset.gen.go
// в той же директории, где находится пакет.
func writeResetFile(pkgPath string, methods []string) error {
	content := "// Code generated by reset tool. DO NOT EDIT.\n\npackage " + filepath.Base(pkgPath) + "\n\n"

	for _, method := range methods {
		content += method
	}

	resetFilePath := filepath.Join(pkgPath, "reset.gen.go")

	return os.WriteFile(resetFilePath, []byte(content), 0644)
}
