# gmcp — Unified Google Workspace CLI & MCP Server

**gmcp** is an all-in-one developer CLI and Model Context Protocol (MCP) server written in Go. It connects AI assistants (Claude Desktop, Cursor, Antigravity, Zed) and terminal workflows directly to Google Workspace services.

---

## ✨ Features

- **📬 Gmail:** Search, read full threads, send emails with attachments (`--attach`), reply to threads (`reply`), scheduled send (`--delay`, `--at`), and draft management (`draft`, `drafts`, `send-draft`, `delete-draft`).
- **📅 Google Calendar:** List upcoming events, natural language additions (`add`), structured creation with exact timestamps (`create`), attendee invitations, and complete RSVP response handling (`accept`, `decline`, `maybe`, `respond`).
- **📹 Google Meet:** Instantly provision persistent video conference rooms (`gmcp meet create`), send invitations directly via email (`gmcp meet send`), or attach to calendar events and emails with `--meet`.
- **📁 Google Drive:** Search files, read file contents, download files, upload local files, create files, permanently delete files, and auto-export Google Docs (to plain text/PDF) and Google Sheets (to CSV).
- **📋 Google Tasks:** Full lifecycle for task lists (`lists`, `create-list`, `delete-list`), todos with notes and due dates, hierarchical subtasks (`--parent`), attachment links (`--link`), and completion marking (`done`).
- **📊 Google Sheets:** Create spreadsheets (`create`), add sheet tabs (`add-sheet`), read cell ranges (`read`), append rows (`append`), and update cells (`update`).
- **📄 Google Docs:** Create documents (`create`), read document text (`read`), and append text (`append`).
- **▶️ YouTube:** Search videos, fetch view counts, likes, and metadata.
- **⚡ UNIX Composability:** All listing/querying commands support the `--json` flag to pipe directly into `jq`.
- **🚀 Single Static Binary:** Fast startup (~3ms), zero runtime dependencies, cross-platform Go architecture.

---

## 🛠️ Installation

```bash
# Clone and build
git clone https://github.com/erikkubica/gmcp.git
cd gmcp
go build -o ~/.local/bin/gmcp ./cmd/gmcp
```

Or install via `go install`:
```bash
go install github.com/erikkubica/gmcp/cmd/gmcp@latest
```

---

## 🔑 Authentication

1. Place your Google Cloud OAuth `credentials.json` at:
   `~/.config/gmcp/credentials.json`
2. Run the interactive browser login:
   ```bash
   gmcp auth login
   ```
3. Check authentication status anytime:
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

# Send email with attachment and auto-generated Google Meet link
gmcp mail send --to "client@example.com" \
  --subject "Project Kickoff" \
  --body "Looking forward to speaking." \
  --meet \
  --attach ./contract.pdf

# Schedule send
gmcp mail send --to "client@example.com" --subject "Update" --body "Hello" --delay 10m

# Reply to an existing thread
gmcp mail reply <message_id> --body "Thanks, let's meet tomorrow." --meet

# Manage drafts
gmcp mail draft --to "lead@company.com" --subject "Proposal" --body "Draft proposal content."
gmcp mail drafts
gmcp mail send-draft <draft_id>
gmcp mail delete-draft <draft_id>
```

### Google Meet & Calendar
```bash
# Instant Google Meet generation
gmcp meet create "Team Sync" --start "2026-10-06T15:00:00+07:00" --end "2026-10-06T15:30:00+07:00"

# Send Meet invitation via email
gmcp meet send "1-on-1 Catchup" --to "partner@example.com"

# View upcoming events (shows Google Meet links and RSVP status)
gmcp cal list --max 10 --json

# Natural language event creation
gmcp cal add "Lunch with Alex tomorrow at 1pm"

# Structured event with Google Meet and attendee invitations
gmcp cal create \
  --title "Client Architecture Review" \
  --start "2026-10-08T10:00:00+07:00" \
  --end "2026-10-08T11:00:00+07:00" \
  --meet \
  --attendees "lead@client.com,dev@client.com"

# RSVP to event invitations
gmcp cal accept <event_id>
gmcp cal decline <event_id>
gmcp cal maybe <event_id>

# Delete event
gmcp cal delete <event_id>
```

### Google Tasks
```bash
# List and manage task lists
gmcp tasks lists
gmcp tasks create-list "Q4 Roadmap"
gmcp tasks delete-list <list_id>

# List tasks (shows indented subtasks tree)
gmcp tasks list --list <list_id> --json

# Add task with notes, due date, and attachment link
gmcp tasks add "Review Sprint Backlog" \
  --notes "Prioritize auth features" \
  --due "2026-10-10T00:00:00.000Z" \
  --link "https://docs.google.com/spreadsheets/d/..."

# Add nested subtask
gmcp tasks add "Verify OAuth refresh token" --parent <parent_task_id>

# Mark task as completed
gmcp tasks done <task_id>
```

### Google Sheets & Docs
```bash
# Create spreadsheet & add sheets
gmcp sheets create "Project Budget"
gmcp sheets add-sheet <spreadsheet_id> "Expenses"

# Append row & update cells
gmcp sheets append <spreadsheet_id> "Expenses!A1" "2026-10-06" "Server Hosting" "45 EUR"
gmcp sheets update <spreadsheet_id> "Expenses!C1" "50 EUR"

# Read cell range
gmcp sheets read <spreadsheet_id> "Expenses!A1:D10" --json

# Create & read Google Docs
gmcp docs create "Meeting Notes"
gmcp docs read <doc_id>
gmcp docs append <doc_id> "Key decisions made during sprint kickoff.\n"
```

### Google Drive
```bash
# List / search Drive files
gmcp drive list --query "report" --json

# Read & export Google Docs / Sheets / text files
gmcp drive read <file_id>

# Download & upload files
gmcp drive download <file_id> ./downloaded_report.pdf
gmcp drive upload ./document.pdf --name "Final_Spec.pdf"

# Delete file
gmcp drive delete <file_id>
```

---

## 🤖 Model Context Protocol (MCP) Server

Run `gmcp` as a high-performance stdio MCP server for Claude Desktop, Cursor, Zed, or Antigravity:

```bash
gmcp serve
```

### Claude Desktop Configuration (`claude_desktop_config.json`)
```json
{
  "mcpServers": {
    "gmcp": {
      "command": "gmcp",
      "args": ["serve"]
    }
  }
}
```

### Available MCP Tools (28 Tools)

| Service | MCP Tools |
| :--- | :--- |
| **Gmail** | `gmail_list_messages`, `gmail_get_message`, `gmail_send_message`, `gmail_reply_message`, `gmail_create_draft`, `gmail_list_drafts`, `gmail_send_draft` |
| **Calendar** | `calendar_list_events`, `calendar_quick_add`, `calendar_create_event` (with Meet & attendees), `calendar_delete_event`, `calendar_respond_event` (RSVP) |
| **Meet** | `meet_create_session` |
| **Drive** | `drive_list_files`, `drive_read_file`, `drive_upload_file`, `drive_delete_file` |
| **Tasks** | `tasks_list`, `tasks_add` (with subtasks & links), `tasks_complete`, `tasks_delete`, `tasks_list_tasklists`, `tasks_create_tasklist`, `tasks_delete_tasklist` |
| **Sheets** | `sheets_read_range`, `sheets_append_row` |
| **Docs** | `docs_create_document`, `docs_read_document`, `docs_append_text` |
| **YouTube** | `youtube_search`, `youtube_video_details` |

---

## 📄 License

MIT
