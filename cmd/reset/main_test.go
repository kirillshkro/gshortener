package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go/ast"
)

func TestGenerateResetMethod(t *testing.T) {
	tests := []struct {
		name     string
		structName string
		expected string
	}{
		{
			name:        "simple struct name",
			structName:  "User",
			expected:    "func (s *User) ResetUser()",
		},
		{
			name:        "struct with capital letter",
			structName:  "Config",
			expected:    "func (s *Config) ResetConfig()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateResetMethod(tt.structName, nil)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHasGenerateResetComment(t *testing.T) {
	tests := []struct {
		name     string
		comments string
		expected bool
	}{
		{
			name:     "with generate:reset comment",
			comments: "// generate:reset;",
			expected: true,
		},
		{
			name:     "without generate:reset comment",
			comments: "// some comment",
			expected: false,
		},
		{
			name:     "empty comment",
			comments: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var commentGroup *ast.CommentGroup
			if tt.comments != "" {
				commentGroup = &ast.CommentGroup{}
				commentGroup.List = []*ast.Comment{{Text: tt.comments}}
			}
			result := hasGenerateResetComment(commentGroup)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFindPackages(t *testing.T) {
	tempDir := t.TempDir()
	
	pkgDir := filepath.Join(tempDir, "testpkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	testFile := filepath.Join(pkgDir, "test.go")
	content := `package testpkg

type TestStruct struct {
	Name string
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := `module testmod

go 1.21
`
	if err := os.WriteFile(goModFile, []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	pkgs, err := findPackages(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	found := false
	for _, pkg := range pkgs {
		if strings.Contains(pkg, "testpkg") {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("expected to find testpkg, got %v", pkgs)
	}
}

func TestProcessPackage(t *testing.T) {
	tempDir := t.TempDir()
	
	pkgDir := filepath.Join(tempDir, "testpkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	testFile := filepath.Join(pkgDir, "test.go")
	content := `package testpkg

type (
	// generate:reset;
	User struct {
		Name string
		Age  int
	}
)
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	
	goModFile := filepath.Join(pkgDir, "go.mod")
	goModContent := `module testmod

go 1.21
`
	if err := os.WriteFile(goModFile, []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	err := processPackage(pkgDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	resetFile := filepath.Join(pkgDir, "reset.go")
	if _, err := os.Stat(resetFile); os.IsNotExist(err) {
		t.Error("expected reset.go to be created")
	}
	
	contentBytes, err := os.ReadFile(resetFile)
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}
	
	contentStr := string(contentBytes)
	if !strings.Contains(contentStr, "func (s *User) ResetUser()") {
		t.Errorf("expected to find ResetUser method in reset.go, got: %s", contentStr)
	}
}

func TestWriteResetFile(t *testing.T) {
	tempDir := t.TempDir()
	
	methods := []string{
		"func (s *User) ResetUser() {}\n\n",
		"func (s *Config) ResetConfig() {}\n\n",
	}
	
	err := writeResetFile(tempDir, methods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	resetFile := filepath.Join(tempDir, "reset.go")
	if _, err := os.Stat(resetFile); os.IsNotExist(err) {
		t.Error("expected reset.go to be created")
	}
	
	contentBytes, err := os.ReadFile(resetFile)
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}
	
	contentStr := string(contentBytes)
	if !strings.Contains(contentStr, "func (s *User) ResetUser()") {
		t.Errorf("expected to find User method in reset.go, got: %s", contentStr)
	}
	
	if !strings.Contains(contentStr, "func (s *Config) ResetConfig()") {
		t.Errorf("expected to find Config method in reset.go, got: %s", contentStr)
	}
}
