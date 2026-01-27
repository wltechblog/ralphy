package state

const (
	VERSION         = "1.0.9"
	stateDirName    = ".opencode"
	stateFileName   = "ralph-loop.state.json"
	historyFileName = "ralph-history.json"
	contextFileName = "ralph-context.md"
	tasksFileName   = "ralph-tasks.md"
)

type RalphState struct {
	Active            bool   `json:"active"`
	Iteration         int    `json:"iteration"`
	MaxIterations     int    `json:"maxIterations"`
	CompletionPromise string `json:"completionPromise"`
	TaskPromise       string `json:"taskPromise"`
	Prompt            string `json:"prompt"`
	StartedAt         string `json:"startedAt"`
	Model             string `json:"model"`
}

type IterationHistory struct {
	Iteration          int            `json:"iteration"`
	StartedAt          string         `json:"startedAt"`
	EndedAt            string         `json:"endedAt"`
	DurationMs         int64          `json:"durationMs"`
	ToolsUsed          map[string]int `json:"toolsUsed"`
	FilesModified      []string       `json:"filesModified"`
	ExitCode           int            `json:"exitCode"`
	CompletionDetected bool           `json:"completionDetected"`
	Errors             []string       `json:"errors"`
}

type RalphHistory struct {
	Iterations         []IterationHistory `json:"iterations"`
	TotalDurationMs    int64              `json:"totalDurationMs"`
	StruggleIndicators StruggleIndicators `json:"struggleIndicators"`
}

type StruggleIndicators struct {
	RepeatedErrors       map[string]int `json:"repeatedErrors"`
	NoProgressIterations int            `json:"noProgressIterations"`
	ShortIterations      int            `json:"shortIterations"`
}
