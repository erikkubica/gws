# gmcp — Unified Google Workspace CLI & MCP Server

**gmcp** is an all-in-one developer CLI and Model Context Protocol (MCP) server written in Go. It connects AI assistants (Claude Desktop, Cursor, Antigravity, Zed) and terminal workflows directly to Google Workspace services.

---

## ✨ Features

- **📬 Gmail:** Search, read full threads, send emails with attachments (`--attach`), create drafts, and reply to existing threads (`reply`).
- **📅 Google Calendar:** List upcoming events, natural language additions (`quick_add`), structured creation with exact timestamps, and event deletion.
- **📁 Google Drive:** Search files, read file contents, upload local files, permanently delete files, and auto-export Google Docs (to plain text) and Google Sheets (to CSV).
- **📊 Google Sheets:** Read cell ranges (`read`) and append rows (`append`).
- **✅ Google Tasks:** List tasks, create todos with notes and due dates, mark as completed, and delete tasks.
- **▶️ YouTube:** Search videos, fetch view counts, likes, and metadata.
- **⚡ UNIX Composability:** All listing/querying commands support the `--json` flag to pipe directly into `jq`.
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
# List recent emails (plain text or JSON)
gmcp mail list --max 5
gmcp mail list --max 5 --json | jq .

# Search specific emails
gmcp mail list --query "from:recruiter is:unread"

# Send an email with attachment
gmcp mail send --to "client@example.com" \
  --subject "Senior Developer Application" \
  --body "Attached is my CV." \
  --attach ~/Documents/erik-kubica-cv.pdf

# Reply to an existing email thread
gmcp mail reply <message_id> \
  --body "Dobrý deň, ďakujem za odpoveď. V prílohe posielam CV." \
  --attach ~/Documents/erik-kubica-cv.pdf

# Create a draft
gmcp mail draft --to "client@example.com" --subject "Proposal" --body "Draft proposal content."
```

### Google Sheets
```bash
# Read cell range (e.g. A1:E10)
gmcp sheets read <spreadsheet_id> "Sheet1!A1:E10"
gmcp sheets read <spreadsheet_id> "Sheet1!A1:E10" --json

# Append a row of values
gmcp sheets append <spreadsheet_id> "Sheet1!A1" "2026-10-06" "GoodRequest" "Contacted" "25 EUR/h"
```

### Calendar
```bash
# View upcoming events (with JSON support)
gmcp cal list --max 10 --json

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
gmcp tasks list --json

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
gmcp drive list --query "CV" --json

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
gmcp yt search "golang mcp" --max 5 --json

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

### Available MCP Tools (21 Tools)

| Service | MCP Tools |
| :--- | :--- |
| **Gmail** | `gmail_list_messages`, `gmail_get_message`, `gmail_send_message` (with attachment), `gmail_reply_message`, `gmail_create_draft` |
| **Sheets** | `sheets_read_range`, `sheets_append_row` |
| **Calendar** | `calendar_list_events`, `calendar_quick_add`, `calendar_create_event`, `calendar_delete_event` |
| **Tasks** | `tasks_list`, `tasks_add`, `tasks_complete`, `tasks_delete` |
| **Drive** | `drive_list_files`, `drive_read_file`, `drive_upload_file`, `drive_delete_file` |
| **YouTube** | `youtube_search`, `youtube_video_details` |
