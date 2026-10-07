#!/usr/bin/env bash
set -euo pipefail

# gws — Universal Installer and Multi-Editor Plugin Orchestrator

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${HOME}/.local/bin"
TARGET_EDITOR="all"
SKIP_BUILD=0
SKIP_PLUGIN=0
UNINSTALL=0

log_info() { echo "==> $1"; }
log_success() { echo "  ✔ $1"; }
log_warn() { echo "  ⚠ $1"; }
log_error() { echo "  ✖ $1" >&2; }

show_usage() {
  cat <<'EOF'
gws Universal Installer

Usage:
  ./install.sh [options]

Options:
  --bin-dir <dir>      Installation directory for gws binary (default: ~/.local/bin)
  --editor <name>      Target editor plugin to install (e.g., antigravity, or "all"; default: all)
  --skip-build         Skip compiling the Go binary
  --skip-plugin        Skip installing editor plugins
  --uninstall          Uninstall the binary and all editor plugins
  -h, --help           Show this help message

Available editor plugins in plugin/:
EOF
  for d in "$REPO_ROOT"/plugin/*/; do
    if [ -f "${d}install.sh" ]; then
      echo "  - $(basename "$d")"
    fi
  done
  exit 0
}

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --bin-dir) BIN_DIR="$2"; shift 2 ;;
      --editor) TARGET_EDITOR="$2"; shift 2 ;;
      --skip-build) SKIP_BUILD=1; shift ;;
      --skip-plugin) SKIP_PLUGIN=1; shift ;;
      --uninstall) UNINSTALL=1; shift ;;
      -h|--help) show_usage ;;
      *) log_error "Unknown argument: $1"; show_usage ;;
    esac
  done
}

check_prerequisites() {
  if [ "$SKIP_BUILD" -eq 0 ] && ! command -v go >/dev/null 2>&1; then
    if [ -f "$REPO_ROOT/bin/gws" ]; then
      log_warn "Go compiler not found, using prebuilt bin/gws binary."
    else
      log_error "Go compiler not found in PATH. Please install Go or pass --skip-build."
      exit 1
    fi
  fi
}

migrate_config_dir() {
  local old_dir="$HOME/.config/gmcp"
  local new_dir="$HOME/.config/gws"
  if [ -d "$old_dir" ]; then
    mkdir -p "$new_dir"
    for f in credentials.json token.json; do
      if [ -f "$old_dir/$f" ] && [ ! -f "$new_dir/$f" ]; then
        cp "$old_dir/$f" "$new_dir/$f"
        chmod 600 "$new_dir/$f" 2>/dev/null || true
        log_success "Migrated $f from ~/.config/gmcp to ~/.config/gws"
      fi
    done
  fi
}

build_binary() {
  [ "$SKIP_BUILD" -eq 1 ] && return 0
  log_info "Building gws binary..."
  mkdir -p "$BIN_DIR"
  if command -v go >/dev/null 2>&1; then
    go build -o "$BIN_DIR/gws" "$REPO_ROOT/cmd/gws"
  elif [ -f "$REPO_ROOT/bin/gws" ]; then
    cp "$REPO_ROOT/bin/gws" "$BIN_DIR/gws"
  fi
  chmod +x "$BIN_DIR/gws"
  log_success "Binary installed at: $BIN_DIR/gws"
  if [ -f "$BIN_DIR/gmcp" ]; then
    rm -f "$BIN_DIR/gmcp"
    log_success "Cleaned up old gmcp binary at: $BIN_DIR/gmcp"
  fi
}

verify_binary() {
  if [ ! -x "$BIN_DIR/gws" ]; then
    log_error "gws binary not found or not executable at $BIN_DIR/gws"
    exit 1
  fi
  "$BIN_DIR/gws" --help >/dev/null 2>&1 || true
  case ":$PATH:" in
    *":$BIN_DIR:"*) ;;
    *) log_warn "$BIN_DIR is not in your PATH. Consider adding it to your shell profile." ;;
  esac
}

install_editor_plugin() {
  local editor_dir="$1"
  local script="${editor_dir}/install.sh"
  local editor_name
  editor_name="$(basename "$editor_dir")"
  if [ -f "$script" ] && [ -x "$script" ]; then
    log_info "Running installer for editor plugin: $editor_name..."
    "$script" --bin-path "$BIN_DIR/gws"
  else
    log_warn "No executable install.sh found in $editor_dir"
  fi
}

install_plugins() {
  [ "$SKIP_PLUGIN" -eq 1 ] && return 0
  if [ "$TARGET_EDITOR" != "all" ]; then
    local specific="$REPO_ROOT/plugin/$TARGET_EDITOR"
    if [ ! -d "$specific" ]; then
      log_error "Editor plugin directory not found: $specific"
      exit 1
    fi
    install_editor_plugin "$specific"
    return 0
  fi
  for d in "$REPO_ROOT"/plugin/*/; do
    if [ -d "$d" ]; then
      install_editor_plugin "$d"
    fi
  done
}

uninstall_plugins() {
  for d in "$REPO_ROOT"/plugin/*/; do
    local script="${d}install.sh"
    if [ -x "$script" ]; then
      "$script" --uninstall || true
    fi
  done
}

perform_uninstall() {
  log_info "Uninstalling gws and all plugins..."
  uninstall_plugins
  for b in "$BIN_DIR/gws" "$BIN_DIR/gmcp"; do
    if [ -f "$b" ]; then
      rm -f "$b" && log_success "Removed binary: $b"
    fi
  done
  log_success "Uninstall completed."
  exit 0
}

check_auth_status() {
  local tok="$HOME/.config/gws/token.json"
  local cred="$HOME/.config/gws/credentials.json"
  echo ""
  log_info "Checking Google OAuth authentication status..."
  if [ -f "$tok" ]; then
    log_success "OAuth token found at $tok"
  elif [ -f "$cred" ]; then
    log_warn "Credentials found at $cred, but authorization token is missing."
    echo "       Run 'gws auth login' to authenticate via your browser."
  else
    log_warn "No Google OAuth credentials found."
    echo "       1. Create OAuth 2.0 Client ID in Google Cloud Console"
    echo "       2. Save credentials JSON to: ~/.config/gws/credentials.json"
    echo "       3. Run: gws auth login"
  fi
}

main() {
  parse_args "$@"
  if [ "$UNINSTALL" -eq 1 ]; then
    perform_uninstall
  fi
  check_prerequisites
  migrate_config_dir
  build_binary
  verify_binary
  install_plugins
  check_auth_status
  echo ""
  log_success "All requested installations complete!"
}

main "$@"
