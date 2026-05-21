package main

import (
	"testing"
)

func TestParseFuckResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		expectedCmd string
		expectedExp string
		expectErr   bool
	}{
		{
			name: "Standard response",
			response: `ls -al

The -al flag is used to list all files in long format.`,
			expectedCmd: "ls -al",
			expectedExp: "The -al flag is used to list all files in long format.",
			expectErr:   false,
		},
		{
			name: "Response with backticks",
			response: "```ls -al```\n\nExplanation here.",
			expectedCmd: "ls -al",
			expectedExp: "Explanation here.",
			expectErr:   false,
		},
		{
			name: "Response with single backticks",
			response: "`ls -al` \n\nExplanation here.",
			expectedCmd: "ls -al",
			expectedExp: "Explanation here.",
			expectErr:   false,
		},
		{
			name:        "Empty response",
			response:    "",
			expectedCmd: "",
			expectedExp: "",
			expectErr:   true,
		},
		{
			name: "No explanation",
			response: "ls -al",
			expectedCmd: "ls -al",
			expectedExp: "",
			expectErr:   false,
		},
		{
			name: "Multiple blank lines",
			response: "ls -al\n\n\nExplanation with\nmultiple lines.",
			expectedCmd: "ls -al",
			expectedExp: "Explanation with\nmultiple lines.",
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, exp, err := parseFuckResponse(tt.response)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseFuckResponse() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if cmd != tt.expectedCmd {
				t.Errorf("parseFuckResponse() gotCmd = %v, want %v", cmd, tt.expectedCmd)
			}
			if exp != tt.expectedExp {
				t.Errorf("parseFuckResponse() gotExp = %v, want %v", exp, tt.expectedExp)
			}
		})
	}
}
