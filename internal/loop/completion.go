package loop

import (
	"regexp"
	"strings"
)

func CheckCompletion(output string, promise string) bool {
	// 1. Remove code blocks to avoid false positives in code snippets or examples
	cleanOutput := removeCodeBlocks(output)
	
	// 2. Trim whitespace from both ends
	cleanOutput = strings.TrimSpace(cleanOutput)

	// 3. Look for the promise tag at the very end of the cleaned output
	escaped := escapeRegex(promise)
	// We use $ to ensure it's at the end of the non-code-block text
	pattern := regexp.MustCompile(`(?i)<promise>\s*` + escaped + `\s*</promise>$`)
	
	return pattern.MatchString(cleanOutput)
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
