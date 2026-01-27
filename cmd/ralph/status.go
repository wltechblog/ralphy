package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wltechblog/ralphy/internal/state"
	"github.com/wltechblog/ralphy/internal/tools"
)

func printStatus() {
	s, err := state.LoadState()
	if err != nil {
		fmt.Println("⏹️  No active loop")
		return
	}

	h, err := state.LoadHistory()
	if err != nil {
		h = &state.RalphHistory{}
	}

	ctx, _ := state.LoadContext()
	
	progressFile := "PROGRESS.md"
	var progress string
	if data, err := os.ReadFile(progressFile); err == nil {
		progress = string(data)
	}

	fmt.Println(`
╔══════════════════════════════════════════════════════════════════╗
║                    Ralph Wiggum Status                           ║
╚══════════════════════════════════════════════════════════════════╝`)

	if s != nil && s.Active {
		startedAt, _ := time.Parse(time.RFC3339, s.StartedAt)
		elapsed := time.Since(startedAt)
		fmt.Println("🔄 ACTIVE LOOP")
		fmt.Printf("   Iteration:    %d", s.Iteration)
		if s.MaxIterations > 0 {
			fmt.Printf(" / %d", s.MaxIterations)
		}
		fmt.Println(" (unlimited)")
		fmt.Printf("   Started:      %s\n", s.StartedAt)
		fmt.Printf("   Elapsed:      %s\n", tools.FormatDurationLong(elapsed.Milliseconds()))
		fmt.Printf("   Promise:      %s\n", s.CompletionPromise)
		fmt.Printf("   Task Promise: %s\n", s.TaskPromise)
		if s.Model != "" {
			fmt.Printf("   Model:        %s\n", s.Model)
		}
		preview := truncate(s.Prompt, 60)
		fmt.Printf("   Prompt:       %s%s\n", preview, ellipsis(s.Prompt, 60))
	} else {
		fmt.Println("⏹️  No active loop")
	}

	tasks, _, err := state.LoadTasks()
	if err == nil && len(tasks) > 0 {
		fmt.Println("\n📋 CURRENT TASKS:")
		for i, task := range tasks {
			statusIcon := "⏸️"
			if task.Status == "complete" {
				statusIcon = "✅"
			} else if task.Status == "in-progress" {
				statusIcon = "🔄"
			}
			fmt.Printf("   %d. %s %s\n", i+1, statusIcon, task.Text)

			for _, subtask := range task.Subtasks {
				subStatusIcon := "⏸️"
				if subtask.Status == "complete" {
					subStatusIcon = "✅"
				} else if subtask.Status == "in-progress" {
					subStatusIcon = "🔄"
				}
				fmt.Printf("      %s %s\n", subStatusIcon, subtask.Text)
			}
		}
		
		completeCount := 0
		inProgressCount := 0
		for _, t := range tasks {
			if t.Status == "complete" {
				completeCount++
			} else if t.Status == "in-progress" {
				inProgressCount++
			}
		}
		fmt.Printf("\n   Progress: %d/%d complete, %d in progress\n", completeCount, len(tasks), inProgressCount)
	}

	if progress != "" {
		fmt.Println("\n📈 AGENT PROGRESS (from PROGRESS.md):")
		// Indent the progress lines
		lines := strings.Split(strings.TrimSpace(progress), "\n")
		for i, line := range lines {
			if i > 10 {
				fmt.Printf("   ... (%d more lines)\n", len(lines)-10)
				break
			}
			fmt.Printf("   %s\n", line)
		}
	}

	if ctx != "" {
		fmt.Println("\n📝 PENDING CONTEXT (will be injected next iteration):")
		lines := fmt.Sprintf("   %s", ctx)
		fmt.Println(lines)
	}

	if len(h.Iterations) > 0 {
		fmt.Printf("\n📊 HISTORY (%d iterations)\n", len(h.Iterations))
		fmt.Printf("   Total time:   %s\n", tools.FormatDurationLong(h.TotalDurationMs))

		recent := h.Iterations
		if len(recent) > 5 {
			recent = recent[len(recent)-5:]
		}
		fmt.Println("\n   Recent iterations:")
		for _, iter := range recent {
			toolsSummary := tools.FormatToolSummary(iter.ToolsUsed, 3)
			var status string
			if iter.CompletionDetected {
				status = "✅"
			} else if iter.ExitCode != 0 {
				status = "❌"
			} else {
				status = "🔄"
			}
			if toolsSummary == "" {
				toolsSummary = "no tools"
			}
			fmt.Printf("   %s #%d: %s | %s\n", status, iter.Iteration, tools.FormatDurationLong(iter.DurationMs), toolsSummary)
		}

		struggle := h.StruggleIndicators
		if struggle.NoProgressIterations >= 3 || struggle.ShortIterations >= 3 || hasRepeatedErrors(struggle) {
			fmt.Println("\n⚠️  STRUGGLE INDICATORS:")
			if struggle.NoProgressIterations >= 3 {
				fmt.Printf("   - No file changes in %d iterations\n", struggle.NoProgressIterations)
			}
			if struggle.ShortIterations >= 3 {
				fmt.Printf("   - %d very short iterations (< 30s)\n", struggle.ShortIterations)
			}
			topErrors := getTopErrors(struggle, 3)
			for _, err := range topErrors {
				fmt.Printf("   - Same error %dx: \"%s...\"\n", err.count, truncate(err.msg, 50))
			}
			fmt.Println("\n   💡 Consider using: ralph --add-context \"your hint here\"")
		}
	}

	fmt.Println("")
}

type errorCount struct {
	msg   string
	count int
}

func hasRepeatedErrors(s state.StruggleIndicators) bool {
	for _, count := range s.RepeatedErrors {
		if count >= 2 {
			return true
		}
	}
	return false
}

func getTopErrors(s state.StruggleIndicators, limit int) []errorCount {
	var errors []errorCount
	for msg, count := range s.RepeatedErrors {
		if count >= 2 {
			errors = append(errors, errorCount{msg: msg, count: count})
		}
	}
	if len(errors) <= limit {
		return errors
	}
	return errors[:limit]
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func ellipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return ""
	}
	return "..."
}
