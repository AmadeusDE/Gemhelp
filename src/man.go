package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func FindManPagePath(command string) (string, bool) {
	// Standard search sections
	sections := []string{"1", "8", "6", "2", "3", "4", "5", "7", "1p", "3p"}
	for _, sec := range sections {
		for _, ext := range []string{".gz", ""} {
			path := filepath.Join("/usr/share/man", "man"+sec, command+"."+sec+ext)
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
		}
	}

	// Fallback to globbing in case it's in a non-standard subfolder
	matches, _ := filepath.Glob("/usr/share/man/man*/" + command + ".[1-8]*")
	if len(matches) > 0 {
		return matches[0], true
	}
	return "", false
}

func ParseManPage(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var reader io.Reader = file
	if filepath.Ext(path) == ".gz" {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return "", err
		}
		defer gz.Close()
		reader = gz
	}

	return parseRoff(reader)
}

func parseRoff(r io.Reader) (string, error) {
	var out strings.Builder
	scanner := bufio.NewScanner(r)

	inPreformat := false
	pendingTag := false
	indent := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments
		if strings.HasPrefix(line, `.\"`) || strings.HasPrefix(line, `.'`) {
			continue
		}

		// Handle macro lines
		if strings.HasPrefix(line, ".") {
			// Extract macro command and args
			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}
			macro := parts[0]
			args := parseMacroArgs(line)

			// Trim trailing spaces/newlines from builder only for block-level macros
			isBlockMacro := false
			switch macro {
			case ".TH", ".SH", ".SS", ".PP", ".LP", ".P", ".sp", ".nf", ".fi", ".TP", ".IP", ".RS", ".RE":
				isBlockMacro = true
			}

			if isBlockMacro {
				currText := out.String()
				if len(currText) > 0 && (strings.HasSuffix(currText, " ") || strings.HasSuffix(currText, "\t")) {
					trimmed := strings.TrimRight(currText, " \t")
					out.Reset()
					out.WriteString(trimmed)
				}
			}

			switch macro {
			case ".TH":
				if len(args) >= 2 {
					out.WriteString(fmt.Sprintf("# %s(%s)\n\n", args[0], args[1]))
				}
			case ".SH":
				pendingTag = false
				heading := strings.Join(args, " ")
				out.WriteString(fmt.Sprintf("\n\n## %s\n\n", cleanHeading(heading)))
			case ".SS":
				pendingTag = false
				sub := strings.Join(args, " ")
				out.WriteString(fmt.Sprintf("\n\n### %s\n\n", cleanHeading(sub)))
			case ".PP", ".LP", ".P":
				pendingTag = false
				out.WriteString("\n\n")
			case ".sp":
				out.WriteString("\n\n")
			case ".nf":
				inPreformat = true
				out.WriteString("\n```\n")
			case ".fi":
				inPreformat = false
				out.WriteString("\n```\n")
			case ".RS":
				indent += 2
			case ".RE":
				indent -= 2
				if indent < 0 {
					indent = 0
				}
			case ".TP":
				pendingTag = true
			case ".IP":
				pendingTag = false
				tag := ""
				if len(args) > 0 {
					tag = cleanInlineFormatting(args[0])
				}
				// Ensure preceding blank line
				curr := out.String()
				if len(curr) > 0 && !strings.HasSuffix(curr, "\n\n") {
					if strings.HasSuffix(curr, "\n") {
						out.WriteString("\n")
					} else {
						out.WriteString("\n\n")
					}
				}
				if tag != "" && tag != "•" && tag != "*" {
					out.WriteString(fmt.Sprintf("* **%s** ", tag))
				} else {
					out.WriteString("* ")
				}
			case ".B":
				pendingTag = false
				text := strings.Join(args, " ")
				out.WriteString(fmt.Sprintf("**%s** ", cleanInlineFormatting(text)))
			case ".I":
				pendingTag = false
				text := strings.Join(args, " ")
				out.WriteString(fmt.Sprintf("*%s* ", cleanInlineFormatting(text)))
			case ".BR", ".BI", ".IR", ".IB", ".RB", ".RI":
				pendingTag = false
				formatted := formatAlternating(macro, args)
				out.WriteString(formatted + " ")
			}
			continue
		}

		// Text lines
		cleanedLine := cleanInlineFormatting(stripTtyLinks(line))

		if inPreformat {
			out.WriteString(cleanedLine + "\n")
		} else {
			if pendingTag {
				// Strip outer bold from cleanedLine to prevent double bolding
				tag := cleanedLine
				if strings.HasPrefix(tag, "**") && strings.HasSuffix(tag, "**") && len(tag) >= 4 {
					tag = tag[2 : len(tag)-2]
				}
				// Ensure preceding blank line
				curr := out.String()
				if len(curr) > 0 && !strings.HasSuffix(curr, "\n\n") {
					if strings.HasSuffix(curr, "\n") {
						out.WriteString("\n")
					} else {
						out.WriteString("\n\n")
					}
				}
				out.WriteString(fmt.Sprintf("* **%s**\n  ", tag))
				pendingTag = false
			} else {
				// Indented text block
				if indent > 0 {
					out.WriteString(strings.Repeat(" ", indent) + cleanedLine + "\n")
				} else {
					out.WriteString(cleanedLine + " ")
				}
			}
		}
	}

	if inPreformat {
		out.WriteString("\n```\n")
	}

	result := out.String()
	// Clean up duplicate spaces/newlines
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")
	return strings.TrimSpace(result), scanner.Err()
}

func parseMacroArgs(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false

	runes := []rune(line)
	i := 0
	// Skip macro identifier
	for i < len(runes) && !isWhitespace(runes[i]) {
		i++
	}

	for i < len(runes) {
		r := runes[i]
		if isWhitespace(r) && !inQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else if r == '"' {
			inQuotes = !inQuotes
		} else {
			current.WriteRune(r)
		}
		i++
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func cleanHeading(h string) string {
	h = strings.Trim(h, "\"")
	return strings.ToUpper(h)
}

func formatAlternating(macro string, args []string) string {
	if len(args) == 0 {
		return ""
	}

	var out strings.Builder
	isOdd := true

	for _, arg := range args {
		cleaned := cleanInlineFormatting(arg)
		if cleaned == "" {
			continue
		}

		switch macro {
		case ".BR":
			if isOdd {
				out.WriteString(fmt.Sprintf("**%s**", cleaned))
			} else {
				out.WriteString(cleaned)
			}
		case ".BI":
			if isOdd {
				out.WriteString(fmt.Sprintf("**%s**", cleaned))
			} else {
				out.WriteString(fmt.Sprintf("*%s*", cleaned))
			}
		case ".IR":
			if isOdd {
				out.WriteString(fmt.Sprintf("*%s*", cleaned))
			} else {
				out.WriteString(cleaned)
			}
		case ".IB":
			if isOdd {
				out.WriteString(fmt.Sprintf("*%s*", cleaned))
			} else {
				out.WriteString(fmt.Sprintf("**%s**", cleaned))
			}
		case ".RB":
			if isOdd {
				out.WriteString(cleaned)
			} else {
				out.WriteString(fmt.Sprintf("**%s**", cleaned))
			}
		case ".RI":
			if isOdd {
				out.WriteString(cleaned)
			} else {
				out.WriteString(fmt.Sprintf("*%s*", cleaned))
			}
		}
		isOdd = !isOdd
	}

	return out.String()
}

func cleanInlineFormatting(s string) string {
	// Common ROFF escape cleaning
	s = strings.ReplaceAll(s, `\-`, "-")
	s = strings.ReplaceAll(s, `\&`, "")
	s = strings.ReplaceAll(s, `\e`, `\`)
	s = strings.ReplaceAll(s, `\*(lq`, `"`)
	s = strings.ReplaceAll(s, `\*(rq`, `"`)
	s = strings.ReplaceAll(s, `\(aq`, `'`)
	s = strings.ReplaceAll(s, `\(dq`, `"`)
	s = strings.ReplaceAll(s, `\(em`, " — ")
	s = strings.ReplaceAll(s, `\(en`, "-")
	s = strings.ReplaceAll(s, `\(bu`, "• ")
	s = strings.ReplaceAll(s, `\,`, "")
	s = strings.ReplaceAll(s, `\/`, "")
	s = strings.ReplaceAll(s, `\^`, "")
	s = strings.ReplaceAll(s, `\|`, "")

	// Process inline font escape sequences statefully
	var out strings.Builder
	runes := []rune(s)
	boldOpen := false
	italicOpen := false
	codeOpen := false

	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) && runes[i+1] == 'f' && i+2 < len(runes) {
			fontType := runes[i+2]
			i += 2 // Skip \f and character type

			switch fontType {
			case 'B':
				if !boldOpen {
					out.WriteString("**")
					boldOpen = true
				}
			case 'I':
				if !italicOpen {
					out.WriteString("*")
					italicOpen = true
				}
			case 'C':
				// Handle code fonts like \f(CW
				if i+1 < len(runes) && runes[i+1] == '(' {
					i += 3 // Skip (CW
				}
				if !codeOpen {
					out.WriteString("`")
					codeOpen = true
				}
			case 'R', 'P':
				// Close active markers
				if boldOpen {
					out.WriteString("**")
					boldOpen = false
				}
				if italicOpen {
					out.WriteString("*")
					italicOpen = false
				}
				if codeOpen {
					out.WriteString("`")
					codeOpen = false
				}
			}
		} else {
			out.WriteRune(runes[i])
		}
	}

	// Close open markers
	if boldOpen {
		out.WriteString("**")
	}
	if italicOpen {
		out.WriteString("*")
	}
	if codeOpen {
		out.WriteString("`")
	}

	return out.String()
}

func stripTtyLinks(s string) string {
	// Strip \X'tty: link URL' and \X'tty: link'
	for {
		startIdx := strings.Index(s, "\\X'tty: link")
		if startIdx == -1 {
			break
		}
		endIdx := startIdx + len("\\X'tty: link")
		for endIdx < len(s) && s[endIdx] != '\'' {
			endIdx++
		}
		if endIdx >= len(s) {
			break
		}
		s = s[:startIdx] + s[endIdx+1:]
	}
	s = strings.ReplaceAll(s, "\\X'tty: link'", "")
	return s
}
