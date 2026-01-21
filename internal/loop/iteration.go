package loop

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wltechblog/ralphy/internal/git"
	"github.com/wltechblog/ralphy/internal/opencode"
	"github.com/wltechblog/ralphy/internal/state"
	"github.com/wltechblog/ralphy/internal/tools"
)

type IterationResult struct {
	ExitCode           int
	CompletionDetected bool
	DurationMs         int64
	ToolCounts         map[string]int
	FilesModified      []string
	Errors             []string
}

func RunIteration(s *state.RalphState, h *state.RalphHistory, autoCommit bool, timeout time.Duration) (*IterationResult, error) {
	fmt.Printf("\n🔄 Iteration %d", s.Iteration)
	if s.MaxIterations > 0 {
		fmt.Printf(" / %d", s.MaxIterations)
	}
	fmt.Println("")
	fmt.Println(strings.Repeat("─", 68))

	contextAtStart, _ := state.LoadContext()

	snapshotBefore, err := git.CaptureFileSnapshot()
	if err != nil {
		snapshotBefore = &git.FileSnapshot{Files: map[string]string{}}
	}

	fullPrompt := opencode.BuildPrompt(s, contextAtStart)
	iterationStart := time.Now()

	var result *IterationResult
	var exitCode int

	opencodeResult, code, err := opencode.RunOpenCode(&opencode.RunOpenCodeOptions{
		Prompt:              fullPrompt,
		Model:               s.Model,
		StreamOutput:        true,
		VerboseTools:        false,
		DisablePlugins:      false,
		AllowAllPermissions: false,
		IterationStart:      iterationStart,
		Timeout:             timeout,
	})

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			fmt.Printf("\n⏳ Iteration %d timed out after %v of inactivity.\n", s.Iteration, timeout)
			state.SaveContext(fmt.Sprintf("Iteration %d timed out after %v of inactivity. Please try again or take a different approach.", s.Iteration, timeout))
			
			// Return a partial result to keep history happy, but marked as failure
			return &IterationResult{
				ExitCode:           -1,
				CompletionDetected: false,
				DurationMs:         time.Since(iterationStart).Milliseconds(),
				ToolCounts:         map[string]int{},
				FilesModified:      []string{},
				Errors:             []string{err.Error()},
			}, nil
		}
		return nil, fmt.Errorf("failed to run opencode: %w", err)
	}

	exitCode = code
	iterationDuration := time.Since(iterationStart)

	snapshotAfter, _ := git.CaptureFileSnapshot()
	filesModified := git.GetModifiedFilesSinceSnapshot(snapshotBefore, snapshotAfter)

	combinedOutput := opencodeResult.StdoutText + "\n" + opencodeResult.StderrText
	completionDetected := CheckCompletion(combinedOutput, s.CompletionPromise)

	errors := tools.ExtractErrors(combinedOutput)

	result = &IterationResult{
		ExitCode:           exitCode,
		CompletionDetected: completionDetected,
		DurationMs:         iterationDuration.Milliseconds(),
		ToolCounts:         opencodeResult.ToolCounts,
		FilesModified:      filesModified,
		Errors:             errors,
	}

	printIterationSummary(s.Iteration, iterationDuration.Milliseconds(), opencodeResult.ToolCounts, exitCode, completionDetected)

	state.AddIteration(h, &state.IterationHistory{
		Iteration:          s.Iteration,
		StartedAt:          iterationStart.Format(time.RFC3339),
		EndedAt:            time.Now().Format(time.RFC3339),
		DurationMs:         iterationDuration.Milliseconds(),
		ToolsUsed:          opencodeResult.ToolCounts,
		FilesModified:      filesModified,
		ExitCode:           exitCode,
		CompletionDetected: completionDetected,
		Errors:             errors,
	})

	state.UpdateStruggleIndicators(h, &state.IterationHistory{
		Iteration:     s.Iteration,
		FilesModified: filesModified,
		DurationMs:    iterationDuration.Milliseconds(),
		Errors:        errors,
	})

	state.SaveHistory(h)

	if s.Iteration > 2 && (h.StruggleIndicators.NoProgressIterations >= 3 || h.StruggleIndicators.ShortIterations >= 3) {
		fmt.Println("\n⚠️  Potential struggle detected:")
		if h.StruggleIndicators.NoProgressIterations >= 3 {
			fmt.Printf("   - No file changes in %d iterations\n", h.StruggleIndicators.NoProgressIterations)
		}
		if h.StruggleIndicators.ShortIterations >= 3 {
			fmt.Printf("   - %d very short iterations\n", h.StruggleIndicators.ShortIterations)
		}
		fmt.Println("   💡 Tip: Use 'ralph --add-context \"hint\"' in another terminal to guide the agent")
	}

	if DetectPlaceholderPluginError(combinedOutput) {
		fmt.Fprintln(os.Stderr, "\n❌ OpenCode tried to load legacy 'ralph-wiggum' plugin. This package is CLI-only.")
		fmt.Fprintln(os.Stderr, "Remove 'ralph-wiggum' from your opencode.json plugin list, or re-run with --no-plugins.")
		return nil, fmt.Errorf("placeholder plugin detected")
	}

	if exitCode != 0 {
		fmt.Printf("\n⚠️  OpenCode exited with code %d. Continuing to next iteration.\n", exitCode)
	}

	if autoCommit {
		message := fmt.Sprintf("Ralph iteration %d: work in progress", s.Iteration)
		if completionDetected {
			message = fmt.Sprintf("Ralph iteration %d: task completed", s.Iteration)
		}

		committed, err := git.AutoCommit(message)
		if err != nil {
			fmt.Printf("⚠️  Git auto-commit failed: %v\n", err)
		} else if committed {
			fmt.Println("📝 Auto-committed changes")
		}
	}

	if completionDetected {
		fmt.Printf("\n╔══════════════════════════════════════════════════════════════════╗\n")
		fmt.Printf("║  ✅ Completion promise detected: <promise>%s</promise>\n", s.CompletionPromise)
		fmt.Printf("║  Task completed in %d iteration(s)\n", s.Iteration)
		fmt.Printf("║  Total time: %s\n", tools.FormatDurationLong(h.TotalDurationMs))
		fmt.Printf("╚══════════════════════════════════════════════════════════════════╝\n")
		state.ClearState()
		state.ClearHistory()
		state.ClearContext()
		return result, nil
	}

	if contextAtStart != "" {
		fmt.Println("📝 Context was consumed this iteration")
		state.ClearContext()
	}

	s.Iteration++
	state.SaveState(s)

	time.Sleep(1 * time.Second)

	return result, nil
}

func printIterationSummary(iteration int, elapsedMs int64, toolCounts map[string]int, exitCode int, completionDetected bool) {
	fmt.Println("\nIteration Summary")
	fmt.Println("────────────────────────────────────────────────────────────────────")
	fmt.Printf("Iteration: %d\n", iteration)
	fmt.Printf("Elapsed:   %s\n", tools.FormatDuration(elapsedMs))

	toolSummary := tools.FormatToolSummary(toolCounts, 6)
	if toolSummary != "" {
		fmt.Printf("Tools:     %s\n", toolSummary)
	} else {
		fmt.Println("Tools:     none")
	}

	fmt.Printf("Exit code: %d\n", exitCode)
	fmt.Printf("Completion promise: %t\n", completionDetected)
}
