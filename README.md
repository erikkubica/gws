# gmcp — Unified Google Workspace CLI & MCP Server

**gmcp** is an all-in-one developer CLI and Model Context Protocol (MCP) server written in Go. It connects AI assistants (Claude Desktop, Cursor, Antigravity, Zed) and terminal workflows directly to Google Workspace services.

---

## ✨ Features

- **📬 Gmail:** Search, read full threads, send emails, and create drafts.
- **📅 Google Calendar:** List upcoming events, natural language additions (`quick_add`), structured creation with exact timestamps, and event deletion.
- **📁 Google Drive:** Search files, read file contents, upload local files, permanently delete files, and auto-export Google Docs (to plain text) and Google Sheets (to CSV).
- **✅ Google Tasks:** List tasks, create todos with notes and due dates, mark as completed, and delete tasks.
- **▶️ YouTube:** Search videos, fetch view counts, likes, and metadata.
- **⚡ Dual Mode:** Functions both as a terminal CLI tool (`gmcp mail ...`) and an stdio MCP Server (`gmcp serve`).
- **🚀 Single Static Binary:** Fast startup (~3ms), zero runtime dependencies, cross-platform.

---

## 🛠️ Installation

```bash
cd ~/.gemini/antigravity/scratch/gmcp
go build -o /usr/local/bin/gmcp ./cmd/gmcp
```

---

## 🔑 Authentication

1. Make sure your `credentials.json` is located at:
   `~/.config/gmcp/credentials.json`
2. Run the interactive browser login:
   ```bash
   gmcp auth login
   ```
3. Check status anytime:
   ```bash
   gmcp auth status
   ```

---

## 💻 CLI Usage

### Gmail
```bash
# List recent emails
gmcp mail list --max 5

# Search specific emails
gmcp mail list --query "from:recruiter is:unread"

# Read message by ID
gmcp mail read <message_id>

# Send an email
gmcp mail send --to "client@example.com" --subject "Project Update" --body "Everything is deployed."

# Create a draft
gmcp mail draft --to "client@example.com" --subject "Proposal" --body "Draft proposal content."
```

### Calendar
```bash
# View upcoming events
gmcp cal list --max 10

# Natural language event creation
gmcp cal add "Coffee with Marek on Friday at 10am"

# Structured event creation with exact RFC3339 timestamps
gmcp cal create --title "Technical Interview" --start "2026-10-08T10:00:00+07:00" --end "2026-10-08T11:00:00+07:00" --loc "Google Meet"

# Delete event
gmcp cal delete <event_id>
```

### Google Tasks
```bash
# List all tasks
gmcp tasks list

# Add a new task with notes and due date
gmcp tasks add "Review B2B agency contract" --notes "Verify hourly rate and payment terms" --due "2026-10-10T00:00:00.000Z"

# Mark task as completed
gmcp tasks done <task_id>

# Delete task
gmcp tasks delete <task_id>
```

### Drive
```bash
# List / search Drive files
gmcp drive list --query "CV"

# Read Google Doc / Sheet / file content
gmcp drive read <file_id>

# Upload a local file
gmcp drive upload ./report.pdf --name "Final_Report.pdf"

# Delete a file from Drive
gmcp drive delete <file_id>
```

### YouTube
```bash
# Search videos
gmcp yt search "Golang MCP server tutorial" --max 5

# Video statistics
gmcp yt info <video_id>
```

---

## 🤖 MCP Server Setup

Add this to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "google-workspace": {
      "command": "gmcp",
      "args": ["serve"]
    }
  }
}
```

### Available MCP Tools

| Tool | Parameters | Description |
| :--- | :--- | :--- |
| `gmail_list_messages` | `query`, `max` | Search and list Gmail messages |
| `gmail_get_message` | `id` (required) | Read complete body and headers of an email |
| `gmail_send_message` | `to`, `subject`, `body` | Send an email directly |
| `gmail_create_draft` | `to`, `subject`, `body` | Create a Gmail draft |
| `calendar_list_events`| `max`, `calendar_id` | List upcoming calendar schedule |
| `calendar_quick_add` | `text` (required) | Natural language calendar event creation |
| `calendar_create_event`| `title`, `start`, `end`, `desc`, `loc` | Structured calendar event creation |
| `calendar_delete_event`| `event_id` (required) | Delete a calendar event |
| `tasks_list` | `max`, `list_id` | List tasks and todos |
| `tasks_add` | `title` (required), `notes`, `due` | Add a new task |
| `tasks_complete` | `task_id` (required) | Mark task as completed |
| `tasks_delete` | `task_id` (required) | Delete a task |
| `drive_list_files` | `query`, `max` | Search and list files in Google Drive |
| `drive_read_file` | `file_id` (required) | Read/export document and sheet content |
| `drive_upload_file` | `path` (required), `name` | Upload a local file to Drive |
| `drive_delete_file` | `file_id` (required) | Permanently delete a file from Drive |
| `youtube_search` | `query` (required), `max`| Search YouTube videos |
| `youtube_video_details` | `video_id` (required) | Fetch stats, view counts, and details |
