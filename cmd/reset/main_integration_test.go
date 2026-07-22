package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainIntegration(t *testing.T) {
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

	// generate:reset;
	Config struct {
		Timeout int
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

	if !strings.Contains(contentStr, "func (s *Config) ResetConfig()") {
		t.Errorf("expected to find ResetConfig method in reset.go, got: %s", contentStr)
	}

	if !strings.Contains(contentStr, "package testpkg") {
		t.Errorf("expected package declaration in reset.go, got: %s", contentStr)
	}
}
