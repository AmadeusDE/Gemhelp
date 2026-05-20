package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	APIKey   string `json:"api_key"`
	Language string `json:"language"`
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "gemhelp"), nil
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() (*Config, error) {
	// First check environment override
	envKey := os.Getenv("GEMINI_API_KEY")
	if envKey == "" {
		envKey = os.Getenv("GEMHELP_API_KEY")
	}

	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	var cfg Config
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file: %w", err)
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check config file existence: %w", err)
	}

	// Environment overrides
	if envKey != "" {
		cfg.APIKey = envKey
	}

	// Default language if not set
	if cfg.Language == "" {
		lang := os.Getenv("LANG")
		if len(lang) >= 2 {
			cfg.Language = strings.ToLower(lang[:2])
		} else {
			cfg.Language = "en"
		}
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with owner-only read/write permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// PromptConfig runs the first-time setup wizard
func PromptConfig(ctx context.Context) (*Config, error) {
	fmt.Println("==================================================")
	fmt.Println("          Welcome to Gemhelp Setup!               ")
	fmt.Println("==================================================")
	fmt.Println("It looks like you don't have Gemhelp configured yet.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	// 1. Prompt for API Key
	var apiKey string
	for {
		fmt.Print("Please enter your Gemini API Key: ")
		if !scanner.Scan() {
			return nil, errors.New("aborted by user")
		}
		apiKey = strings.TrimSpace(scanner.Text())
		if apiKey == "" {
			fmt.Println("Error: API Key is required.")
			fmt.Println()
			continue
		}

		fmt.Println("Validating API Key...")
		err := ValidateAPIKey(ctx, apiKey)
		if err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			fmt.Println("Please try again.")
			fmt.Println()
			continue
		}
		fmt.Println("API Key validated successfully!")
		break
	}

	// 2. Prompt for Language
	defaultLang := "en"
	langEnv := os.Getenv("LANG")
	if len(langEnv) >= 2 {
		defaultLang = strings.ToLower(langEnv[:2])
	}
	fmt.Printf("Please enter your preferred language code [default: %s]: ", defaultLang)
	if !scanner.Scan() {
		return nil, errors.New("aborted by user")
	}
	lang := strings.TrimSpace(scanner.Text())
	if lang == "" {
		lang = defaultLang
	}

	cfg := &Config{
		APIKey:   apiKey,
		Language: lang,
	}

	if err := SaveConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("Configuration saved successfully to ~/.config/gemhelp/config.json")
	fmt.Println("==================================================")
	fmt.Println()

	return cfg, nil
}
