package loop

import (
	"regexp"
	"strings"
)

func CheckCompletion(output string, promise string) bool {
	// 1. Remove code blocks to avoid false positives in code snippets or examples
	cleanOutput := removeCodeBlocks(output)
	
	// 2. Look for the promise tag
	escaped := escapeRegex(promise)
	pattern := regexp.MustCompile(`(?i)<promise>\s*` + escaped + `\s*</promise>`)
	
	matches := pattern.FindAllStringIndex(cleanOutput, -1)
	if len(matches) == 0 {
		return false
	}

	// 3. Get the last occurrence and check if it's near the end
	// AI sometimes adds closing remarks like "Done!", "I hope this helps", etc.
	// We allow up to 250 characters of trailing text after the promise.
	lastMatch := matches[len(matches)-1]
	remainingText := cleanOutput[lastMatch[1]:]
	
	return len(strings.TrimSpace(remainingText)) < 250
}

func removeCodeBlocks(s string) string {
	// Matches triple-backtick code blocks and replaces them with spaces to preserve indices if needed
	// but here we just want to ignore their content.
	re := regexp.MustCompile("(?s)```.*?```")
	return re.ReplaceAllString(s, "")
}

func escapeRegex(str string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`*`, `\*`,
		`+`, `\+`,
		`?`, `\?`,
		`^`, `\^`,
		`$`, `\$`,
		`.`, `\.`,
		`|`, `\|`,
		`(`, `\(`,
		`)`, `\)`,
		`[`, `\[`,
		`]`, `\]`,
		`{`, `\{`,
		`}`, `\}`,
	)
	return replacer.Replace(str)
}

func DetectPlaceholderPluginError(output string) bool {
	return strings.Contains(output, "ralph-wiggum is not yet ready for use. This is a placeholder package.")
}
