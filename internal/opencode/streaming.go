package opencode

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/wltechblog/ralphy/internal/tools"
)

type StreamResult struct {
	StdoutText string
	StderrText string
	ToolCounts map[string]int
}

func StreamProcessOutput(stdout, stderr io.Reader, compactTools bool, iterationStart time.Time) (*StreamResult, error) {
	toolCounts := make(map[string]int)
	var stdoutText, stderrText strings.Builder

	lastPrintedAt := time.Now()
	lastActivityAt := time.Now()
	lastToolSummaryAt := time.Time{}

	toolSummaryInterval := 3 * time.Second
	heartbeatInterval := 10 * time.Second

	maybePrintToolSummary := func(force bool) {
		if !compactTools || len(toolCounts) == 0 {
			return
		}
		now := time.Now()
		if !force && now.Sub(lastToolSummaryAt) < toolSummaryInterval {
			return
		}
		summary := tools.FormatToolSummary(toolCounts, 6)
		if summary != "" {
			fmt.Printf("| Tools    %s\n", summary)
			lastPrintedAt = now
			lastToolSummaryAt = now
		}
	}

	handleLine := func(line string, isError bool) {
		lastActivityAt = time.Now()
		toolMatch := toolPattern(line)
		if compactTools && toolMatch != "" {
			toolCounts[toolMatch]++
			maybePrintToolSummary(false)
			return
		}
		if line == "" {
			fmt.Println("")
			lastPrintedAt = time.Now()
			return
		}
		if isError {
			fmt.Fprintln(os.Stderr, line)
		} else {
			fmt.Println(line)
		}
		lastPrintedAt = time.Now()
	}

	stream := func(r io.Reader, text *strings.Builder, isError bool) error {
		if r == nil {
			return nil
		}
		scanner := bufio.NewScanner(r)
		var buffer string

		for scanner.Scan() {
			chunk := scanner.Text()
			text.WriteString(chunk)
			buffer += chunk + "\n"
			lines := strings.Split(buffer, "\n")
			buffer = lines[len(lines)-1]

			for i := 0; i < len(lines)-1; i++ {
				handleLine(lines[i], isError)
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		if buffer != "" {
			handleLine(buffer, isError)
		}

		return nil
	}

	heartbeatTimer := time.NewTicker(heartbeatInterval)
	defer heartbeatTimer.Stop()

	done := make(chan bool, 2)
	errChan := make(chan error, 2)

	go func() {
		if err := stream(stdout, &stdoutText, false); err != nil {
			errChan <- err
		}
		done <- true
	}()

	go func() {
		if err := stream(stderr, &stderrText, true); err != nil {
			errChan <- err
		}
		done <- true
	}()

	for {
		select {
		case <-done:
			return &StreamResult{
				StdoutText: stdoutText.String(),
				StderrText: stderrText.String(),
				ToolCounts: toolCounts,
			}, nil
		case err := <-errChan:
			return nil, err
		case <-heartbeatTimer.C:
			now := time.Now()
			if now.Sub(lastPrintedAt) >= heartbeatInterval {
				elapsed := tools.FormatDuration(now.Sub(iterationStart).Milliseconds())
				sinceActivity := tools.FormatDuration(now.Sub(lastActivityAt).Milliseconds())
				fmt.Printf("⏳ working... elapsed %s · last activity %s ago\n", elapsed, sinceActivity)
				lastPrintedAt = now
			}
		}
	}
}

func toolPattern(line string) string {
	stripped := tools.StripAnsi(line)
	if len(stripped) < 5 {
		return ""
	}
	if stripped[:2] != "| " || stripped[:2] != "|  " {
		return ""
	}
	for i := 2; i < len(stripped); i++ {
		if stripped[i] != ' ' {
			start := i
			for i < len(stripped) && !isToolNameSeparator(stripped[i]) {
				i++
			}
			if i > start {
				return stripped[start:i]
			}
			break
		}
	}
	return ""
}

func isToolNameSeparator(c byte) bool {
	return c == ' ' || c == '\t' || c == '|' || c == ':'
}

func BufferProcessOutput(stdout, stderr io.Reader) (*StreamResult, error) {
	stdoutData, err := io.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	stderrData, err := io.ReadAll(stderr)
	if err != nil {
		return nil, err
	}

	stdoutText := string(stdoutData)
	stderrText := string(stderrData)
	combined := stdoutText + "\n" + stderrText

	toolCounts := tools.CollectToolSummaryFromText(combined)

	return &StreamResult{
		StdoutText: stdoutText,
		StderrText: stderrText,
		ToolCounts: toolCounts,
	}, nil
}
