package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/tasks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTasksTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildTasksListTool(), handleListTasks(ctx))
	s.AddTool(buildTasksAddTool(), handleAddTask(ctx))
	s.AddTool(buildTasksDoneTool(), handleDoneTask(ctx))
	s.AddTool(buildTasksDeleteTool(), handleDeleteTask(ctx))
}

func buildTasksListTool() mcp.Tool {
	return mcp.NewTool("tasks_list",
		mcp.WithDescription("List tasks and todos from Google Tasks"),
		mcp.WithNumber("max", mcp.Description("Max tasks to return (default 20)")),
		mcp.WithString("list_id", mcp.Description("Task list ID (default '@default')")),
	)
}

func buildTasksAddTool() mcp.Tool {
	return mcp.NewTool("tasks_add",
		mcp.WithDescription("Add a new task or todo to Google Tasks"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
		mcp.WithString("notes", mcp.Description("Task description or notes")),
		mcp.WithString("due", mcp.Description("Due date in RFC3339 (e.g. '2026-10-10T00:00:00.000Z')")),
		mcp.WithString("list_id", mcp.Description("Task list ID (default '@default')")),
	)
}

func buildTasksDoneTool() mcp.Tool {
	return mcp.NewTool("tasks_complete",
		mcp.WithDescription("Mark a task as completed in Google Tasks"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("The ID of the task to complete")),
		mcp.WithString("list_id", mcp.Description("Task list ID (default '@default')")),
	)
}

func buildTasksDeleteTool() mcp.Tool {
	return mcp.NewTool("tasks_delete",
		mcp.WithDescription("Permanently delete a task from Google Tasks"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("The ID of the task to delete")),
		mcp.WithString("list_id", mcp.Description("Task list ID (default '@default')")),
	)
}

func handleListTasks(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		max := int64(req.GetInt("max", 20))
		listID := req.GetString("list_id", "@default")
		items, err := svc.ListTasks(listID, max)
		if err != nil {
			return mcp.NewToolResultError("tasks error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(items, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleAddTask(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		title, _ := req.RequireString("title")
		notes := req.GetString("notes", "")
		due := req.GetString("due", "")
		listID := req.GetString("list_id", "@default")

		t, err := svc.CreateTask(listID, title, notes, due)
		if err != nil {
			return mcp.NewToolResultError("create task error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Task '%s' created (ID: %s)", t.Title, t.Id)), nil
	}
}

func handleDoneTask(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		taskID, _ := req.RequireString("task_id")
		listID := req.GetString("list_id", "@default")

		t, err := svc.CompleteTask(listID, taskID)
		if err != nil {
			return mcp.NewToolResultError("complete task error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Task '%s' marked as completed", t.Title)), nil
	}
}

func handleDeleteTask(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(ctx)
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		taskID, _ := req.RequireString("task_id")
		listID := req.GetString("list_id", "@default")

		if err := svc.DeleteTask(listID, taskID); err != nil {
			return mcp.NewToolResultError("delete task error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Task %s deleted successfully", taskID)), nil
	}
}
