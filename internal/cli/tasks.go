package cli

import (
	"context"
	"fmt"

	"github.com/erikkubica/gmcp/internal/services/tasks"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage Google Tasks (todos, milestones, checklists)",
}

var (
	taskMax   int64
	taskNotes string
	taskDue   string
	taskList  string
)

var tasksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(context.Background())
		if err != nil {
			return err
		}
		items, err := svc.ListTasks(taskList, taskMax)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}
		for _, t := range items {
			icon := "[ ]"
			if t.Status == "completed" {
				icon = "[x]"
			}
			fmt.Printf("%s %s (ID: %s)\n", icon, t.Title, t.ID)
			if t.Notes != "" {
				fmt.Printf("    Notes: %s\n", t.Notes)
			}
			if t.Due != "" {
				fmt.Printf("    Due: %s\n", t.Due)
			}
		}
		return nil
	},
}

var tasksAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(context.Background())
		if err != nil {
			return err
		}
		t, err := svc.CreateTask(taskList, args[0], taskNotes, taskDue)
		if err != nil {
			return err
		}
		fmt.Printf("Task created: %s (ID: %s)\n", t.Title, t.Id)
		return nil
	},
}

var tasksDoneCmd = &cobra.Command{
	Use:   "done [task_id]",
	Short: "Mark a task as completed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(context.Background())
		if err != nil {
			return err
		}
		t, err := svc.CompleteTask(taskList, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Task completed: %s\n", t.Title)
		return nil
	},
}

var tasksDeleteCmd = &cobra.Command{
	Use:   "delete [task_id]",
	Short: "Delete a task by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(context.Background())
		if err != nil {
			return err
		}
		if err := svc.DeleteTask(taskList, args[0]); err != nil {
			return err
		}
		fmt.Printf("Task %s deleted successfully.\n", args[0])
		return nil
	},
}

func init() {
	tasksListCmd.Flags().Int64VarP(&taskMax, "max", "m", 20, "Max tasks")
	tasksListCmd.Flags().StringVarP(&taskList, "list", "l", "@default", "Task list ID")

	tasksAddCmd.Flags().StringVarP(&taskNotes, "notes", "n", "", "Task notes/description")
	tasksAddCmd.Flags().StringVarP(&taskDue, "due", "d", "", "Due date RFC3339 (e.g. 2026-10-10T00:00:00.000Z)")
	tasksAddCmd.Flags().StringVarP(&taskList, "list", "l", "@default", "Task list ID")

	tasksDoneCmd.Flags().StringVarP(&taskList, "list", "l", "@default", "Task list ID")
	tasksDeleteCmd.Flags().StringVarP(&taskList, "list", "l", "@default", "Task list ID")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksAddCmd)
	tasksCmd.AddCommand(tasksDoneCmd)
	tasksCmd.AddCommand(tasksDeleteCmd)
}
