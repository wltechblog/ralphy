package loop

import (
	"regexp"
	"strings"
)

func CheckCompletion(output string, promise string) bool {
	// 1. Look for the promise tag in the raw output first
	// (AI might wrap the promise in a code block if it wraps the whole response)
	if check(output, promise) {
		return true
	}

	// 2. Remove code blocks and check again (standard case)
	cleanOutput := removeCodeBlocks(output)
	return check(cleanOutput, promise)
}

func check(text string, promise string) bool {
	escaped := regexp.QuoteMeta(promise)
	pattern := regexp.MustCompile(`(?i)<promise>\s*` + escaped + `\s*</promise>`)
	
	matches := pattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return false
	}

	// Get the last occurrence and check if it's near the end
	lastMatch := matches[len(matches)-1]
	remainingText := text[lastMatch[1]:]
	
	return len(strings.TrimSpace(remainingText)) < 500
}

func removeCodeBlocks(s string) string {
	// Matches triple-backtick code blocks and replaces them with spaces to preserve indices if needed
	// but here we just want to ignore their content.
	re := regexp.MustCompile("(?s)```.*?```")
	return re.ReplaceAllString(s, "")
}

func DetectPlaceholderPluginError(output string) bool {
	return strings.Contains(output, "ralph-wiggum is not yet ready for use. This is a placeholder package.")
}
