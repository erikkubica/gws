# gmcp — Unified Google Workspace CLI & MCP Server

**gmcp** is an all-in-one developer CLI and Model Context Protocol (MCP) server written in Go. It connects AI assistants (Claude Desktop, Cursor, Antigravity, Zed) and terminal workflows directly to Google Workspace services.

---

## ✨ Features

- **📬 Gmail:** Search, read full threads, send emails, and create drafts.
- **📅 Google Calendar:** List upcoming events, parse natural language additions (`quick_add`).
- **📁 Google Drive:** Search files, read file contents, and auto-export Google Docs (to plain text) and Google Sheets (to CSV).
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
   This starts an ephemeral local server, opens your default browser for Google OAuth2 consent, captures the authorization code, and saves the auto-refreshing token to `~/.config/gmcp/token.json`.
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
```

### Drive
```bash
# List / search Drive files
gmcp drive list --query "CV"

# Read Google Doc / Sheet / file content
gmcp drive read <file_id>
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

To use `gmcp` with **Claude Desktop**, **Cursor**, or **Antigravity**, add the server configuration:

### Claude Desktop (`claude_desktop_config.json`)
```json
{
  "mcpServers": {
    "google-workspace": {
      "command": "/usr/local/bin/gmcp",
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
| `drive_list_files` | `query`, `max` | Search and list files in Google Drive |
| `drive_read_file` | `file_id` (required) | Read/export document and sheet content |
| `youtube_search` | `query` (required), `max`| Search YouTube videos |
| `youtube_video_details` | `video_id` (required) | Fetch stats, view counts, and details |

---

## 🏗️ Architecture

```text
gmcp/
├── cmd/
│   └── gmcp/main.go          # Entrypoint
├── internal/
│   ├── auth/                 # OAuth2 loopback server & token persistence
│   ├── cli/                  # Cobra commands (auth, mail, cal, drive, yt, serve)
│   ├── services/             # Google API service adapters
│   │   ├── gmail/
│   │   ├── calendar/
│   │   ├── drive/
│   │   └── youtube/
│   └── mcp/                  # mark3labs/mcp-go stdio tool definitions
└── bin/
    └── gmcp                  # Compiled static binary
```
