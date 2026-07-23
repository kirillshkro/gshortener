// cmd/staticlint/doc.go
/*
Package staticlint implements a multichecker for Go code analysis.

# Overview

This multichecker combines multiple static analysis tools to provide
comprehensive code quality checking for Go projects. It integrates:

- Standard analyzers from golang.org/x/tools/go/analysis/passes
- All SA class analyzers from staticcheck.io
- Selected analyzers from other staticcheck classes (ST, S, QF)
- Public analyzers from the Go ecosystem
- Custom analyzer for os.Exit detection

# Installation

To install the multichecker:

go install github.com/kirillshkro/gshortener/cmd/staticlint@latest

# Usage

Run the multichecker on your Go files:

staticlint ./...

Or on specific packages:

staticlint path/to/your/package

# Configuration

The multichecker accepts standard flags for configuration:

- -tests: include tests (default true)
- -fix: apply suggested fixes
- -json: output in JSON format
- -c: number of parallel workers (default: number of CPUs)

# Analyzers Description

## Standard Analyzers (golang.org/x/tools/go/analysis/passes)

These are the standard static analysis tools from the Go tools repository:

1. asmdecl - checks assembly declarations against Go declarations
2. assign - detects useless assignments
3. atomic - checks for common mistakes with atomic operations
4. bools - detects common mistakes in boolean expressions
5. buildtag - validates build tags
6. cgocall - detects cgo calls in non-Go code
7. composite - checks for unkeyed composite literals
8. copylock - detects copying of lock values
9. errorsas - checks for correct usage of errors.As
10. fieldalignment - detects structs with suboptimal field alignment
11. httpresponse - checks for HTTP response body close issues
12. ifaceassert - detects impossible interface assertions
13. loopclosure - detects loop variable capture issues
14. lostcancel - detects lost cancel functions in contexts
15. nilfunc - detects nil function calls
16. printf - checks printf format strings
17. shadow - detects variable shadowing
18. shift - checks shift operations for correctness
19. sigchanyzer - detects unbuffered signal channels
20. slog - checks for slog usage issues
21. stdmethods - checks standard library method signatures
22. stringintconv - detects string to integer conversions
23. structtag - checks struct tags for correctness
24. testinggoroutine - detects goroutine usage in tests
25. tests - checks test functions for common mistakes
26. timeformat - checks time format strings
27. unmarshal - checks unmarshal usage
28. unreachable - detects unreachable code
29. unsafeptr - checks unsafe.Pointer usage
30. unusedresult - checks for unused function results

## Staticcheck SA Class Analyzers

All SA class analyzers are included. These are the "static analysis" checks
that detect various code issues:

- SA1000: Invalid regular expression
- SA1001: Invalid template
- SA1002: Invalid format in time.Parse
- SA1003: Unsupported type in encoding/binary
- SA1004: Suspiciously small untyped integer in bytes.Equal
- SA1005: Invalid first argument for fmt.Printf
- SA1006: Printf with dynamic first argument and no further arguments
- SA1007: Invalid regular expression in url.Parse
- SA1008: Invalid regexp in http.Request
- SA1010: (*regexp.Regexp).FindAll with negative n
- SA1011: Various array slice allocation issues
- SA1012: nil context in http.Request
- SA1013: HttpRequest with context but no body
- SA1014: Non-pointer argument to unmarshal
- SA1015: Using time.Tick in long-running code
- SA1016: Trapping a signal with a channel
- SA1017: Struct field with tag but not exported
- SA1018: String method with pointer receiver
- SA1019: Deprecated identifier usage
- SA1020: Non-constant format string
- SA1021: Using bytes.Equal on string
- SA1022: Unused context in function
- SA1023: Empty branch
- SA1024: Time since using relative time
- SA1025: Unsupported string conversion
- SA1026: Invalid URL
- SA1027: Invalid timezone
- SA1028: sort.Slice with wrong type
- SA1029: Invalid key type in context.WithValue
- SA1030: Invalid string in strconv

## Other Staticcheck Class Analyzers

Additional analyzers from other Staticcheck classes:

### ST Class (Style) - Code style checks:
- ST1000: Incorrect or missing package comment
- ST1003: Poorly chosen identifier names
- ST1005: Incorrect error message format
- ST1008: Incorrect order of return values
- ST1012: Incorrect error variable names
- ST1016: Methods with same names

### S Class (Simple) - Code simplification:
- S1000: Use copy instead of manual loop
- S1001: Use for range instead of manual loop
- S1002: Use switch instead of if-else chain
- S1004: Use bytes.Equal instead of manual loop
- S1005: Use blank identifier
- S1007: Use strings.ReplaceAll instead of Replace

### QF Class (Quickfix) - Quick fixes:
- QF1001: Convert to interface
- QF1002: Convert to string
- QF1003: Use %v instead of %s
- QF1004: Use strings.ReplaceAll
- QF1005: Use fmt.Errorf

## Public Analyzers

The following public analyzers are included:

1. godox - Detects TODO, FIXME and other comment tags
   - Flags comments with keywords like TODO, FIXME, BUG, etc.
   - Helps track unfinished work and known issues

2. nolintlint - Checks nolint directives
   - Validates that nolint directives are necessary
   - Ensures proper formatting of nolint comments
   - Detects unused nolint directives

## Custom Analyzer

### osexitcheck

This custom analyzer checks for direct calls to os.Exit in the main function
of the main package. It helps prevent issues where defers are not executed
and cleanup is not properly performed.

Why this check is important:
- os.Exit terminates the program immediately
- Defers are NOT executed when os.Exit is called
- Resources may not be properly cleaned up
- It's better to return error codes or use panic

Example of code that would be flagged:
    package main

    import "os"

    func main() {
        os.Exit(1) // This will be reported
    }

Recommended fix: use panic, return with error code, or restructure code
to avoid direct exit.

# Output Format

The multichecker outputs findings in a human-readable format by default.
Use -json flag for JSON output:

staticlint -json ./...

Example JSON output:
{
  "Issues": [
    {
      "Severity": "error",
      "Code": {
        "Value": "S1000"
      },
      "Pos": {
        "Filename": "main.go",
        "Offset": 100,
        "Line": 10,
        "Column": 5
      },
      "Message": "use copy instead of loop"
    }
  ]
}

# Exit Codes

- 0: No issues found
- 1: Issues found or error occurred

# Development

To add new analyzers, modify the main.go file and add them to the allAnalyzers
slice. Make sure to import the appropriate packages.

To test the multichecker:

go test ./cmd/staticlint/...

# Example

Run the linter on your project:

$ staticlint ./...
cmd/main.go:10:2: direct call to os.Exit in main function is forbidden, use return or panic instead (osexitcheck)
pkg/handler.go:25:5: S1000: use copy instead of loop
pkg/service.go:42:1: ST1003: variable 'userId' should be 'userID'

# License

This tool is distributed under the MIT license.
*/
package main
