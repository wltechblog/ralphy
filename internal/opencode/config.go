package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wltechblog/ralphy/internal/state"
)

func LoadPluginsFromConfig(configPath string) []string {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return []string{}
	}

	raw := string(data)

	withoutBlock := regexp.MustCompile(`\/\*[\s\S]*?\*\/`).ReplaceAllString(raw, "")
	withoutLine := regexp.MustCompile(`^\s*\/\/.*$`).ReplaceAllString(withoutBlock, "")

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(withoutLine), &parsed); err != nil {
		return []string{}
	}

	plugins, ok := parsed["plugin"]
	if !ok {
		return []string{}
	}

	pluginList, ok := plugins.([]interface{})
	if !ok {
		return []string{}
	}

	var result []string
	for _, p := range pluginList {
		if plugin, ok := p.(string); ok {
			result = append(result, plugin)
		}
	}

	return result
}

func ensureRalphConfig(options *ConfigOptions) (string, error) {
	if err := state.EnsureStateDir(); err != nil {
		return "", err
	}

	stateDir, err := state.GetStateDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(stateDir, "ralph-opencode.config.json")

	config := make(map[string]interface{})
	config["$schema"] = "https://opencode.ai/config.json"

	if options.FilterPlugins {
		var plugins []string

		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome == "" {
			home := os.Getenv("HOME")
			xdgConfigHome = filepath.Join(home, ".config")
		}

		userConfigPath := filepath.Join(xdgConfigHome, "opencode", "opencode.json")
		projectConfigPath := filepath.Join(stateDir, "..", ".opencode", "opencode.json")

		plugins = append(plugins, LoadPluginsFromConfig(userConfigPath)...)
		plugins = append(plugins, LoadPluginsFromConfig(projectConfigPath)...)

		uniquePlugins := make(map[string]bool)
		var filteredPlugins []string
		for _, p := range plugins {
			if !uniquePlugins[p] {
				uniquePlugins[p] = true
				authMatch := regexp.MustCompile(`(?i)auth`).MatchString(p)
				if authMatch {
					filteredPlugins = append(filteredPlugins, p)
				}
			}
		}
		config["plugin"] = filteredPlugins
	}

	if options.AllowAllPermissions {
		config["permission"] = map[string]string{
			"read":               "allow",
			"edit":               "allow",
			"glob":               "allow",
			"grep":               "allow",
			"list":               "allow",
			"bash":               "allow",
			"task":               "allow",
			"webfetch":           "allow",
			"websearch":          "allow",
			"codesearch":         "allow",
			"todowrite":          "allow",
			"todoread":           "allow",
			"question":           "allow",
			"lsp":                "allow",
			"external_directory": "allow",
		}
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", err
	}

	return configPath, nil
}

type ConfigOptions struct {
	FilterPlugins       bool
	AllowAllPermissions bool
}

func BuildPrompt(s *state.RalphState, context string) string {
	var contextSection strings.Builder

	if context != "" {
		contextSection.WriteString(`
## Additional Context (added by user mid-loop)

`)
		contextSection.WriteString(context)
		contextSection.WriteString(`

---
`)
	}

	tasksSection := state.GetTasksModeSection(s)
	prompt := fmt.Sprintf(`# Ralph Wiggum Loop - Iteration %d

You are in an iterative development loop working through a task list.
%s%s
## Your Main Goal

%s

## Critical Rules

- **Update your todo list and PROGRESS.md at the start of each iteration** to show progress. PROGRESS.md ensures your status persists across iterations.
- Work on ONE task at a time from .opencode/ralph-tasks.md
- ONLY output <promise>%s</promise> when the current task is complete and marked in ralph-tasks.md
- ONLY output <promise>%s</promise> when ALL tasks are truly done
- Do NOT lie or output false promises to exit the loop
- If stuck, try a different approach
- Check your work before claiming completion

## Current Iteration: %d%s

Now, work on the current task. Good luck!`,
		s.Iteration,
		contextSection.String(),
		tasksSection,
		s.Prompt,
		s.TaskPromise,
		s.CompletionPromise,
		s.Iteration,
		formatMaxIterations(s.MaxIterations),
	)
	return strings.TrimSpace(prompt)
}

func formatMaxIterations(maxIterations int) string {
	if maxIterations > 0 {
		return fmt.Sprintf(" / %d", maxIterations)
	}
	return " (unlimited)"
}
