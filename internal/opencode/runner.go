package opencode

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type RunOpenCodeOptions struct {
	Prompt              string
	Model               string
	StreamOutput        bool
	VerboseTools        bool
	DisablePlugins      bool
	AllowAllPermissions bool
	IterationStart      time.Time
	Verbose             bool
	Timeout             time.Duration
}

func RunOpenCode(opts *RunOpenCodeOptions) (*StreamResult, int, error) {
	args := []string{"run"}

	if opts.Model != "" {
		args = append(args, "-m", opts.Model)
	}

	if opts.Verbose {
		args = append(args, "--log-level", "DEBUG")
	}

	args = append(args, opts.Prompt)

	env := os.Environ()

	if opts.DisablePlugins || opts.AllowAllPermissions {
		configPath, err := ensureRalphConfig(&ConfigOptions{
			FilterPlugins:       opts.DisablePlugins,
			AllowAllPermissions: opts.AllowAllPermissions,
		})
		if err != nil {
			return nil, -1, fmt.Errorf("failed to create Ralph config: %w", err)
		}
		env = append(env, fmt.Sprintf("OPENCODE_CONFIG=%s", configPath))
	}

	cmd := exec.Command("opencode", args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, -1, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, -1, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, -1, fmt.Errorf("failed to start opencode: %w", err)
	}

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	var result *StreamResult
	if opts.StreamOutput {
		result, err = StreamProcessOutput(stdout, stderr, !opts.VerboseTools, opts.IterationStart, opts.Timeout)
		if err != nil {
			cmd.Process.Kill()
			return nil, -1, err
		}
	} else {
		result, err = BufferProcessOutput(stdout, stderr)
		if err != nil {
			cmd.Process.Kill()
			return nil, -1, err
		}
	}

	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return result, exitCode, nil
}
