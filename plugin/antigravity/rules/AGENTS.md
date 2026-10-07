# Google Workspace Rules & Agent Safety Guidelines

These rules govern all interactions with Google Workspace (Gmail, Calendar, Drive, Docs, Sheets, Meet, Tasks) via the `gws` tool and MCP server.

## 1. Explicit Confirmation for Destructive & External Actions
- **Sending Emails**: NEVER send an email (`gmail_send_message` or `gws mail send`) without showing the user the exact recipient, subject, and body, and obtaining explicit confirmation. Prefer creating a draft (`gmail_create_draft`) first.
- **Deleting Resources**: NEVER delete files (`drive_delete_file`), calendar events (`calendar_delete_event`), or tasks (`tasks_delete`) without explicit user consent.
- **Calendar Invites**: Avoid altering attendees or sending calendar updates to external guests without confirmation.

## 2. Data Integrity & Mutating Operations
- **Sheets & Docs**: When writing or appending data to spreadsheets and documents, verify spreadsheet IDs and cell ranges. Avoid overwriting existing headers or formula ranges.
- **Fail-Closed on Auth Errors**: If authentication expires (`token.json`), do not attempt repeated retry loops. Instruct the user to run `gws auth login`.
