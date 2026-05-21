package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/genai"
)

// RunFuckCommand sends a failed command (and optional error output) to Gemini
// with full tool access to diagnose and correct the command.
func RunFuckCommand(ctx context.Context, apiKey string, lang string, failedCommand string, errorOutput string) (correctedCmd string, explanation string, err error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return "", "", err
	}

	systemPrompt := "You are a command correction assistant for Linux/Unix terminals. " +
		"You have access to tools for fetching man pages, TLDR pages, and Arch Wiki content. " +
		"Use these tools to look up the correct syntax, flags, and usage for commands when needed. " +
		"Given a failed shell command and optional error output, diagnose the problem and provide the fix.\n\n" +
		"Your FINAL response (after any tool calls) MUST follow this EXACT format:\n" +
		"Line 1: The corrected command (ONLY the command, no backticks, no markdown, no explanation)\n" +
		"Line 2: Empty line\n" +
		"Line 3+: A brief explanation of what was wrong and what you changed.\n\n" +
		"If the command looks correct and you cannot determine a fix, respond with the original command on line 1 " +
		"and explain that the command appears correct on line 3+."

	systemInst := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: systemPrompt},
		},
	}

	prompt := fmt.Sprintf("Failed command: %s", failedCommand)
	if errorOutput != "" {
		prompt += fmt.Sprintf("\n\nError output:\n%s", errorOutput)
	}

	history := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: prompt},
			},
		},
	}

	tools := buildTools()

	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInst,
		Tools:             tools,
	}

	// Use the same tool-calling loop as RunGeminiConversation
	for loop := 0; loop < 5; loop++ {
		resp, err := callModelWithRetryAndFallback(ctx, client, history, config)
		if err != nil {
			return "", "", err
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
			return "", "", fmt.Errorf("received empty response from Gemini")
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
			var fullResponse strings.Builder
			for _, part := range modelContent.Parts {
				if part.Text != "" {
					fullResponse.WriteString(part.Text)
				}
			}
			return parseFuckResponse(fullResponse.String())
		}

		// Execute tool calls in parallel (reusing the shared helper)
		toolResponseParts := executeToolCallsBatch(ctx, toolCalls, lang)
		history = append(history, &genai.Content{
			Role:  "user",
			Parts: toolResponseParts,
		})
	}

	return "", "", fmt.Errorf("reached maximum tool call loops limit (5)")
}

// parseFuckResponse extracts the corrected command (first non-empty line)
// and the explanation (everything after the first blank line separator).
func parseFuckResponse(response string) (correctedCmd string, explanation string, err error) {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	if len(lines) == 0 {
		return "", "", fmt.Errorf("empty response from model")
	}

	// First non-empty line is the corrected command
	correctedCmd = strings.TrimSpace(lines[0])
	// Strip backticks if the model wrapped it despite instructions
	correctedCmd = strings.Trim(correctedCmd, "`")
	correctedCmd = strings.TrimSpace(correctedCmd)

	if correctedCmd == "" {
		return "", "", fmt.Errorf("could not parse corrected command from response")
	}

	// Everything after the first blank line is the explanation
	var explanationLines []string
	foundBlank := false
	for i := 1; i < len(lines); i++ {
		if !foundBlank && strings.TrimSpace(lines[i]) == "" {
			foundBlank = true
			continue
		}
		if foundBlank {
			explanationLines = append(explanationLines, lines[i])
		}
	}

	explanation = strings.TrimSpace(strings.Join(explanationLines, "\n"))
	return correctedCmd, explanation, nil
}

// readStdinIfPiped reads from stdin if data is being piped (not a terminal).
func readStdinIfPiped() string {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return ""
	}
	if (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return "" // stdin is a terminal, not piped
	}

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return strings.Join(lines, "\n")
}
