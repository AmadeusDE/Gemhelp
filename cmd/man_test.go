package main

import (
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
