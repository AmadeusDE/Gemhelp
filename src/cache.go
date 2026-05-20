package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CacheEntry struct {
	Query    string `json:"query"`
	Response string `json:"response"`
}

func GetCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".cache", "gemhelp", "responses"), nil
}

func GetCacheFilePath(query string) (string, error) {
	dir, err := GetCacheDir()
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(query))
	filename := hex.EncodeToString(hash[:]) + ".json"
	return filepath.Join(dir, filename), nil
}

func GetCachedResponse(query string) (string, bool) {
	path, err := GetCacheFilePath(query)
	if err != nil {
		return "", false
	}

	if _, err := os.Stat(path); err != nil {
		return "", false
	}

	file, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer file.Close()

	var entry CacheEntry
	if err := json.NewDecoder(file).Decode(&entry); err != nil {
		return "", false
	}

	return entry.Response, true
}

func SaveCachedResponse(query string, response string) error {
	dir, err := GetCacheDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	path, err := GetCacheFilePath(query)
	if err != nil {
		return err
	}

	entry := CacheEntry{
		Query:    query,
		Response: response,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

func ClearCache() error {
	dir, err := GetCacheDir()
	if err != nil {
		return err
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove cache directory: %w", err)
	}

	return nil
}
