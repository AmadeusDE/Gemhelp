package main

import (
	"os"
	"regexp"
	"strings"
)

var forceTerminal = false

// IsTerminal returns true if stdout is a terminal, NO_COLOR is not set, and TERM is not dumb.
func IsTerminal() bool {
	if forceTerminal {
		return true
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// FormatMarkdown parses basic markdown to ANSI escape codes if outputting to a terminal.
func FormatMarkdown(input string) string {
	if !IsTerminal() {
		return input
	}

	lines := strings.Split(input, "\n")
	var formattedLines []string
	inCodeBlock := false

	// Regular expressions for inline styles
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	reItalic := regexp.MustCompile(`\*(.*?)\*`)
	reCode := regexp.MustCompile("`(.*?)`")
	reLink := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code block boundaries
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				formattedLines = append(formattedLines, "\x1b[90m┌─── Code Block ──────────────────────────────────\x1b[0m")
			} else {
				formattedLines = append(formattedLines, "\x1b[90m└─────────────────────────────────────────────────\x1b[0m")
			}
			continue
		}

		if inCodeBlock {
			// Code content: green text
			formattedLines = append(formattedLines, "\x1b[32m" + line + "\x1b[0m")
			continue
		}

		// Headers
		if strings.HasPrefix(line, "# ") {
			title := line[2:]
			formattedLines = append(formattedLines, "\x1b[1;36m" + title + "\x1b[0m")
			continue
		} else if strings.HasPrefix(line, "## ") {
			title := line[3:]
			formattedLines = append(formattedLines, "\x1b[1;32m" + title + "\x1b[0m")
			continue
		} else if strings.HasPrefix(line, "### ") {
			title := line[4:]
			formattedLines = append(formattedLines, "\x1b[1;33m" + title + "\x1b[0m")
			continue
		} else if strings.HasPrefix(line, "#### ") {
			title := line[5:]
			formattedLines = append(formattedLines, "\x1b[1;35m" + title + "\x1b[0m")
			continue
		}

		// Blockquotes: gray italic text with left-border line
		if strings.HasPrefix(line, "> ") {
			content := line[2:]
			content = applyInlineStyles(content, reBold, reItalic, reCode, reLink, "\x1b[3;90m")
			formattedLines = append(formattedLines, "\x1b[90m│\x1b[0m \x1b[3;90m" + content + "\x1b[0m")
			continue
		}

		// List items: pretty cyan bullet points
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			content := line[2:]
			content = applyInlineStyles(content, reBold, reItalic, reCode, reLink, "")
			formattedLines = append(formattedLines, "  \x1b[36m•\x1b[0m " + content)
			continue
		}

		// Normal lines
		styledLine := applyInlineStyles(line, reBold, reItalic, reCode, reLink, "")
		formattedLines = append(formattedLines, styledLine)
	}

	return strings.Join(formattedLines, "\n")
}

func applyInlineStyles(line string, reBold, reItalic, reCode, reLink *regexp.Regexp, defaultStyle string) string {
	// Links: text (url)
	line = reLink.ReplaceAllString(line, "$1 (\x1b[4m$2\x1b[24m)")

	// Inline code: yellow
	line = reCode.ReplaceAllString(line, "\x1b[33m$1\x1b[0m" + defaultStyle)

	// Bold
	line = reBold.ReplaceAllString(line, "\x1b[1m$1\x1b[22m" + defaultStyle)

	// Italic
	line = reItalic.ReplaceAllString(line, "\x1b[3m$1\x1b[23m" + defaultStyle)

	return line
}
