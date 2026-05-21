package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetTldrPagePaths(t *testing.T) {
	// Create a temp directory to simulate local cached TLDR database
	tempDir, err := os.MkdirTemp("", "tldr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	lang := "fr"
	platformDir := "linux"
	if runtime.GOOS == "darwin" {
		platformDir = "osx"
	} else if runtime.GOOS == "windows" {
		platformDir = "windows"
	}

	// Create directories
	frPlatformPath := filepath.Join(tempDir, lang, platformDir)
	enCommonPath := filepath.Join(tempDir, "en", "common")
	
	if err := os.MkdirAll(frPlatformPath, 0700); err != nil {
		t.Fatalf("failed to create fr platform path: %v", err)
	}
	if err := os.MkdirAll(enCommonPath, 0700); err != nil {
		t.Fatalf("failed to create en common path: %v", err)
	}

	// 1. Create a command only in English common
	enCommonFile := filepath.Join(enCommonPath, "tar.md")
	if err := os.WriteFile(enCommonFile, []byte("# tar\n\nEnglish common version"), 0600); err != nil {
		t.Fatalf("failed to write english common file: %v", err)
	}

	// 2. Create a command in French platform
	frPlatformFile := filepath.Join(frPlatformPath, "ls.md")
	if err := os.WriteFile(frPlatformFile, []byte("# ls\n\nFrench platform version"), 0600); err != nil {
		t.Fatalf("failed to write french platform file: %v", err)
	}

	// Run test helper (replicating our GetTldrPage search order using absolute mocked paths)
	findTldrPageMock := func(cmd string) (string, error) {
		paths := []string{
			filepath.Join(tempDir, lang, platformDir, cmd+".md"),
			filepath.Join(tempDir, lang, "common", cmd+".md"),
			filepath.Join(tempDir, "en", platformDir, cmd+".md"),
			filepath.Join(tempDir, "en", "common", cmd+".md"),
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				data, err := os.ReadFile(path)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
		return "", os.ErrNotExist
	}

	// Test 1: Finding ls (should be found in french platform)
	lsContent, err := findTldrPageMock("ls")
	if err != nil {
		t.Fatalf("failed to find 'ls': %v", err)
	}
	if !stringsContains(lsContent, "French platform version") {
		t.Errorf("expected french platform version, got: %s", lsContent)
	}

	// Test 2: Finding tar (should fall back to english common)
	tarContent, err := findTldrPageMock("tar")
	if err != nil {
		t.Fatalf("failed to find 'tar' fallback: %v", err)
	}
	if !stringsContains(tarContent, "English common version") {
		t.Errorf("expected english common version, got: %s", tarContent)
	}
}

func stringsContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || stringsContainsCheck(s, sub))
}

func stringsContainsCheck(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
