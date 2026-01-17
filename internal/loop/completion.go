package loop

import (
	"regexp"
	"strings"
)

func CheckCompletion(output string, promise string) bool {
	escaped := escapeRegex(promise)
	pattern := regexp.MustCompile(`(?i)<promise>\s*` + escaped + `\s*</promise>`)
	return pattern.MatchString(output)
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
