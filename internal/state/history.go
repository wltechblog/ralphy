package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrHistoryNotFound = errors.New("history file not found")
)

func getHistoryPath() (string, error) {
	stateDir, err := getStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, historyFileName), nil
}

func LoadHistory() (*RalphHistory, error) {
	historyPath, err := getHistoryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &RalphHistory{
				Iterations:      []IterationHistory{},
				TotalDurationMs: 0,
				StruggleIndicators: StruggleIndicators{
					RepeatedErrors:       map[string]int{},
					NoProgressIterations: 0,
					ShortIterations:      0,
				},
			}, nil
		}
		return nil, err
	}

	var history RalphHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return &RalphHistory{
			Iterations:      []IterationHistory{},
			TotalDurationMs: 0,
			StruggleIndicators: StruggleIndicators{
				RepeatedErrors:       map[string]int{},
				NoProgressIterations: 0,
				ShortIterations:      0,
			},
		}, nil
	}

	return &history, nil
}

func SaveHistory(history *RalphHistory) error {
	if err := ensureStateDir(); err != nil {
		return err
	}

	historyPath, err := getHistoryPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0644)
}

func ClearHistory() error {
	historyPath, err := getHistoryPath()
	if err != nil {
		return err
	}

	if err := os.Remove(historyPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func AddIteration(history *RalphHistory, iter *IterationHistory) {
	history.Iterations = append(history.Iterations, *iter)
	history.TotalDurationMs += iter.DurationMs
}

func UpdateStruggleIndicators(history *RalphHistory, iter *IterationHistory) {
	if len(iter.FilesModified) == 0 {
		history.StruggleIndicators.NoProgressIterations++
	} else {
		history.StruggleIndicators.NoProgressIterations = 0
	}

	if iter.DurationMs < 30000 {
		history.StruggleIndicators.ShortIterations++
	} else {
		history.StruggleIndicators.ShortIterations = 0
	}

	if len(iter.Errors) == 0 {
		history.StruggleIndicators.RepeatedErrors = map[string]int{}
	} else {
		for _, err := range iter.Errors {
			key := truncate(err, 100)
			history.StruggleIndicators.RepeatedErrors[key]++
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
