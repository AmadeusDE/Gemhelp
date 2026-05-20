package main

import (
	"strings"
	"testing"
)

func TestFormatMarkdownTerminal(t *testing.T) {
	// Enable terminal forcing for tests
	forceTerminal = true
	defer func() { forceTerminal = false }()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "H1 Header",
			input:    "# Title",
			expected: "\x1b[1;36mTitle\x1b[0m",
		},
		{
			name:     "H2 Header",
			input:    "## Subtitle",
			expected: "\x1b[1;32mSubtitle\x1b[0m",
		},
		{
			name:     "H3 Header",
			input:    "### Sub-subtitle",
			expected: "\x1b[1;33mSub-subtitle\x1b[0m",
		},
		{
			name:     "H4 Header",
			input:    "#### Section",
			expected: "\x1b[1;35mSection\x1b[0m",
		},
		{
			name:     "List item",
			input:    "- bullet",
			expected: "  \x1b[36m•\x1b[0m bullet",
		},
		{
			name:     "Blockquote",
			input:    "> quote",
			expected: "\x1b[90m│\x1b[0m \x1b[3;90mquote\x1b[0m",
		},
		{
			name:     "Bold text",
			input:    "this is **bold** text",
			expected: "this is \x1b[1mbold\x1b[22m text",
		},
		{
			name:     "Italic text",
			input:    "this is *italic* text",
			expected: "this is \x1b[3mitalic\x1b[23m text",
		},
		{
			name:     "Inline code",
			input:    "run `ls -la` now",
			expected: "run \x1b[33mls -la\x1b[0m now",
		},
		{
			name:     "Link formatting",
			input:    "check [Arch Wiki](https://wiki.archlinux.org)",
			expected: "check Arch Wiki (\x1b[4mhttps://wiki.archlinux.org\x1b[24m)",
		},
		{
			name:     "Combination style",
			input:    "> quote with **bold** and `code` inside",
			expected: "\x1b[90m│\x1b[0m \x1b[3;90mquote with \x1b[1mbold\x1b[22m\x1b[3;90m and \x1b[33mcode\x1b[0m\x1b[3;90m inside\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatMarkdown(tt.input)
			if output != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestFormatMarkdownNoTerminal(t *testing.T) {
	forceTerminal = false

	input := "# Heading\n- item\n**bold**"
	output := FormatMarkdown(input)
	if output != input {
		t.Errorf("expected no-op when not in terminal, got differences")
	}
}

func TestFormatMarkdownCodeBlockBorder(t *testing.T) {
	forceTerminal = true
	defer func() { forceTerminal = false }()

	input := "```go\nfmt.Println(\"test\")\n```"
	output := FormatMarkdown(input)

	if !strings.Contains(output, "Code Block") {
		t.Errorf("expected output to contain Code Block border")
	}
	if !strings.Contains(output, "\x1b[32mfmt.Println(\"test\")\x1b[0m") {
		t.Errorf("expected code block line to be styled green")
	}
}
