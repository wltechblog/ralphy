package state

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GetContextPath() (string, error) {
	stateDir, err := GetStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, contextFileName), nil
}

func getContextPath() (string, error) {
	return GetContextPath()
}

func LoadContext() (string, error) {
	contextPath, err := GetContextPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", nil
	}

	return trimmed, nil
}

func SaveContext(context string) error {
	if err := EnsureStateDir(); err != nil {
		return err
	}

	contextPath, err := GetContextPath()
	if err != nil {
		return err
	}

	timestamp := time.Now().Format(time.RFC3339)
	newEntry := "\n## Context added at " + timestamp + "\n" + context + "\n"

	existing, err := LoadContext()
	if err != nil {
		return err
	}

	var content string
	if existing != "" {
		content = existing + newEntry
	} else {
		content = "# Ralph Loop Context\n" + newEntry
	}

	return os.WriteFile(contextPath, []byte(content), 0644)
}

func ClearContext() error {
	contextPath, err := GetContextPath()
	if err != nil {
		return err
	}

	if err := os.Remove(contextPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
