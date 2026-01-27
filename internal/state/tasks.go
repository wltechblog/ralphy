package state

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Task struct {
	Text         string `json:"text"`
	Status       string `json:"status"` // "todo", "in-progress", "complete"
	Subtasks     []Task `json:"subtasks"`
	OriginalLine string `json:"originalLine"`
}

var (
	topLevelTaskRegex = regexp.MustCompile(`^- \[([ x/])\]\s*(.+)`)
	subtaskRegex      = regexp.MustCompile(`^\s+- \[([ x/])\]\s*(.+)`)
)

func ParseTasks(content string) []Task {
	var tasks []Task
	lines := strings.Split(content, "\n")
	var currentTask *Task

	for _, line := range lines {
		if matches := topLevelTaskRegex.FindStringSubmatch(line); matches != nil {
			if currentTask != nil {
				tasks = append(tasks, *currentTask)
			}
			statusChar := matches[1]
			text := matches[2]
			status := "todo"
			if statusChar == "x" {
				status = "complete"
			} else if statusChar == "/" {
				status = "in-progress"
			}
			currentTask = &Task{
				Text:         text,
				Status:       status,
				Subtasks:     []Task{},
				OriginalLine: line,
			}
			continue
		}

		if matches := subtaskRegex.FindStringSubmatch(line); matches != nil {
			if currentTask != nil {
				statusChar := matches[1]
				text := matches[2]
				status := "todo"
				if statusChar == "x" {
					status = "complete"
				} else if statusChar == "/" {
					status = "in-progress"
				}
				currentTask.Subtasks = append(currentTask.Subtasks, Task{
					Text:         text,
					Status:       status,
					Subtasks:     []Task{},
					OriginalLine: line,
				})
			}
		}
	}

	if currentTask != nil {
		tasks = append(tasks, *currentTask)
	}

	return tasks
}

func LoadTasks() ([]Task, string, error) {
	path := filepath.Join(stateDirName, tasksFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", err
	}
	content := string(data)
	return ParseTasks(content), content, nil
}

func SaveTasks(content string) error {
	if err := os.MkdirAll(stateDirName, 0755); err != nil {
		return err
	}
	path := filepath.Join(stateDirName, tasksFileName)
	return os.WriteFile(path, []byte(content), 0644)
}

func FindCurrentTask(tasks []Task) *Task {
	for _, task := range tasks {
		if task.Status == "in-progress" {
			return &task
		}
	}
	return nil
}

func FindNextTask(tasks []Task) *Task {
	for _, task := range tasks {
		if task.Status == "todo" {
			return &task
		}
	}
	return nil
}

func AllTasksComplete(tasks []Task) bool {
	if len(tasks) == 0 {
		return false
	}
	for _, task := range tasks {
		if task.Status != "complete" {
			return false
		}
	}
	return true
}

func AddTask(description string) error {
	_, content, err := LoadTasks()
	if err != nil {
		return err
	}

	if content == "" {
		content = "# Ralph Tasks\n\n"
	}

	content = strings.TrimRight(content, "\n") + "\n" + fmt.Sprintf("- [ ] %s\n", description)
	return SaveTasks(content)
}

func RemoveTask(index int) error {
	tasks, content, err := LoadTasks()
	if err != nil {
		return err
	}

	if index < 1 || index > len(tasks) {
		return fmt.Errorf("task index %d out of range (1-%d)", index, len(tasks))
	}

	lines := strings.Split(content, "\n")
	var newLines []string
	inRemovedTask := false
	currentTaskLine := 0

	for _, line := range lines {
		if topLevelTaskRegex.MatchString(line) {
			currentTaskLine++
			if currentTaskLine == index {
				inRemovedTask = true
				continue
			} else {
				inRemovedTask = false
			}
		}

		if inRemovedTask && strings.HasPrefix(line, "    ") {
			continue
		}
		
		// Also handle tabs or different indentation if needed, but the original used spaces
		if inRemovedTask && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && strings.TrimSpace(line) != "" {
			continue
		}

		newLines = append(newLines, line)
	}

	return SaveTasks(strings.Join(newLines, "\n"))
}

func GetTasksModeSection(s *RalphState) string {
	tasks, tasksContent, err := LoadTasks()
	if err != nil || tasksContent == "" {
		return "\n## TASKS MODE: Enabled (no tasks file found)\n\nCreate .opencode/ralph-tasks.md with your task list, or use `ralphy --add-task \"description\"` to add tasks.\n"
	}

	currentTask := FindCurrentTask(tasks)
	nextTask := FindNextTask(tasks)

	var taskInstructions string
	if currentTask != nil {
		taskInstructions = fmt.Sprintf(`
🔄 CURRENT TASK: "%s"
   Focus on completing this specific task.
   When done: Mark as [x] in .opencode/ralph-tasks.md and output <promise>%s</promise>`, currentTask.Text, s.TaskPromise)
	} else if nextTask != nil {
		taskInstructions = fmt.Sprintf(`
📍 NEXT TASK: "%s"
   Mark as [/] in .opencode/ralph-tasks.md before starting.
   When done: Mark as [x] and output <promise>%s</promise>`, nextTask.Text, s.TaskPromise)
	} else if AllTasksComplete(tasks) {
		taskInstructions = fmt.Sprintf(`
✅ ALL TASKS COMPLETE!
   Output <promise>%s</promise> to finish.`, s.CompletionPromise)
	} else {
		taskInstructions = "\n📋 No tasks found. Add tasks to .opencode/ralph-tasks.md or use `ralphy --add-task`"
	}

	return fmt.Sprintf(`
## TASKS MODE: Working through task list

Current tasks from .opencode/ralph-tasks.md:
%s%s%s
%s

### Task Workflow
1. Find any task marked [/] (in progress). If none, pick the first [ ] task.
2. Mark the task as [/] in ralph-tasks.md before starting.
3. Complete the task.
4. Mark as [x] when verified complete.
5. Output <promise>%s</promise> to move to the next task.
6. Only output <promise>%s</promise> when ALL tasks are [x].

---
`, "```markdown\n", strings.TrimSpace(tasksContent), "\n```", taskInstructions, s.TaskPromise, s.CompletionPromise)
}
