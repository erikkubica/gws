package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/services/tasks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTasksTools(ctx context.Context, s *server.MCPServer) {
	s.AddTool(buildTasksListTool(), handleListTasks(ctx))
	s.AddTool(buildTasksAddTool(), handleAddTask(ctx))
	s.AddTool(buildTasksDoneTool(), handleDoneTask(ctx))
	s.AddTool(buildTasksDeleteTool(), handleDeleteTask(ctx))
	s.AddTool(buildTasksListListsTool(), handleListTaskLists(ctx))
	s.AddTool(buildTasksCreateListTool(), handleCreateTaskList(ctx))
	s.AddTool(buildTasksDeleteListTool(), handleDeleteTaskList(ctx))
}

func buildTasksListTool() mcp.Tool {
	return mcp.NewTool("tasks_list",
		mcp.WithDescription("List tasks and todos from Google Tasks"),
		mcp.WithNumber("max", mcp.Description("Max tasks to return (default 20)")),
		mcp.WithString("list_id", mcp.Description("Task list ID (default '@default')")),
		accountOption(),
	)
}

func buildTasksAddTool() mcp.Tool {
	return mcp.NewTool("tasks_add",
		mcp.WithDescription("Add a new task or todo to Google Tasks"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
		mcp.WithString("notes", mcp.Description("Task description or notes")),
		mcp.WithString("due", mcp.Description("Due date in RFC3339 (e.g. '2026-10-10T00:00:00.000Z')")),
		mcp.WithString("list_id", mcp.Description("Task list ID (default '@default')")),
		mcp.WithString("parent_id", mcp.Description("Parent task ID for creating a subtask")),
		mcp.WithString("link", mcp.Description("Attachment link or URL to append to notes")),
		accountOption(),
	)
}

func buildTasksDoneTool() mcp.Tool {
	return mcp.NewTool("tasks_complete",
		mcp.WithDescription("Mark a task as completed in Google Tasks"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("The ID of the task to complete")),
		mcp.WithString("list_id", mcp.Description("Task list ID (default '@default')")),
		accountOption(),
	)
}

func buildTasksDeleteTool() mcp.Tool {
	return mcp.NewTool("tasks_delete",
		mcp.WithDescription("Permanently delete a task from Google Tasks"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("The ID of the task to delete")),
		mcp.WithString("list_id", mcp.Description("Task list ID (default '@default')")),
		accountOption(),
	)
}

func buildTasksListListsTool() mcp.Tool {
	return mcp.NewTool("tasks_list_tasklists",
		mcp.WithDescription("List all task lists belonging to the user"),
		accountOption(),
	)
}

func buildTasksCreateListTool() mcp.Tool {
	return mcp.NewTool("tasks_create_tasklist",
		mcp.WithDescription("Create a new task list"),
		mcp.WithString("title", mcp.Required(), mcp.Description("Task list title")),
		accountOption(),
	)
}

func buildTasksDeleteListTool() mcp.Tool {
	return mcp.NewTool("tasks_delete_tasklist",
		mcp.WithDescription("Delete a task list by ID"),
		mcp.WithString("list_id", mcp.Required(), mcp.Description("Task list ID to delete")),
		accountOption(),
	)
}

func handleListTasks(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(withAccountContext(c, req))
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
		svc, err := tasks.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		notes := buildTaskNotes(req)
		listID := req.GetString("list_id", "@default")
		parentID := req.GetString("parent_id", "")
		due := req.GetString("due", "")

		t, err := svc.CreateTask(listID, parentID, title, notes, due)
		if err != nil {
			return mcp.NewToolResultError("create task error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Task '%s' created (ID: %s)", t.Title, t.Id)), nil
	}
}

func buildTaskNotes(req mcp.CallToolRequest) string {
	notes := req.GetString("notes", "")
	if link := req.GetString("link", ""); link != "" {
		if notes != "" {
			notes += "\n"
		}
		notes += "Attachment: " + link
	}
	return notes
}

func handleDoneTask(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
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
		svc, err := tasks.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		listID := req.GetString("list_id", "@default")

		if err := svc.DeleteTask(listID, taskID); err != nil {
			return mcp.NewToolResultError("delete task error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Task %s deleted successfully", taskID)), nil
	}
}

func handleListTaskLists(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		lists, err := svc.ListTaskLists()
		if err != nil {
			return mcp.NewToolResultError("list tasklists error: " + err.Error()), nil
		}
		b, _ := json.MarshalIndent(lists, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	}
}

func handleCreateTaskList(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		tl, err := svc.CreateTaskList(title)
		if err != nil {
			return mcp.NewToolResultError("create tasklist error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Task list '%s' created (ID: %s)", tl.Title, tl.Id)), nil
	}
}

func handleDeleteTaskList(ctx context.Context) server.ToolHandlerFunc {
	return func(c context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		svc, err := tasks.NewService(withAccountContext(c, req))
		if err != nil {
			return mcp.NewToolResultError("auth error: " + err.Error()), nil
		}
		listID, err := req.RequireString("list_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := svc.DeleteTaskList(listID); err != nil {
			return mcp.NewToolResultError("delete tasklist error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Task list %s deleted successfully", listID)), nil
	}
}
