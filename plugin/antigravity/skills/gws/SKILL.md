---
name: gws
description: All-in-one Google Workspace tool (Gmail, Calendar, Meet, Drive, Tasks, Sheets, Docs, YouTube). Use whenever the user asks to read, search, send, or manage emails, schedule calendar events, create Google Meet links, accept/decline invites, access Drive files, create/edit Google Sheets, read Docs, manage todos/tasks, or search YouTube.
---

# `gws` — Google Workspace Powerhouse

`gws` is an all-in-one CLI and MCP tool with persistent OAuth installed at `~/.local/bin/gws` (or in your `$PATH`).
It provides instant terminal access to Gmail, Calendar, Meet, Drive, Tasks, Sheets, Docs, and YouTube without needing any external tools (replaces `gmcli`, `gccli`, `gdcli`).

Authentication is configured in `~/.config/gws/token.json`.

## Quick Reference

### 📬 Gmail (`gws mail`)
- Search/list messages: `gws mail list --query "<query>" --max 10 [--json]`
- Read message: `gws mail read <message_id> [--json]`
- Send email: `gws mail send --to <email> --subject "<subj>" --body "<body>" [--attach <path>] [--meet]`
- Reply to thread: `gws mail reply <message_id> --body "<body>" [--attach <path>] [--meet]`
- Schedule send: `gws mail send ... --delay 10m` or `--at "2026-10-06T09:00:00+07:00"`
- Manage drafts: `gws mail draft ...`, `gws mail drafts`, `gws mail send-draft <draft_id>`, `gws mail delete-draft <draft_id>`

### 📹 Google Meet (`gws meet`)
- Quick Meet generation: `gws meet create "<title>" [--start <iso>] [--end <iso>] [--attendees "a@b.com,c@d.com"]`
- Send Meet invitation email: `gws meet send "<title>" --to <email> [--start <iso>] [--end <iso>]`

### 📅 Calendar (`gws cal`)
- List upcoming events: `gws cal list [--max 10] [--json]`
- Quick-add event (natural language): `gws cal add "Meeting tomorrow at 3pm"`
- Structured event creation: `gws cal create --title "<title>" --start "<rfc3339>" --end "<rfc3339>" [--meet] [--attendees "a@b.com,c@d.com"]`
- RSVP response:
  - Accept: `gws cal accept <event_id>`
  - Decline: `gws cal decline <event_id>`
  - Maybe / tentative: `gws cal maybe <event_id>`
  - Custom response: `gws cal respond <event_id> <accepted|declined|tentative>`
- Delete event: `gws cal delete <event_id>`

### 📁 Google Drive (`gws drive`)
- List & search files: `gws drive list [--query "<search>"] [--max 20] [--json]`
- Read / export file text: `gws drive read <file_id> [--json]`
- Download file: `gws drive download <file_id> <local_destination_path>`
- Upload file: `gws drive upload <local_file_path> [--name <name>]`
- Create file: `gws drive create "<filename>" [--mime "<mimetype>"]`
- Delete file: `gws drive delete <file_id>`

### 📋 Google Tasks (`gws tasks`)
- List task lists: `gws tasks lists [--json]`
- Create task list: `gws tasks create-list "<title>"`
- Delete task list: `gws tasks delete-list <list_id>`
- List tasks: `gws tasks list [--list <list_id>] [--max 20] [--json]`
- Add task: `gws tasks add "<title>" [--notes "<notes>"] [--due "<rfc3339>"] [--link "<url>"] [--list <list_id>]`
- Add subtask: `gws tasks add "<title>" --parent <parent_task_id> [--list <list_id>]`
- Complete task: `gws tasks done <task_id> [--list <list_id>]`
- Delete task: `gws tasks delete <task_id> [--list <list_id>]`

### 📊 Google Sheets (`gws sheets`)
- Create spreadsheet: `gws sheets create "<title>"`
- Add sheet/tab: `gws sheets add-sheet <spreadsheet_id> "<sheet_title>"`
- Append row: `gws sheets append <spreadsheet_id> "<range>" "val1" "val2" ...`
- Read range: `gws sheets read <spreadsheet_id> "<range>" [--json]`
- Update cell(s): `gws sheets update <spreadsheet_id> "<range>" "val1" "val2" ...`

### 📄 Google Docs (`gws docs`)
- Create document: `gws docs create "<title>"`
- Read document: `gws docs read <doc_id> [--json]`
- Append text: `gws docs append <doc_id> "<text>"`

### ▶️ YouTube (`gws yt`)
- Search videos: `gws yt search "<query>" [--max 10] [--json]`
- Video statistics: `gws yt stats <video_id> [--json]`

### ⚙️ GCP Setup & Multi-Account Auth (`gws gcp`, `gws auth`)
- Import GCP credentials JSON: `gws gcp import <path/to/credentials.json>`
- Set GCP credentials directly: `gws gcp set <client_id> <client_secret>`
- View GCP OAuth app status: `gws gcp status`
- Interactive browser login: `gws auth login`
- List all authenticated accounts: `gws auth list`
- Switch active account: `gws auth switch <email>`
- Override account on any command: `gws --account <email> ...` (or `-a <email>`)
- Log out account: `gws auth logout [email] [--all]`
- View auth status: `gws auth status`

### 🤖 MCP Server & Remote Auth (`gws mcp`)
- Run local stdio server: `gws mcp server` (or `gws serve`)
- Run remote SSE server: `gws mcp server --port 8080 [--host 0.0.0.0]`
- Generate secure MCP secret token: `gws mcp token`
- View active MCP token: `gws mcp token --show`
- Revoke active MCP token: `gws mcp token --revoke`

## 👥 Multi-Account Usage in MCP Tools
- **Account Discovery**: Call `workspace_list_accounts` to inspect configured accounts and identify the active one.
- **Switching Active Account**: Call `workspace_switch_account` with `account: "<email>"` to switch the default account.
- **Per-Tool Override**: Every MCP tool (Gmail, Calendar, Drive, Docs, Sheets, Tasks, YouTube, Meet) accepts an optional `account` string argument (e.g. `account: "work@example.com"`). When provided, the tool executes specifically for that account without altering the active default.

