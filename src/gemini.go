package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
)

func ValidateAPIKey(ctx context.Context, apiKey string) error {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return err
	}

	// Minimal GenerateContent call to test validity of key
	_, err = client.Models.GenerateContent(ctx, "gemini-3.5-flash", []*genai.Content{
		{Parts: []*genai.Part{{Text: "test"}}},
	}, nil)
	return err
}

func callModelWithRetryAndFallback(ctx context.Context, client *genai.Client, history []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	models := []string{"gemini-3.5-flash", "gemini-3.1-flash-lite", "gemini-2.5-flash"}
	var lastErr error

	for _, model := range models {
		backoff := 2 * time.Second
		for attempt := 0; attempt < 3; attempt++ {
			// Rate-limit spacing delay
			time.Sleep(500 * time.Millisecond)

			resp, err := client.Models.GenerateContent(ctx, model, history, config)
			if err == nil {
				return resp, nil
			}

			lastErr = err
			errStr := err.Error()

			// Check for rate limit or resource exhaustion
			if strings.Contains(errStr, "429") || strings.Contains(strings.ToUpper(errStr), "RESOURCE_EXHAUSTED") || strings.Contains(strings.ToUpper(errStr), "RATE_LIMIT") {
				fmt.Printf("Rate limit hit on %s. Retrying in %v...\n", model, backoff)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
					backoff *= 2
					continue
				}
			}

			// For any other error (e.g. invalid model), break loop and go to fallback model
			break
		}

		fmt.Printf("Model %s failed: %v. Falling back to next model...\n", model, lastErr)
	}

	return nil, fmt.Errorf("all models failed: %w", lastErr)
}

func RunGeminiConversation(ctx context.Context, apiKey string, lang string, command string, question string) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return "", err
	}

	systemPrompt := "You are a terminal command help assistant. " +
		"You help the user understand commands and resolve command-line tasks. " +
		"You have access to tools for fetching man pages, TLDR pages, and Arch Wiki content. " +
		"Use these tools when needed to verify information. " +
		"Be concise, clear, and direct. Use Markdown for styling. " +
		"Keep your responses focused and highly practical with clear code block examples."

	systemInst := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: systemPrompt},
		},
	}

	prompt := fmt.Sprintf("Describe the command '%s'.", command)
	if question != "" {
		prompt = fmt.Sprintf("About the command '%s': %s", command, question)
	}

	history := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: prompt},
			},
		},
	}

	tools := []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "get_tldr_page",
					Description: "Gets the TLDR page for a command, providing a quick description and common examples.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"command": {
								Type:        genai.TypeString,
								Description: "The command name (e.g. 'ls', 'grep').",
							},
						},
						Required: []string{"command"},
					},
				},
				{
					Name:        "get_man_page",
					Description: "Gets the local man page content for a command, showing detailed flag definitions and usage.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"command": {
								Type:        genai.TypeString,
								Description: "The command name (e.g. 'ls', 'grep').",
							},
						},
						Required: []string{"command"},
					},
				},
				{
					Name:        "search_arch_wiki",
					Description: "Searches the Arch Wiki for a query and returns a list of matching page titles.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"query": {
								Type:        genai.TypeString,
								Description: "The search query.",
							},
						},
						Required: []string{"query"},
					},
				},
				{
					Name:        "get_arch_wiki_page",
					Description: "Retrieves the content of an Arch Wiki page by its title.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"title": {
								Type:        genai.TypeString,
								Description: "The page title.",
							},
						},
						Required: []string{"title"},
					},
				},
			},
		},
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInst,
		Tools:             tools,
	}

	for loop := 0; loop < 5; loop++ {
		resp, err := callModelWithRetryAndFallback(ctx, client, history, config)
		if err != nil {
			return "", err
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
			return "", fmt.Errorf("received empty response from Gemini")
		}

		modelContent := resp.Candidates[0].Content
		history = append(history, modelContent)

		var hasToolCall bool
		var toolCalls []*genai.FunctionCall
		for _, part := range modelContent.Parts {
			if part.FunctionCall != nil {
				hasToolCall = true
				toolCalls = append(toolCalls, part.FunctionCall)
			}
		}

		if !hasToolCall {
			var finalResponse strings.Builder
			for _, part := range modelContent.Parts {
				if part.Text != "" {
					finalResponse.WriteString(part.Text)
				}
			}
			return finalResponse.String(), nil
		}

		// Parallel execution of all tool calls in the turn
		var wg sync.WaitGroup
		toolResponseParts := make([]*genai.Part, len(toolCalls))
		for idx, fnCall := range toolCalls {
			wg.Add(1)
			go func(i int, fc *genai.FunctionCall) {
				defer wg.Done()
				result := executeToolCall(ctx, fc.Name, fc.Args, lang)
				toolResponseParts[i] = &genai.Part{
					FunctionResponse: &genai.FunctionResponse{
						Name:     fc.Name,
						Response: result,
					},
				}
			}(idx, fnCall)
		}
		wg.Wait()

		// Append the batch of tool response parts back to history
		history = append(history, &genai.Content{
			Role:  "user",
			Parts: toolResponseParts,
		})
	}

	return "", fmt.Errorf("reached maximum tool call loops limit (5)")
}

func executeToolCall(ctx context.Context, name string, args map[string]interface{}, lang string) map[string]interface{} {
	result := make(map[string]interface{})
	switch name {
	case "get_tldr_page":
		cmd, _ := args["command"].(string)
		if cmd == "" {
			result["status"] = "error"
			result["error"] = "missing command parameter"
			return result
		}
		page, err := GetTldrPage(cmd, lang)
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["status"] = "success"
			result["content"] = page
		}

	case "get_man_page":
		cmd, _ := args["command"].(string)
		if cmd == "" {
			result["status"] = "error"
			result["error"] = "missing command parameter"
			return result
		}
		path, exists := FindManPagePath(cmd)
		if !exists {
			result["status"] = "error"
			result["error"] = fmt.Sprintf("man page for '%s' not found locally", cmd)
		} else {
			page, err := ParseManPage(path)
			if err != nil {
				result["status"] = "error"
				result["error"] = err.Error()
			} else {
				result["status"] = "success"
				result["content"] = page
			}
		}

	case "search_arch_wiki":
		query, _ := args["query"].(string)
		if query == "" {
			result["status"] = "error"
			result["error"] = "missing query parameter"
			return result
		}
		titles, err := SearchArchWiki(query)
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["status"] = "success"
			result["results"] = titles
		}

	case "get_arch_wiki_page":
		title, _ := args["title"].(string)
		if title == "" {
			result["status"] = "error"
			result["error"] = "missing title parameter"
			return result
		}
		page, err := GetArchWikiPage(title)
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["status"] = "success"
			result["content"] = page
		}

	default:
		result["status"] = "error"
		result["error"] = fmt.Sprintf("unknown tool name: %s", name)
	}

	return result
}
