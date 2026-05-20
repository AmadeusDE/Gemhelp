package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type WikiSearchResponse struct {
	Query struct {
		Search []struct {
			Title string `json:"title"`
		} `json:"search"`
	} `json:"query"`
}

type WikiParseResponse struct {
	Parse struct {
		Title    string `json:"title"`
		Wikitext struct {
			Content string `json:"*"`
		} `json:"wikitext"`
	} `json:"parse"`
}

func SearchArchWiki(query string) ([]string, error) {
	apiURL := "https://wiki.archlinux.org/api.php?action=query&list=search&srsearch=" +
		url.QueryEscape(query) + "&format=json"

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arch wiki api returned status %s", resp.Status)
	}

	var data WikiSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var titles []string
	for _, item := range data.Query.Search {
		titles = append(titles, item.Title)
	}

	return titles, nil
}

func GetArchWikiPage(title string) (string, error) {
	apiURL := "https://wiki.archlinux.org/api.php?action=parse&page=" +
		url.QueryEscape(title) + "&prop=wikitext&format=json"

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("arch wiki api returned status %s", resp.Status)
	}

	var data WikiParseResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	cleanedContent := cleanWikitext(data.Parse.Wikitext.Content)
	return fmt.Sprintf("# Arch Wiki: %s\n\n%s", data.Parse.Title, cleanedContent), nil
}

func cleanWikitext(w string) string {
	// Replaces bold and italics
	w = regexp.MustCompile(`'''(.*?)'''`).ReplaceAllString(w, "**$1**")
	w = regexp.MustCompile(`''(.*?)''`).ReplaceAllString(w, "*$1*")

	// Headings
	w = regexp.MustCompile(`(?m)^=====(.*?)=====`).ReplaceAllString(w, "#### $1")
	w = regexp.MustCompile(`(?m)^====(.*?)====`).ReplaceAllString(w, "#### $1")
	w = regexp.MustCompile(`(?m)^===(.*?)===`).ReplaceAllString(w, "### $1")
	w = regexp.MustCompile(`(?m)^==(.*?)==`).ReplaceAllString(w, "## $1")

	// MediaWiki Links [[Target|Text]] or [[Target]]
	w = regexp.MustCompile(`\[\[([^|\]]+)\|([^\]]+)\]\]`).ReplaceAllString(w, "$2")
	w = regexp.MustCompile(`\[\[([^\]]+)\]\]`).ReplaceAllString(w, "$1")
	
	// External links [http://url text]
	w = regexp.MustCompile(`\[https?://\S+\s+([^\]]+)\]`).ReplaceAllString(w, "$1")

	// Clean templates statefully starting from inside to handle nesting
	w = cleanTemplates(w)

	// HTML comments
	w = regexp.MustCompile(`(?s)<!--.*?-->`).ReplaceAllString(w, "")

	// HTML tags (basic sanitization)
	w = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(w, "")

	// Formatting entities
	w = strings.ReplaceAll(w, "&lt;", "<")
	w = strings.ReplaceAll(w, "&gt;", ">")
	w = strings.ReplaceAll(w, "&amp;", "&")

	// Excess empty lines
	w = regexp.MustCompile(`\n{3,}`).ReplaceAllString(w, "\n\n")

	return strings.TrimSpace(w)
}

func cleanTemplates(s string) string {
	for {
		start := strings.LastIndex(s, "{{")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "}}")
		if end == -1 {
			break
		}
		end += start

		templateContent := s[start+2 : end]
		replacement := ""

		parts := strings.Split(templateContent, "|")
		if len(parts) > 0 {
			tmplName := strings.ToLower(strings.TrimSpace(parts[0]))
			switch tmplName {
			case "ic": // Inline code
				if len(parts) > 1 {
					replacement = "`" + parts[1] + "`"
				}
			case "hc": // Host code
				if len(parts) > 1 {
					code := parts[1]
					if len(parts) > 2 {
						code = strings.Join(parts[1:], "\n")
					}
					replacement = "\n```\n" + code + "\n```\n"
				}
			case "note", "tip", "warning", "important":
				if len(parts) > 1 {
					text := ""
					for _, p := range parts[1:] {
						if !strings.Contains(p, "=") {
							text = p
							break
						} else if strings.HasPrefix(strings.TrimSpace(p), "text=") {
							text = strings.TrimPrefix(strings.TrimSpace(p), "text=")
							break
						}
					}
					if text == "" {
						text = parts[1]
					}
					replacement = fmt.Sprintf("\n> **%s**: %s\n", strings.ToUpper(tmplName), text)
				}
			default:
				// Strip translation matrices, language headers, etc.
				replacement = ""
			}
		}

		s = s[:start] + replacement + s[end+2:]
	}
	return s
}
