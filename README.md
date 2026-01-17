# Ralph Wiggum for OpenCode — Go Edition

**Iterative AI development loops. Same prompt. Persistent progress.**

This is a **Go port** of the original [Ralph Wiggum for OpenCode](https://github.com/Th0rgal/opencode-ralph-wiggum) (originally built with Bun + TypeScript).

> Based on the original technique by [Geoffrey Huntley](https://ghuntley.com/ralph/)

---

## Why a Go Port?

The original Ralph is excellent, but JavaScript/TypeScript has limitations for CLI tools:

- **Runtime overhead**: Requires Bun, Node.js, or similar runtime to be available
- **Distribution complexity**: Larger binaries, dependency management challenges
- **Startup latency**: Noticeable delays on every invocation
- **Production concerns**: Server-side JavaScript introduces operational complexity

**Go solves these problems:**

- **Single binary**: No runtime dependencies. Copy and run anywhere.
- **Fast startup**: Instant execution, no warm-up time.
- **Zero dependencies**: Uses only the Go standard library for maximum reliability.
- **Better for CLI**: Go is purpose-built for command-line tools.
- **Easy deployment**: Trivial to distribute and integrate into workflows.

This port maintains **100% feature parity** with the original while delivering a better tool for serious CLI work.

---

## What is Ralph?

Ralph is a development methodology where an AI agent receives the **same prompt repeatedly** until it completes a task. Each iteration, the AI sees its previous work in files and git history, enabling self-correction and incremental progress.

```bash
# The essence of Ralph:
while true; do
  opencode run "Build feature X. Output <promise>DONE</promise> when complete."
done
```

**The AI doesn't talk to itself.** It sees the same prompt each time, but the files have changed from previous iterations. This creates a feedback loop where the AI iteratively improves its work until success.

## Why Ralph?

| Benefit | How it works |
|---------|--------------|
| **Self-Correction** | AI sees test failures from previous runs, fixes them |
| **Persistence** | Walk away, come back to completed work |
| **Iteration** | Complex tasks broken into incremental progress |
| **Automation** | No babysitting—loop handles retries |
| **Observability** | Monitor progress with `--status`, see history and struggle indicators |
| **Mid-Loop Guidance** | Inject hints with `--add-context` without stopping the loop |

---

## Installation

**Prerequisites:** [OpenCode](https://opencode.ai)

### From Release (Recommended)

Download the latest binary from [releases](https://github.com/wltechblog/ralphy/releases) and add it to your `$PATH`.

### Using `go install`

```bash
go install github.com/wltechblog/ralphy/cmd/ralph@latest
```

The `ralphy` binary will be placed in `$GOPATH/bin/` (typically `~/go/bin/`).

### From Source

```bash
git clone https://github.com/wltechblog/ralphy
cd ralphy
go build ./cmd/ralph -o ralphy
```

The `ralphy` binary will be in the current directory.

---

## Quick Start

```bash
# Simple task with iteration limit
ralphy "Create a hello.txt file with 'Hello World'. Output <promise>DONE</promise> when complete." \
  --max-iterations 5

# Build something real
ralphy "Build a REST API for todos with CRUD operations and tests. \
  Run tests after each change. Output <promise>COMPLETE</promise> when all tests pass." \
  --max-iterations 20
```

---

## Commands

### Running a Loop

```bash
ralphy "<prompt>" [options]

Options:
  --max-iterations N       Stop after N iterations (default: unlimited)
  --model MODEL            OpenCode model to use
  --prompt-file FILE       Read prompt from a file
  -f FILE                  Shorthand for --prompt-file
  --no-stream              Buffer output, print at the end
  --verbose-tools          Print every tool line (disable compact summary)
  --no-plugins             Disable non-auth OpenCode plugins
  --no-commit              Don't auto-commit after iterations
  --add-context TEXT       Add context hint for next iteration
  --clear-context          Clear pending context
  --status                 Show loop status and history
  --version                Show version
  --help                   Show this help message
```

### Monitoring & Control

```bash
# Check status of active loop (run from another terminal)
ralphy --status

# Add context/hints for the next iteration
ralphy --add-context "Focus on fixing the auth module first"

# Clear pending context
ralphy --clear-context
```

### Status Dashboard

The `--status` command shows:
- **Active loop info**: Current iteration, elapsed time, prompt
- **Pending context**: Any hints queued for next iteration
- **Iteration history**: Last 5 iterations with tools used, duration
- **Struggle indicators**: Warnings if agent is stuck

```
╔══════════════════════════════════════════════════════════════════╗
║                    Ralph Wiggum Status                           ║
╚══════════════════════════════════════════════════════════════════╝

🔄 ACTIVE LOOP
   Iteration:    3 / 10
   Elapsed:      5m 23s
   Promise:      COMPLETE
   Prompt:       Build a REST API...

📊 HISTORY (3 iterations)
   Total time:   5m 23s

   Recent iterations:
   🔄 #1: 2m 10s | Bash:5 Write:3 Read:2
   🔄 #2: 1m 45s | Edit:4 Bash:3 Read:2
   🔄 #3: 1m 28s | Bash:2 Edit:1

⚠️  STRUGGLE INDICATORS:
   - No file changes in 3 iterations
   💡 Consider using: ralphy --add-context "your hint here"
```

### Mid-Loop Context Injection

Guide a struggling agent without stopping the loop:

```bash
# In another terminal while loop is running
ralphy --add-context "The bug is in utils/parser.ts line 42"
ralphy --add-context "Try using the singleton pattern for config"
```

Context is automatically consumed after one iteration.

---

## Writing Good Prompts

### Include Clear Success Criteria

❌ Bad:
```
Build a todo API
```

✅ Good:
```
Build a REST API for todos with:
- CRUD endpoints (GET, POST, PUT, DELETE)
- Input validation
- Tests for each endpoint

Run tests after changes. Output <promise>COMPLETE</promise> when all tests pass.
```

### Use Verifiable Conditions

❌ Bad:
```
Make the code better
```

✅ Good:
```
Refactor auth.ts to:
1. Extract validation into separate functions
2. Add error handling for network failures
3. Ensure all existing tests still pass

Output <promise>DONE</promise> when refactored and tests pass.
```

### Always Set Max Iterations

```bash
# Safety net for runaway loops
ralphy "Your task" --max-iterations 20
```

---

## When to Use Ralph

**Good for:**
- Tasks with automatic verification (tests, linters, type checking)
- Well-defined tasks with clear completion criteria
- Greenfield projects where you can walk away
- Iterative refinement (getting tests to pass)

**Not good for:**
- Tasks requiring human judgment
- One-shot operations
- Unclear success criteria
- Production debugging

---

## How It Works

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   ┌──────────┐    same prompt    ┌──────────┐              │
│   │          │ ───────────────▶  │          │              │
│   │  ralphy  │                   │ OpenCode │              │
│   │   CLI    │ ◀─────────────── │          │              │
│   │          │   output + files  │          │              │
│   └──────────┘                   └──────────┘              │
│        │                              │                     │
│        │ check for                    │ modify              │
│        │ <promise>                    │ files               │
│        ▼                              ▼                     │
│   ┌──────────┐                   ┌──────────┐              │
│   │ Complete │                   │   Git    │              │
│   │   or     │                   │  Repo    │              │
│   │  Retry   │                   │ (state)  │              │
│   └──────────┘                   └──────────┘              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

1. Ralphy sends your prompt to OpenCode
2. OpenCode works on the task, modifies files
3. Ralphy checks output for completion promise
4. If not found, repeat with same prompt
5. AI sees previous work in files
6. Loop until success or max iterations

---

## Project Structure

```
ralphy/
├── cmd/ralph/                    # CLI entrypoint
│   ├── main.go                   # Main entry point
│   └── status.go                 # Status command implementation
├── internal/
│   ├── loop/                     # Core loop orchestration
│   ├── state/                    # State management and persistence
│   ├── opencode/                 # OpenCode API integration
│   ├── git/                      # Git functionality
│   ├── tools/                    # Utility functions
│   └── ui/                       # UI components
├── go.mod                        # Module definition
└── README.md                     # This file
```

### State Files (in `.opencode/`)

During operation, Ralphy stores state in `.opencode/`:
- `ralph-loop.state.json` — Active loop state
- `ralph-history.json` — Iteration history and metrics
- `ralph-context.md` — Pending context for next iteration

---

## Building from Source

**Requirements:** Go 1.24+

```bash
git clone https://github.com/wltechblog/ralphy
cd ralphy
go build ./cmd/ralph -o ralphy
./ralphy --help
```

To install globally:
```bash
go install ./cmd/ralph@latest
```

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## Credits

This is a **community port** to Go by [WL Tech Blog](https://github.com/wltechblog).

**Original project:** [Ralph Wiggum for OpenCode](https://github.com/Th0rgal/opencode-ralph-wiggum) by [Th0rgal](https://github.com/Th0rgal)  
**Original technique:** [Geoffrey Huntley — Ralph](https://ghuntley.com/ralph/)  
**Related projects:**
- [Ralph Orchestrator](https://github.com/mikeyobrien/ralph-orchestrator)
- [OpenAgent](https://github.com/Th0rgal/openagent)

---

## License

MIT