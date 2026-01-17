package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wltechblog/ralphy/internal/loop"
	"github.com/wltechblog/ralphy/internal/state"
)

func main() {
	flagHelp := flag.Bool("help", false, "Show help")
	flagVersion := flag.Bool("version", false, "Show version")
	flagStatus := flag.Bool("status", false, "Show Ralphy loop status")
	addContext := flag.String("add-context", "", "Add context for next iteration")
	flagClearContext := flag.Bool("clear-context", false, "Clear pending context")

	maxIterations := flag.Int("max-iterations", 0, "Maximum iterations before stopping (default: unlimited)")
	completionPromise := flag.String("completion-promise", "COMPLETE", "Phrase that signals completion")
	model := flag.String("model", "", "Model to use (e.g., anthropic/claude-sonnet)")
	promptFile := flag.String("prompt-file", "", "Read prompt content from a file")
	noStream := flag.Bool("no-stream", false, "Buffer OpenCode output and print at the end")
	verboseTools := flag.Bool("verbose-tools", false, "Print every tool line")
	noPlugins := flag.Bool("no-plugins", false, "Disable non-auth OpenCode plugins")
	noCommit := flag.Bool("no-commit", false, "Don't auto-commit after each iteration")
	allowAll := flag.Bool("allow-all", false, "Auto-approve all tool permissions")

	flag.Usage = func() {
		fmt.Println(`
Ralphy Wiggum Loop - Iterative AI development with OpenCode

Usage:
  ralphy "<prompt>" [options]
  ralphy --prompt-file <path> [options]

Arguments:
  prompt              Task description for the AI to work on

Options:
  --max-iterations N  Maximum iterations before stopping (default: unlimited)
  --completion-promise TEXT  Phrase that signals completion (default: COMPLETE)
  --model MODEL       Model to use (e.g., anthropic/claude-sonnet)
  --prompt-file, --file, -f  Read prompt content from a file
  --no-stream         Buffer OpenCode output and print at the end
  --verbose-tools     Print every tool line (disable compact tool summary)
  --no-plugins        Disable non-auth OpenCode plugins for this run
  --no-commit         Don't auto-commit after each iteration
  --allow-all         Auto-approve all tool permissions (for non-interactive use)
  --version, -v       Show version
  --help, -h          Show this help

Commands:
  --status            Show current ralphy loop status and history
  --add-context TEXT  Add context for the next iteration (or edit .opencode/ralph-context.md)
  --clear-context     Clear any pending context

Examples:
  ralphy "Build a REST API for todos"
  ralphy "Fix auth bug" --max-iterations 10
  ralphy "Add tests" --completion-promise "ALL TESTS PASS" --model openai/gpt-5.1
  ralphy --prompt-file ./prompt.md --max-iterations 5
  ralphy --status                                        # Check loop status
  ralphy --add-context "Focus on the auth module first"  # Add hint for next iteration

How it works:
  1. Sends your prompt to OpenCode
  2. AI works on the task
  3. Checks output for completion promise
  4. If not complete, repeats with same prompt
  5. AI sees its previous work in files
  6. Continues until promise detected or max iterations

To stop manually: Ctrl+C

Learn more: https://ghuntley.com/ralph/
`)
	}

	flag.Parse()

	if *flagHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *flagVersion {
		fmt.Printf("ralphy %s\n", state.VERSION)
		os.Exit(0)
	}

	if *flagStatus {
		printStatus()
		os.Exit(0)
	}

	if *addContext != "" {
		if *addContext == "" {
			fmt.Fprintln(os.Stderr, "Error: --add-context requires a text argument")
			fmt.Fprintln(os.Stderr, "Usage: ralphy --add-context \"Your context or hint here\"")
			os.Exit(1)
		}

		if err := state.SaveContext(*addContext); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving context: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Context added for next iteration")
		contextPath := ".opencode/ralphy-context.md"
		fmt.Printf("   File: %s\n", contextPath)

		existingState, _ := state.LoadState()
		if existingState != nil && existingState.Active {
			fmt.Printf("   Will be picked up in iteration %d\n", existingState.Iteration+1)
		} else {
			fmt.Println("   Will be used when loop starts")
		}
		os.Exit(0)
	}

	if *flagClearContext {
		if err := state.ClearContext(); err != nil {
			fmt.Fprintln(os.Stderr, "Error clearing context:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Context cleared")
		os.Exit(0)
	}

	args := flag.Args()
	var promptParts []string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			promptParts = append(promptParts, arg)
		}
	}

	var prompt string
	var promptSource string

	if *promptFile != "" {
		promptSource = *promptFile
		content, err := os.ReadFile(*promptFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Prompt file not found: %s\n", *promptFile)
			os.Exit(1)
		}
		prompt = string(content)
	} else if len(promptParts) == 1 {
		promptSource = promptParts[0]
		content, err := os.ReadFile(promptParts[0])
		if err == nil {
			prompt = string(content)
		} else {
			prompt = strings.Join(promptParts, " ")
		}
	} else {
		prompt = strings.Join(promptParts, " ")
	}

	if prompt == "" {
		fmt.Fprintln(os.Stderr, "Error: No prompt provided")
		fmt.Fprintln(os.Stderr, "Usage: ralphy \"Your task description\" [options]")
		fmt.Fprintln(os.Stderr, "Run 'ralphy --help' for more information")
		os.Exit(1)
	}

	opts := &RunOptions{
		Prompt:              prompt,
		PromptSource:        promptSource,
		MaxIterations:       *maxIterations,
		CompletionPromise:   *completionPromise,
		Model:               *model,
		StreamOutput:        !*noStream,
		VerboseTools:        *verboseTools,
		DisablePlugins:      *noPlugins,
		AutoCommit:          !*noCommit,
		AllowAllPermissions: *allowAll,
	}

	if err := loop.RunLoop(&loop.LoopOptions{
		Prompt:              opts.Prompt,
		PromptSource:        opts.PromptSource,
		MaxIterations:       opts.MaxIterations,
		CompletionPromise:   opts.CompletionPromise,
		Model:               opts.Model,
		StreamOutput:        opts.StreamOutput,
		VerboseTools:        opts.VerboseTools,
		DisablePlugins:      opts.DisablePlugins,
		AutoCommit:          opts.AutoCommit,
		AllowAllPermissions: opts.AllowAllPermissions,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		state.ClearState()
		os.Exit(1)
	}
}

type RunOptions struct {
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
}
