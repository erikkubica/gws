package tasks

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/tasks/v1"
)

// TaskSummary holds minimal task metadata.
type TaskSummary struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Notes   string `json:"notes,omitempty"`
	Status  string `json:"status"`
	Due     string `json:"due,omitempty"`
	Parent  string `json:"parent,omitempty"`
	Updated string `json:"updated,omitempty"`
}

// TaskListSummary holds metadata for task lists.
type TaskListSummary struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Updated string `json:"updated,omitempty"`
}

// ListTaskLists returns all task lists belonging to the user.
func (s *Service) ListTaskLists() ([]TaskListSummary, error) {
	res, err := s.client.Tasklists.List().Do()
	if err != nil {
		return nil, fmt.Errorf("list tasklists: %w", err)
	}
	var lists []TaskListSummary
	for _, l := range res.Items {
		lists = append(lists, TaskListSummary{ID: l.Id, Title: l.Title, Updated: l.Updated})
	}
	return lists, nil
}

// CreateTaskList creates a new custom task list.
func (s *Service) CreateTaskList(title string) (*tasks.TaskList, error) {
	tl := &tasks.TaskList{Title: title}
	res, err := s.client.Tasklists.Insert(tl).Do()
	if err != nil {
		return nil, fmt.Errorf("create tasklist: %w", err)
	}
	return res, nil
}

// DeleteTaskList deletes a task list by ID.
func (s *Service) DeleteTaskList(tasklistID string) error {
	if err := s.client.Tasklists.Delete(tasklistID).Do(); err != nil {
		return fmt.Errorf("delete tasklist %s: %w", tasklistID, err)
	}
	return nil
}

// ListTasks lists tasks from a tasklist (defaulting to "@default").
func (s *Service) ListTasks(listID string, max int64) ([]TaskSummary, error) {
	if listID == "" {
		listID = "@default"
	}
	if max <= 0 {
		max = 20
	}
	res, err := s.client.Tasks.List(listID).MaxResults(max).ShowCompleted(true).ShowHidden(false).Do()
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	var list []TaskSummary
	for _, t := range res.Items {
		list = append(list, buildTaskSummary(t))
	}
	return list, nil
}

func buildTaskSummary(t *tasks.Task) TaskSummary {
	return TaskSummary{
		ID:      t.Id,
		Title:   t.Title,
		Notes:   t.Notes,
		Status:  t.Status,
		Due:     t.Due,
		Parent:  t.Parent,
		Updated: t.Updated,
	}
}

// CreateTask adds a new task with title, notes, due date, and optional parent subtask ID.
func (s *Service) CreateTask(listID, parentID, title, notes, due string) (*tasks.Task, error) {
	if listID == "" {
		listID = "@default"
	}
	task := &tasks.Task{Title: title, Notes: notes, Due: NormalizeDueDate(due)}
	call := s.client.Tasks.Insert(listID, task)
	if parentID != "" {
		call = call.Parent(parentID)
	}
	res, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return res, nil
}

// NormalizeDueDate converts YYYY-MM-DD dates to RFC3339 timestamps for Google Tasks API.
func NormalizeDueDate(due string) string {
	clean := strings.TrimSpace(due)
	if clean == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339, clean); err == nil {
		return t.Format(time.RFC3339)
	}
	if t, err := time.Parse("2006-01-02", clean); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	return clean
}

// CompleteTask marks a task as completed.
func (s *Service) CompleteTask(listID, taskID string) (*tasks.Task, error) {
	if listID == "" {
		listID = "@default"
	}
	task, err := s.client.Tasks.Get(listID, taskID).Do()
	if err != nil {
		return nil, fmt.Errorf("fetch task %s: %w", taskID, err)
	}
	task.Status = "completed"
	res, err := s.client.Tasks.Update(listID, taskID, task).Do()
	if err != nil {
		return nil, fmt.Errorf("update task status: %w", err)
	}
	return res, nil
}

// DeleteTask removes a task from Google Tasks.
func (s *Service) DeleteTask(listID, taskID string) error {
	if listID == "" {
		listID = "@default"
	}
	if err := s.client.Tasks.Delete(listID, taskID).Do(); err != nil {
		return fmt.Errorf("delete task %s: %w", taskID, err)
	}
	return nil
}
