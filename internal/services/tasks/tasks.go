package tasks

import (
	"fmt"

	"google.golang.org/api/tasks/v1"
)

// TaskSummary holds minimal task metadata.
type TaskSummary struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Notes   string `json:"notes,omitempty"`
	Status  string `json:"status"`
	Due     string `json:"due,omitempty"`
	Updated string `json:"updated,omitempty"`
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
		list = append(list, TaskSummary{
			ID:      t.Id,
			Title:   t.Title,
			Notes:   t.Notes,
			Status:  t.Status,
			Due:     t.Due,
			Updated: t.Updated,
		})
	}
	return list, nil
}

// CreateTask adds a new task with title, notes, and optional due date (RFC3339).
func (s *Service) CreateTask(listID, title, notes, due string) (*tasks.Task, error) {
	if listID == "" {
		listID = "@default"
	}
	task := &tasks.Task{
		Title: title,
		Notes: notes,
		Due:   due,
	}
	res, err := s.client.Tasks.Insert(listID, task).Do()
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return res, nil
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
