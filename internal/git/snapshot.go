package git

import (
	"bytes"
	"os/exec"
	"strings"
)

type FileSnapshot struct {
	Files map[string]string
}

func CaptureFileSnapshot() (*FileSnapshot, error) {
	snapshot := &FileSnapshot{
		Files: make(map[string]string),
	}

	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, statusErr := statusCmd.Output()
	if statusErr != nil {
		return snapshot, nil
	}

	lsCmd := exec.Command("git", "ls-files")
	lsOutput, lsErr := lsCmd.Output()
	if lsErr != nil {
		return snapshot, nil
	}

	allFiles := make(map[string]bool)

	for _, line := range strings.Split(string(statusOutput), "\n") {
		if len(line) >= 4 {
			file := strings.TrimSpace(line[3:])
			if file != "" {
				allFiles[file] = true
			}
		}
	}

	for _, line := range strings.Split(string(lsOutput), "\n") {
		file := strings.TrimSpace(line)
		if file != "" {
			allFiles[file] = true
		}
	}

	for file := range allFiles {
		hash, err := getFileHash(file)
		if err == nil {
			snapshot.Files[file] = hash
		}
	}

	return snapshot, nil
}

func getFileHash(file string) (string, error) {
	cmd := exec.Command("git", "hash-object", file)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		cmd := exec.Command("stat", "-f", "'%m'", file)
		output, err = cmd.Output()
		if err != nil {
			return "", nil
		}
	}
	return strings.TrimSpace(string(output)), nil
}

func GetModifiedFilesSinceSnapshot(before *FileSnapshot, after *FileSnapshot) []string {
	var changedFiles []string

	for file, hash := range after.Files {
		prevHash, exists := before.Files[file]
		if !exists || prevHash != hash {
			changedFiles = append(changedFiles, file)
		}
	}

	for file := range before.Files {
		if _, exists := after.Files[file]; !exists {
			changedFiles = append(changedFiles, file)
		}
	}

	return changedFiles
}

func AutoCommit(message string) error {
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil
	}

	if strings.TrimSpace(string(statusOutput)) == "" {
		return nil
	}

	addCmd := exec.Command("git", "add", "-A")
	if err := addCmd.Run(); err != nil {
		return err
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	if err := commitCmd.Run(); err != nil {
		return err
	}

	return nil
}
