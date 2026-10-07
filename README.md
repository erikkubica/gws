# gws — Unified Google Workspace CLI & MCP Server

**gws** is an all-in-one developer CLI and Model Context Protocol (MCP) server written in Go. It connects AI assistants (Claude Desktop, Cursor, Antigravity, Zed) and terminal workflows directly to Google Workspace services.

---

## ✨ Features

- **📬 Gmail:** Search, read full threads, send emails with attachments (`--attach`), reply to threads (`reply`), scheduled send (`--delay`, `--at`), and draft management (`draft`, `drafts`, `send-draft`, `delete-draft`).
- **📅 Google Calendar:** List upcoming events, natural language additions (`add`), structured creation with exact timestamps (`create`), attendee invitations, and complete RSVP response handling (`accept`, `decline`, `maybe`, `respond`).
- **📹 Google Meet:** Instantly provision persistent video conference rooms (`gws meet create`), send invitations directly via email (`gws meet send`), or attach to calendar events and emails with `--meet`.
- **📁 Google Drive:** Search files, read file contents, download files, upload local files, create files, permanently delete files, and auto-export Google Docs (to plain text/PDF) and Google Sheets (to CSV).
- **📋 Google Tasks:** Full lifecycle for task lists (`lists`, `create-list`, `delete-list`), todos with notes and due dates, hierarchical subtasks (`--parent`), attachment links (`--link`), and completion marking (`done`).
- **📊 Google Sheets:** Create spreadsheets (`create`), add sheet tabs (`add-sheet`), read cell ranges (`read`), append rows (`append`), and update cells (`update`).
- **📄 Google Docs:** Create documents (`create`), read document text (`read`), and append text (`append`).
- **💬 Google Chat:** List spaces and direct messages (`spaces`), send messages (`send`), read chat history (`messages`), react with emojis (`react`), inspect reactions (`reactions`), and handle file attachments.
- **▶️ YouTube:** Search videos, fetch view counts, likes, and metadata.
- **⚡ UNIX Composability:** All listing/querying commands support the `--json` flag to pipe directly into `jq`.
- **🚀 Single Static Binary:** Fast startup (~3ms), zero runtime dependencies, cross-platform Go architecture.

---

## 🛠️ Installation

### Quick Install (Binary + Editor Plugins)

Run the included install script to build the binary to `~/.local/bin/gws` and install all editor plugins (or specify `--editor antigravity`):

```bash
git clone https://github.com/erikkubica/gws.git
cd gws
./install.sh
```

Or install a specific editor plugin directly:
```bash
./plugin/antigravity/install.sh
```

Or using `make`:
```bash
make install
```

### Manual Build

```bash
go build -o ~/.local/bin/gws ./cmd/gws
```

Or install via `go install`:
```bash
go install github.com/erikkubica/gws/cmd/gws@latest
```

---

## 🔑 Google Cloud Setup & Authentication

### 1. GCP Project Setup & API Enablement (Google Cloud Shell)

You can run these commands directly in [Google Cloud Shell](https://shell.cloud.google.com) or using the local `gcloud` CLI:

```bash
# 1. Set your project ID (replace with your desired or existing project ID)
export PROJECT_ID="my-gws-tools"

# 2. Create the GCP project (skip if using an existing project)
gcloud projects create $PROJECT_ID --name="Google Workspace CLI"

# 3. Set the active project
gcloud config set project $PROJECT_ID

# 4. Enable all required Google Workspace APIs in one command
gcloud services enable \
  gmail.googleapis.com \
  calendar-json.googleapis.com \
  drive.googleapis.com \
  tasks.googleapis.com \
  sheets.googleapis.com \
  docs.googleapis.com \
  chat.googleapis.com \
  youtube.googleapis.com \
  meet.googleapis.com
```

### 2. Configure OAuth Consent Screen & Credentials

1. **OAuth Consent Screen**:
   * Open [Google Cloud Console > OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent).
   * Choose **User Type**:
     * **Internal**: If using Google Workspace organization accounts (all team members can log in immediately with no verification needed).
     * **External**: If using standard `@gmail.com` accounts (keep Publishing Status in **Testing** and add your email addresses under **Test Users**).
   * Enter **App name** (e.g. `gws CLI`) and your **User support email**, then save.

2. **Create Desktop OAuth Client ID**:
   * Open [Google Cloud Console > Credentials](https://console.cloud.google.com/apis/credentials).
   * Click **Create Credentials** $\rightarrow$ **OAuth client ID**.
   * Application type: **Desktop app**.
   * Name: `gws Desktop Client`.
   * Click **Create** and download the credentials JSON file (or copy the Client ID & Secret).

3. **Google Chat API Configuration (Required for `gws chat`)**:
   * Even when posting as a user, Google Chat requires an app profile for the client attribution pill:
   * Open [Google Cloud Console > Google Chat API > Configuration](https://console.cloud.google.com/apis/api/chat.googleapis.com/hangouts-chat).
   * Enter **App name** (e.g. `gws`), an **Avatar URL** (e.g. `https://developers.google.com/workspace/chat/images/quickstart-app-avatar.png`), and **Description** (`Google Workspace CLI`).
   * Leave **Interactive features** toggled OFF.
   * Under **Visibility**, choose *"Make this Chat app available to everyone in your domain"* (or add your email).
   * Click **Save**.

### 3. Configure GCP OAuth Application (`gws gcp`)

Import the OAuth Client ID JSON downloaded from Google Cloud Console:
```bash
gws gcp import ./credentials.json
```

Or configure the credentials directly without needing a file (ideal for VPS/remote setups):
```bash
gws gcp set "<client_id>" "<client_secret>"
```

Verify your GCP application configuration:
```bash
gws gcp status
```

### 4. User Authentication & Multi-Account Support (`gws auth`)

Authenticate your Google account via browser:
```bash
gws auth login
# Successfully authenticated as work@example.com!
# Active account set to: work@example.com
```

Authenticate multiple accounts (e.g. work and personal):
```bash
gws auth login
# Log in with your second account (e.g. user@example.com)
```

List all authenticated accounts:
```bash
gws auth list
# Authenticated Accounts:
# * user@example.com (valid)
#   work@example.com (valid)
```

Switch the default active account:
```bash
gws auth switch work@example.com
```

Override account on any command using `-a` / `--account`:
```bash
gws mail list -a user@example.com
gws cal list --account work@example.com
```

Log out of a specific account (or all):
```bash
gws auth logout user@example.com
gws auth logout --all
```

Check authentication status anytime:
```bash
gws auth status
```

---

## 💻 CLI Usage

### Gmail
```bash
# List recent emails (plain text or JSON)
gws mail list --max 5
gws mail list --max 5 --json | jq .

# Search specific emails
gws mail list --query "from:recruiter is:unread"

# Send email with attachment and auto-generated Google Meet link
gws mail send --to "client@example.com" \
  --subject "Project Kickoff" \
  --body "Looking forward to speaking." \
  --meet \
  --attach ./contract.pdf

# Schedule send
gws mail send --to "client@example.com" --subject "Update" --body "Hello" --delay 10m

# Reply to an existing thread
gws mail reply <message_id> --body "Thanks, let's meet tomorrow." --meet

# Read full email message content and inspect attachments
gws mail read <message_id>

# Download email attachments (aliases: att, attachment)
gws mail download <message_id>                     # Download single attachment or all if multiple
gws mail download <message_id> --dir ./downloads   # Save to destination directory
gws mail download <message_id> contract.pdf        # Download specific file by name
gws mail download <message_id> <att_id> -o out.pdf # Download by ID to specific output path

# Manage drafts
gws mail draft --to "lead@company.com" --subject "Proposal" --body "Draft proposal content."
gws mail drafts
gws mail send-draft <draft_id>
gws mail delete-draft <draft_id>
```

### Google Meet & Calendar
```bash
# Instant Google Meet generation
gws meet create "Team Sync" --start "2026-10-06T15:00:00+07:00" --end "2026-10-06T15:30:00+07:00"

# Send Meet invitation via email
gws meet send "1-on-1 Catchup" --to "partner@example.com"

# View upcoming events (shows Google Meet links and RSVP status)
gws cal list --max 10 --json

# Natural language event creation
gws cal add "Lunch with Alex tomorrow at 1pm"

# Structured event with Google Meet and attendee invitations
gws cal create \
  --title "Client Architecture Review" \
  --start "2026-10-08T10:00:00+07:00" \
  --end "2026-10-08T11:00:00+07:00" \
  --meet \
  --attendees "lead@client.com,dev@client.com"

# RSVP to event invitations
gws cal accept <event_id>
gws cal decline <event_id>
gws cal maybe <event_id>

# Delete event
gws cal delete <event_id>
```

### Google Tasks
```bash
# List and manage task lists
gws tasks lists
gws tasks create-list "Q4 Roadmap"
gws tasks delete-list <list_id>

# List tasks (shows indented subtasks tree)
gws tasks list --list <list_id> --json

# Add task with notes, due date, and attachment link
gws tasks add "Review Sprint Backlog" \
  --notes "Prioritize auth features" \
  --due "2026-10-10T00:00:00.000Z" \
  --link "https://docs.google.com/spreadsheets/d/..."

# Add nested subtask
gws tasks add "Verify OAuth refresh token" --parent <parent_task_id>

# Mark task as completed
gws tasks done <task_id>
```

### Google Sheets & Docs
```bash
# Create spreadsheet & add sheets
gws sheets create "Project Budget"
gws sheets add-sheet <spreadsheet_id> "Expenses"

# Append row & update cells
gws sheets append <spreadsheet_id> "Expenses!A1" "2026-10-06" "Server Hosting" "45 EUR"
gws sheets update <spreadsheet_id> "Expenses!C1" "50 EUR"

# Read cell range
gws sheets read <spreadsheet_id> "Expenses!A1:D10" --json

# Create & read Google Docs
gws docs create "Meeting Notes"
gws docs read <doc_id>
gws docs append <doc_id> "Key decisions made during sprint kickoff.\n"
```

### Google Chat
```bash
# List joined spaces and direct messages (aliases: rooms, dms, list)
gws chat spaces
gws chat spaces --json

# Search conversations / spaces by name or participant
gws chat spaces "Jan"
gws chat spaces -q "dev"

# Search messages across all conversations (alias: find)
gws chat search "WordPress"
gws chat search "roadmap" --max 10 --json

# List recent messages from a space (aliases: msgs, list <space_id>)
gws chat messages "spaces/AAAA..." --max 10
gws chat messages "spaces/AAAA..." --json

# Send message (with optional file attachment)
gws chat send "spaces/AAAA..." "Hello team! Spec document attached." --attach ./spec.pdf

# Reply to a specific message in a thread
gws chat reply "spaces/AAAA..." <message_id> "Acknowledged, on it!"

# React with an emoji to a message
gws chat react <message_name_or_id> "👍"
gws chat reactions <message_name_or_id>
```

### Google Drive
```bash
# List / search Drive files
gws drive list --query "report" --json

# Read & export Google Docs / Sheets / text files
gws drive read <file_id>

# Download & upload files
gws drive download <file_id> ./downloaded_report.pdf
gws drive upload ./document.pdf --name "Final_Spec.pdf"

# Delete file
gws drive delete <file_id>
```

---

## 🤖 Model Context Protocol (MCP) Server

### Local Mode (Stdio)

Run `gws` as a high-performance stdio MCP server for Claude Desktop, Cursor, Zed, or Antigravity:

```bash
gws mcp server
# or legacy alias:
gws serve
```

#### Editor Configuration (Claude / Antigravity / Cursor)
```json
{
  "mcpServers": {
    "gws": {
      "command": "gws",
      "args": ["mcp", "server"]
    }
  }
}
```

### Remote / VPS Mode (HTTP + SSE with Bearer Token)

To run `gws` on a remote VPS and access it securely from your local editor:

1. **Generate a secure 256-bit MCP secret token on the VPS**:
   ```bash
   gws mcp token
   ```
   *(Stored in `~/.config/gws/mcp_token` with `0600` permissions).*

2. **Start the SSE server**:
   ```bash
   gws mcp server --port 8080 --host 0.0.0.0
   ```
   *Requests are automatically authenticated via constant-time Bearer token comparison.*

3. **Configure your local editor**:
   ```json
   {
     "mcpServers": {
       "gws-remote": {
         "serverUrl": "https://vps.example.com/sse",
         "headers": {
           "Authorization": "Bearer gws_mcp_your_secret_token_here"
         }
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
