package loop

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/wltechblog/ralphy/internal/state"
)

type LoopOptions struct {
	Prompt              string
	PromptSource        string
	MaxIterations       int
	CompletionPromise   string
	Model               string
	StreamOutput        bool
	VerboseTools        bool
	DisablePlugins      bool
	AutoCommit          bool
	AllowAllPermissions bool
	Verbose             bool
	Timeout             time.Duration
}

func RunLoop(opts *LoopOptions) error {
	existingState, err := state.LoadState()
	if err == nil && existingState.Active {
		return fmt.Errorf("a Ralph loop is already active (iteration %d)\nStarted at: %s\nTo cancel it, press Ctrl+C in its terminal or delete %s",
			existingState.Iteration, existingState.StartedAt, ".opencode/ralph-loop.state.json")
	}

	fmt.Println(`
╔══════════════════════════════════════════════════════════════════╗
║                    Ralph Wiggum Loop                            ║
║            Iterative AI Development with OpenCode                ║
╚══════════════════════════════════════════════════════════════════╝`)

	s := &state.RalphState{
		Active:            true,
		Iteration:         1,
		MaxIterations:     opts.MaxIterations,
		CompletionPromise: opts.CompletionPromise,
		Prompt:            opts.Prompt,
		StartedAt:         time.Now().Format(time.RFC3339),
		Model:             opts.Model,
	}

	state.SaveState(s)

	h, err := state.LoadHistory()
	if err != nil {
		h = &state.RalphHistory{
			Iterations:      []state.IterationHistory{},
			TotalDurationMs: 0,
			StruggleIndicators: state.StruggleIndicators{
				RepeatedErrors:       map[string]int{},
				NoProgressIterations: 0,
				ShortIterations:      0,
			},
		}
	}
	state.SaveHistory(h)

	promptPreview := opts.Prompt
	if len(promptPreview) > 80 {
		promptPreview = promptPreview[:80] + "..."
	}

	if opts.PromptSource != "" {
		fmt.Printf("Task: %s\n", opts.PromptSource)
		fmt.Printf("Preview: %s\n", promptPreview)
	} else {
		fmt.Printf("Task: %s\n", promptPreview)
	}

	fmt.Printf("Completion promise: %s\n", opts.CompletionPromise)
	maxIter := "unlimited"
	if opts.MaxIterations > 0 {
		maxIter = fmt.Sprintf("%d", opts.MaxIterations)
	}
	fmt.Printf("Max iterations: %s\n", maxIter)
	if opts.Model != "" {
		fmt.Printf("Model: %s\n", opts.Model)
	}
	if opts.DisablePlugins {
		fmt.Println("OpenCode plugins: non-auth plugins disabled")
	}
	if opts.AllowAllPermissions {
		fmt.Println("Permissions: auto-approve all tools")
	}
	if opts.Timeout > 0 {
		fmt.Printf("Timeout: %v\n", opts.Timeout)
	}

	fmt.Println("")
	fmt.Println("Starting loop... (Ctrl+C to stop)")
	fmt.Println(strings.Repeat("═", 68))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go func() {
		<-sigChan
		fmt.Println("\nGracefully stopping Ralph loop...")
		state.ClearState()
		fmt.Println("Loop cancelled.")
		os.Exit(0)
	}()

	for {
		if opts.MaxIterations > 0 && s.Iteration > opts.MaxIterations {
			fmt.Println("\n╔══════════════════════════════════════════════════════════════════╗")
			fmt.Printf("║  Max iterations (%d) reached. Loop stopped.\n", opts.MaxIterations)
			fmt.Printf("║  Total time: %s\n", formatDurationLong(h.TotalDurationMs))
			fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
			state.ClearState()
			return nil
		}

		result, err := RunIteration(s, h, opts.AutoCommit, opts.Timeout, opts.Verbose, opts.VerboseTools)
		if err == nil {
			h.TotalDurationMs += result.DurationMs
			state.SaveHistory(h)
		}

		if err != nil {
			fmt.Printf("\n❌ Error in iteration %d: %v\n", s.Iteration, err)
			fmt.Println("Continuing to next iteration...")

			iterationDuration := time.Since(time.Now().Add(-time.Second))
			errorRecord := &state.IterationHistory{
				Iteration:          s.Iteration,
				StartedAt:          time.Now().Add(-iterationDuration).Format(time.RFC3339),
				EndedAt:            time.Now().Format(time.RFC3339),
				DurationMs:         iterationDuration.Milliseconds(),
				ToolsUsed:          map[string]int{},
				FilesModified:      []string{},
				ExitCode:           -1,
				CompletionDetected: false,
				Errors:             []string{fmt.Sprintf("%v", err)},
			}
			state.AddIteration(h, errorRecord)
			h.TotalDurationMs += iterationDuration.Milliseconds()
			state.SaveHistory(h)

			s.Iteration++
			state.SaveState(s)

			time.Sleep(2 * time.Second)
			continue
		}

		if result.CompletionDetected {
			return nil
		}
	}
}

func formatDurationLong(ms int64) string {
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
