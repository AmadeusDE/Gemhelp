package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRoff(t *testing.T) {
	input := `.TH TEST "1" "May 2026" "Test Suite"
.SH NAME
test \- a simple test suite utility
.SH SYNOPSIS
.B test
[\fI\,OPTION\/\fR]
.SH DESCRIPTION
Detailed explanation of test.
.PP
Another paragraph with \fBbold\fR and \fIitalic\fR text.
.TP
\fB\-v, \-\-verbose\fP
Enable verbose output.
.TP
\fB\-h, \-\-help\fP
Display help.
.SH SEE ALSO
.BR sh (1),
.BR bash (1)`

	expected := `# TEST(1)

## NAME

test - a simple test suite utility

## SYNOPSIS

**test** [*OPTION*]

## DESCRIPTION

Detailed explanation of test.

Another paragraph with **bold** and *italic* text.

* **-v, --verbose**
  Enable verbose output.

* **-h, --help**
  Display help.

## SEE ALSO

**sh**(1), **bash**(1)`

	output, err := parseRoff(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseRoff returned error: %v", err)
	}

	// Normalize whitespace for easier comparison
	normOutput := normalizeSpaces(output)
	normExpected := normalizeSpaces(expected)

	if normOutput != normExpected {
		t.Errorf("ROFF parser output mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, output)
	}
}

func normalizeSpaces(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}

func TestFindManPagePathOverride(t *testing.T) {
	// Create mock man page directory structure
	tempDir, err := os.MkdirTemp("", "man-test-*")
	if err != nil {
		t.Fatalf("failed to create temp man dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	man1Dir := filepath.Join(tempDir, "man1")
	if err := os.MkdirAll(man1Dir, 0755); err != nil {
		t.Fatalf("failed to create man1 directory: %v", err)
	}

	testFile := filepath.Join(man1Dir, "testcmd.1.gz")
	if err := os.WriteFile(testFile, []byte("mock gzipped content"), 0644); err != nil {
		t.Fatalf("failed to write mock man page: %v", err)
	}

	// Set GEMHELP_MAN_DIR to point to our temp directory
	origManDir := os.Getenv("GEMHELP_MAN_DIR")
	defer os.Setenv("GEMHELP_MAN_DIR", origManDir)
	os.Setenv("GEMHELP_MAN_DIR", tempDir)

	// Call FindManPagePath
	path, exists := FindManPagePath("testcmd")
	if !exists {
		t.Errorf("expected FindManPagePath to find the man page using GEMHELP_MAN_DIR, but it didn't")
	}

	expectedPath := testFile
	if filepath.Clean(path) != filepath.Clean(expectedPath) {
		t.Errorf("expected path '%s', got '%s'", expectedPath, path)
	}

	// Unset GEMHELP_MAN_DIR and ensure it is not found
	os.Setenv("GEMHELP_MAN_DIR", "")
	_, existsEmpty := FindManPagePath("testcmd")
	if existsEmpty {
		t.Errorf("expected FindManPagePath to NOT find 'testcmd' when GEMHELP_MAN_DIR is unset")
	}
}
