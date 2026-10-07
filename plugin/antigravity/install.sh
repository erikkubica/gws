#!/usr/bin/env bash
set -euo pipefail

# gws — Antigravity Plugin Installer

PLUGIN_SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ANTIGRAVITY_CONFIG_DIR="${HOME}/.gemini/config"
PLUGIN_DIR="${ANTIGRAVITY_CONFIG_DIR}/plugins"
PLUGIN_NAME="google-workspace-plugin"
BIN_PATH=""
UNINSTALL=0

log_info() { echo "==> $1"; }
log_success() { echo "  ✔ $1"; }
log_warn() { echo "  ⚠ $1"; }
log_error() { echo "  ✖ $1" >&2; }

show_usage() {
  cat <<'EOF'
gws Antigravity Plugin Installer

Usage:
  ./plugin/antigravity/install.sh [options]

Options:
  --bin-path <path>    Path to gws executable (default: search PATH or ~/.local/bin/gws)
  --plugin-dir <dir>   Antigravity plugins directory (default: ~/.gemini/config/plugins)
  --uninstall          Uninstall the Antigravity plugin
  -h, --help           Show this help message
EOF
  exit 0
}

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --bin-path) BIN_PATH="$2"; shift 2 ;;
      --plugin-dir) PLUGIN_DIR="$2"; shift 2 ;;
      --uninstall) UNINSTALL=1; shift ;;
      -h|--help) show_usage ;;
      *) log_error "Unknown argument: $1"; show_usage ;;
    esac
  done
}

resolve_bin_path() {
  if [ -n "$BIN_PATH" ] && [ -x "$BIN_PATH" ]; then
    return 0
  fi
  if command -v gws >/dev/null 2>&1; then
    BIN_PATH="$(command -v gws)"
  elif [ -x "$HOME/.local/bin/gws" ]; then
    BIN_PATH="$HOME/.local/bin/gws"
  elif command -v gmcp >/dev/null 2>&1; then
    BIN_PATH="$(command -v gmcp)"
  else
    log_error "gws binary not found. Please install gws first or pass --bin-path."
    exit 1
  fi
}

install_plugin_files() {
  log_info "Installing Antigravity plugin to $TARGET_DIR..."
  mkdir -p "$TARGET_DIR/rules" "$TARGET_DIR/skills/gws"
  cp "$PLUGIN_SRC/plugin.json" "$TARGET_DIR/plugin.json"
  cp "$PLUGIN_SRC/rules/AGENTS.md" "$TARGET_DIR/rules/AGENTS.md"
  cp "$PLUGIN_SRC/skills/gws/SKILL.md" "$TARGET_DIR/skills/gws/SKILL.md"
  rm -rf "$TARGET_DIR/skills/gmcp"
  cat > "$TARGET_DIR/mcp_config.json" <<EOF
{
  "mcpServers": {
    "gws": {
      "command": "$BIN_PATH",
      "args": ["mcp", "server"]
    }
  }
}
EOF
  log_success "Plugin files copied and configured with binary: $BIN_PATH"
}

clean_legacy_configs() {
  local cfg="$ANTIGRAVITY_CONFIG_DIR/mcp_config.json"
  if [ -f "$cfg" ]; then
    if grep -qE '("gmcp"|"gws")' "$cfg" 2>/dev/null; then
      cp "$cfg" "$cfg.bak"
      if command -v jq >/dev/null 2>&1; then
        jq 'del(.mcpServers.gmcp, .mcpServers.gws)' "$cfg.bak" > "$cfg"
        log_success "Cleaned legacy standalone gmcp/gws from mcp_config.json"
      fi
    fi
  fi
  for p in "$ANTIGRAVITY_CONFIG_DIR/skills/gmcp" "$HOME/.agents/skills/gmcp" "$ANTIGRAVITY_CONFIG_DIR/skills/gws" "$HOME/.agents/skills/gws"; do
    if [ -d "$p" ]; then
      rm -rf "$p"
      log_success "Removed legacy standalone skill: $p"
    fi
  done
}

enable_plugin_json() {
  local cfg="$ANTIGRAVITY_CONFIG_DIR/config.json"
  [ ! -f "$cfg" ] && echo '{}' > "$cfg"
  if command -v jq >/dev/null 2>&1; then
    local tmp; tmp="$(mktemp)"
    jq '.plugins["google-workspace-plugin"] = {"enabled": true}' "$cfg" > "$tmp" && mv "$tmp" "$cfg"
    log_success "Enabled google-workspace-plugin in $cfg"
  fi
}

enable_plugin_in_agy() {
  if command -v agy >/dev/null 2>&1; then
    log_info "Validating and enabling plugin via Antigravity CLI..."
    agy plugin validate "$TARGET_DIR" || true
    agy plugin enable "$PLUGIN_NAME" || true
    log_success "Plugin enabled in Antigravity"
  else
    enable_plugin_json
  fi
}

perform_uninstall() {
  log_info "Uninstalling Antigravity plugin ($PLUGIN_NAME)..."
  rm -rf "$TARGET_DIR" && log_success "Removed plugin directory: $TARGET_DIR"
  if command -v agy >/dev/null 2>&1; then
    agy plugin disable "$PLUGIN_NAME" 2>/dev/null || true
  fi
  local cfg="$ANTIGRAVITY_CONFIG_DIR/config.json"
  if [ -f "$cfg" ] && command -v jq >/dev/null 2>&1; then
    local tmp; tmp="$(mktemp)"
    jq 'del(.plugins["google-workspace-plugin"])' "$cfg" > "$tmp" && mv "$tmp" "$cfg"
  fi
  log_success "Antigravity plugin uninstalled."
  exit 0
}

main() {
  parse_args "$@"
  TARGET_DIR="$PLUGIN_DIR/$PLUGIN_NAME"
  if [ "$UNINSTALL" -eq 1 ]; then
    perform_uninstall
  fi
  resolve_bin_path
  install_plugin_files
  clean_legacy_configs
  enable_plugin_in_agy
  log_success "Antigravity plugin installation complete!"
}

main "$@"
