package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrStateNotFound = errors.New("state file not found")
)

func GetStateDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, stateDirName), nil
}

func getStateDir() (string, error) {
	return GetStateDir()
}

func EnsureStateDir() error {
	stateDir, err := GetStateDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(stateDir, 0755)
}

func ensureStateDir() error {
	return EnsureStateDir()
}

func getStatePath() (string, error) {
	stateDir, err := getStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, stateFileName), nil
}

func LoadState() (*RalphState, error) {
	statePath, err := getStatePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrStateNotFound
		}
		return nil, err
	}

	var state RalphState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func SaveState(state *RalphState) error {
	if err := ensureStateDir(); err != nil {
		return err
	}

	statePath, err := getStatePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

func ClearState() error {
	statePath, err := getStatePath()
	if err != nil {
		return err
	}

	if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
