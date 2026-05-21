package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndEnvOverrides(t *testing.T) {
	// Create mock config directory
	tempDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp config dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Inject temporary config paths into environment / variables for tests
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tempDir) // redir os.UserHomeDir to tempDir

	// Verify default config creation when file is absent
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load default config: %v", err)
	}

	// Default language fallback test (LANG is empty in test)
	origLang := os.Getenv("LANG")
	defer os.Setenv("LANG", origLang)
	os.Setenv("LANG", "")

	cfg, _ = LoadConfig()
	if cfg.Language != "en" {
		t.Errorf("expected default language 'en', got '%s'", cfg.Language)
	}

	// Test LANG env detection
	os.Setenv("LANG", "fr_FR.UTF-8")
	cfg, _ = LoadConfig()
	if cfg.Language != "fr" {
		t.Errorf("expected language 'fr' from LANG, got '%s'", cfg.Language)
	}

	// Test saving and loading config
	mockConfig := &Config{
		APIKey:   "MOCK_KEY_123",
		Language: "de",
	}

	err = SaveConfig(mockConfig)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.APIKey != "MOCK_KEY_123" {
		t.Errorf("expected API Key 'MOCK_KEY_123', got '%s'", loaded.APIKey)
	}
	if loaded.Language != "de" {
		t.Errorf("expected Language 'de', got '%s'", loaded.Language)
	}

	// Test GEMINI_API_KEY override
	os.Setenv("GEMINI_API_KEY", "OVERRIDDEN_KEY")
	defer os.Setenv("GEMINI_API_KEY", "")

	overridden, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config with overrides: %v", err)
	}

	if overridden.APIKey != "OVERRIDDEN_KEY" {
		t.Errorf("expected overridden API Key, got '%s'", overridden.APIKey)
	}
}

func TestGetConfigPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("skipping test as UserHomeDir is not available")
	}

	expectedPath := filepath.Join(home, ".config", "gemhelp", "config.json")
	actualPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}

	if actualPath != expectedPath {
		t.Errorf("expected config path '%s', got '%s'", expectedPath, actualPath)
	}
}
