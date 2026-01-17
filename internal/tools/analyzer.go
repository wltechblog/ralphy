package tools

import (
	"strings"
)

func ExtractErrors(output string) []string {
	var errors []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		lower := strings.ToLower(line)
		if isErrorPattern(lower) {
			cleaned := strings.TrimSpace(line)
			if len(cleaned) > 200 {
				cleaned = cleaned[:200]
			}
			if cleaned != "" && !contains(errors, cleaned) {
				errors = append(errors, cleaned)
			}
		}
	}

	if len(errors) > 10 {
		errors = errors[:10]
	}

	return errors
}

func isErrorPattern(line string) bool {
	return strings.Contains(line, "error:") ||
		strings.Contains(line, "failed:") ||
		strings.Contains(line, "exception:") ||
		strings.Contains(line, "typeerror") ||
		strings.Contains(line, "syntaxerror") ||
		strings.Contains(line, "referenceerror") ||
		(strings.Contains(line, "test") && strings.Contains(line, "fail"))
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
