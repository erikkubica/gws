package cli

import (
	"encoding/json"
	"fmt"

	"github.com/erikkubica/gws/internal/services/tasks"
	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage Google Tasks (todos, milestones, checklists)",
}

var (
	taskMax    int64
	taskNotes  string
	taskDue    string
	taskList   string
	taskParent string
	taskLink   string
	taskJSON   bool
)

func printTaskItem(t tasks.TaskSummary) {
	icon := "[ ]"
	if t.Status == "completed" {
		icon = "[x]"
	}
	prefix := ""
	if t.Parent != "" {
		prefix = "  └─ "
	}
	fmt.Printf("%s%s %s (ID: %s)\n", prefix, icon, t.Title, t.ID)
	if t.Notes != "" {
		fmt.Printf("%s   Notes: %s\n", prefix, t.Notes)
	}
	if t.Due != "" {
		fmt.Printf("%s   Due: %s\n", prefix, t.Due)
	}
}

func buildTaskNotes(notes, link string) string {
	if link == "" {
		return notes
	}
	if notes == "" {
		return "Attachment: " + link
	}
	return notes + "\nAttachment: " + link
}

var tasksListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List tasks (aliases: ls)",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(cmd.Context())
		if err != nil {
			return err
		}
		items, err := svc.ListTasks(taskList, taskMax)
		if err != nil {
			return err
		}
		if taskJSON {
			b, _ := json.MarshalIndent(items, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		if len(items) == 0 {
			fmt.Println("No tasks found.")
			return nil
		}
		for _, t := range items {
			printTaskItem(t)
		}
		return nil
	},
}

var tasksAddCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(cmd.Context())
		if err != nil {
			return err
		}
		notes := buildTaskNotes(taskNotes, taskLink)
		t, err := svc.CreateTask(taskList, taskParent, args[0], notes, taskDue)
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
		svc, err := tasks.NewService(cmd.Context())
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
	Use:     "delete [task_id]",
	Aliases: []string{"rm"},
	Short:   "Delete a task by ID (aliases: rm)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(cmd.Context())
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

var tasksListsCmd = &cobra.Command{
	Use:   "lists",
	Short: "List all task lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(cmd.Context())
		if err != nil {
			return err
		}
		lists, err := svc.ListTaskLists()
		if err != nil {
			return err
		}
		if taskJSON {
			b, _ := json.MarshalIndent(lists, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		for _, l := range lists {
			fmt.Printf("- %s (ID: %s)\n", l.Title, l.ID)
		}
		return nil
	},
}

var tasksCreateListCmd = &cobra.Command{
	Use:   "create-list [title]",
	Short: "Create a new task list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(cmd.Context())
		if err != nil {
			return err
		}
		tl, err := svc.CreateTaskList(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Task list created: %s (ID: %s)\n", tl.Title, tl.Id)
		return nil
	},
}

var tasksDeleteListCmd = &cobra.Command{
	Use:     "delete-list [list_id]",
	Aliases: []string{"rm-list"},
	Short:   "Delete a task list by ID (aliases: rm-list)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := tasks.NewService(cmd.Context())
		if err != nil {
			return err
		}
		if err := svc.DeleteTaskList(args[0]); err != nil {
			return err
		}
		fmt.Printf("Task list %s deleted successfully.\n", args[0])
		return nil
	},
}

func init() {
	tasksListCmd.Flags().Int64VarP(&taskMax, "max", "m", 20, "Max tasks")
	tasksListCmd.Flags().StringVarP(&taskList, "list", "l", "@default", "Task list ID")
	tasksListCmd.Flags().BoolVar(&taskJSON, "json", false, "Output as JSON")

	tasksAddCmd.Flags().StringVarP(&taskNotes, "notes", "n", "", "Task notes/description")
	tasksAddCmd.Flags().StringVarP(&taskDue, "due", "d", "", "Due date RFC3339 (e.g. 2026-10-10T00:00:00.000Z)")
	tasksAddCmd.Flags().StringVarP(&taskList, "list", "l", "@default", "Task list ID")
	tasksAddCmd.Flags().StringVar(&taskParent, "parent", "", "Parent task ID for subtasks")
	tasksAddCmd.Flags().StringVar(&taskLink, "link", "", "Attachment link / URL to include in notes")

	tasksListsCmd.Flags().BoolVar(&taskJSON, "json", false, "Output as JSON")

	tasksDoneCmd.Flags().StringVarP(&taskList, "list", "l", "@default", "Task list ID")
	tasksDeleteCmd.Flags().StringVarP(&taskList, "list", "l", "@default", "Task list ID")

	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksAddCmd)
	tasksCmd.AddCommand(tasksDoneCmd)
	tasksCmd.AddCommand(tasksDeleteCmd)
	tasksCmd.AddCommand(tasksListsCmd)
	tasksCmd.AddCommand(tasksCreateListCmd)
	tasksCmd.AddCommand(tasksDeleteListCmd)
}
