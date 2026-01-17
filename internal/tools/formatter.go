package tools

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var ansiRegex = regexp.MustCompile(`\x1B\[[0-9;]*m`)

func StripAnsi(input string) string {
	return ansiRegex.ReplaceAllString(input, "")
}

func FormatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}

	totalSeconds := int64(ms / 1000)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func FormatDurationLong(ms int64) string {
	if ms < 0 {
		ms = 0
	}

	totalSeconds := int64(ms / 1000)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func FormatDurationLongFromTime(d time.Duration) string {
	return FormatDurationLong(d.Milliseconds())
}

func FormatToolSummary(toolCounts map[string]int, maxItems int) string {
	if len(toolCounts) == 0 {
		return ""
	}

	type toolCount struct {
		name  string
		count int
	}

	var sorted []toolCount
	for name, count := range toolCounts {
		sorted = append(sorted, toolCount{name, count})
	}

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	shown := maxItems
	if shown > len(sorted) {
		shown = len(sorted)
	}

	var parts []string
	for i := 0; i < shown; i++ {
		parts = append(parts, fmt.Sprintf("%s %d", sorted[i].name, sorted[i].count))
	}

	remaining := len(sorted) - shown
	if remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d more", remaining))
	}

	return strings.Join(parts, " • ")
}

func CollectToolSummaryFromText(text string) map[string]int {
	counts := make(map[string]int)
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		stripped := StripAnsi(line)
		re := regexp.MustCompile(`^\|\s{2}([A-Za-z0-9_-]+)`)
		matches := re.FindStringSubmatch(stripped)
		if len(matches) > 1 {
			tool := matches[1]
			counts[tool]++
		}
	}

	return counts
}
