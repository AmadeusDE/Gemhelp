package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetTldrDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "gemhelp", "tldr"), nil
}

func EnsureTldrPages(lang string) error {
	dir, err := GetTldrDir()
	if err != nil {
		return err
	}

	// Always ensure English is available as it's the global fallback
	enPath := filepath.Join(dir, "en")
	if _, err := os.Stat(enPath); os.IsNotExist(err) {
		fmt.Println("Downloading English TLDR pages fallback...")
		if err := downloadAndExtractTldr("en"); err != nil {
			return fmt.Errorf("failed to download English TLDR: %w", err)
		}
	}

	// Download target language if it's not English
	if lang != "en" {
		langPath := filepath.Join(dir, lang)
		if _, err := os.Stat(langPath); os.IsNotExist(err) {
			fmt.Printf("Downloading %s TLDR pages...\n", lang)
			if err := downloadAndExtractTldr(lang); err != nil {
				// Don't error out completely if a specific language zip isn't available
				// just print warning and fall back to English
				fmt.Printf("Warning: Could not fetch TLDR pages for language '%s' (%v). Falling back to English.\n", lang, err)
			}
		}
	}

	return nil
}

func downloadAndExtractTldr(lang string) error {
	tldrDir, err := GetTldrDir()
	if err != nil {
		return err
	}

	targetDir := filepath.Join(tldrDir, lang)
	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	url := fmt.Sprintf("https://github.com/tldr-pages/tldr/releases/latest/download/tldr-pages.%s.zip", lang)
	
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch zip archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad http response status: %s", resp.Status)
	}

	// Save zip to temporary file
	tmpFile, err := os.CreateTemp("", "tldr-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write zip content: %w", err)
	}

	// Open ZIP
	archive, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer archive.Close()

	for _, file := range archive.File {
		// Clean the path to avoid directory traversal
		cleanedPath := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanedPath, "..") || strings.Contains(cleanedPath, "/") && strings.HasPrefix(filepath.Base(cleanedPath), ".") {
			continue
		}

		// Only extract markdown documentation files and directories
		isDoc := strings.HasSuffix(cleanedPath, ".md") || file.FileInfo().IsDir()
		if !isDoc {
			continue
		}

		destPath := filepath.Join(targetDir, cleanedPath)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0700); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0700); err != nil {
			return err
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return err
		}

		if _, err := io.Copy(destFile, srcFile); err != nil {
			destFile.Close()
			srcFile.Close()
			return err
		}

		destFile.Close()
		srcFile.Close()
	}

	return nil
}

func GetTldrPage(command, lang string) (string, error) {
	if err := EnsureTldrPages(lang); err != nil {
		return "", err
	}

	tldrDir, err := GetTldrDir()
	if err != nil {
		return "", err
	}

	platformDir := "linux"
	switch runtime.GOOS {
	case "darwin":
		platformDir = "osx"
	case "windows":
		platformDir = "windows"
	}

	// 1. Look in preferred language
	paths := []string{
		filepath.Join(tldrDir, lang, platformDir, command+".md"),
		filepath.Join(tldrDir, lang, "common", command+".md"),
	}

	// 2. Look in English if not English
	if lang != "en" {
		paths = append(paths,
			filepath.Join(tldrDir, "en", platformDir, command+".md"),
			filepath.Join(tldrDir, "en", "common", command+".md"),
		)
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

	return "", fmt.Errorf("tldr page for '%s' not found locally", command)
}
