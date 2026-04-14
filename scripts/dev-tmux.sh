#!/usr/bin/env sh

set -eu

SESSION_NAME="${TMUX_SESSION_NAME:-docvault-dev}"
ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
ATTACH=true
LOGIN_SHELL="${SHELL:-/bin/sh}"

if [ "${1:-}" = "--no-attach" ]; then
    ATTACH=false
fi

if ! command -v tmux >/dev/null 2>&1; then
    printf '%s\n' "tmux is required but not installed." >&2
    exit 1
fi

attach_session() {
    if [ "$ATTACH" = "false" ]; then
        exit 0
    fi

    if [ -n "${TMUX:-}" ]; then
        exec tmux switch-client -t "$SESSION_NAME"
    fi

    exec tmux attach-session -t "$SESSION_NAME"
}

if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
    attach_session
fi

tmux new-session -d -s "$SESSION_NAME" -c "$ROOT_DIR" -n infra

infra_main_pane=$(tmux display-message -p -t "$SESSION_NAME:infra.0" "#{pane_id}")
infra_helper_pane=$(tmux split-window -P -F "#{pane_id}" -t "$infra_main_pane" -v -c "$ROOT_DIR")
tmux send-keys -t "$infra_main_pane" "just dev-up && just dev-logs" C-m
tmux send-keys -t "$infra_helper_pane" "printf 'Infra helpers\n\n'; printf 'just dev-ps\n'; printf 'just dev-down\n'; printf 'just dev-restart\n'; exec '$LOGIN_SHELL'" C-m
tmux select-layout -t "$SESSION_NAME:infra" main-horizontal

tmux new-window -t "$SESSION_NAME" -n services -c "$ROOT_DIR"
backend_pane=$(tmux display-message -p -t "$SESSION_NAME:services.0" "#{pane_id}")
reminder_pane=$(tmux split-window -P -F "#{pane_id}" -t "$backend_pane" -h -c "$ROOT_DIR")
ocr_pane=$(tmux split-window -P -F "#{pane_id}" -t "$backend_pane" -v -c "$ROOT_DIR")
processing_pane=$(tmux split-window -P -F "#{pane_id}" -t "$reminder_pane" -v -c "$ROOT_DIR")
tmux send-keys -t "$backend_pane" "just dev-backend" C-m
tmux send-keys -t "$reminder_pane" "just dev-reminder" C-m
tmux send-keys -t "$ocr_pane" "just dev-ocr" C-m
tmux send-keys -t "$processing_pane" "just dev-processing" C-m
tmux select-layout -t "$SESSION_NAME:services" tiled

tmux new-window -t "$SESSION_NAME" -n clients -c "$ROOT_DIR"
web_pane=$(tmux display-message -p -t "$SESSION_NAME:clients.0" "#{pane_id}")
mobile_pane=$(tmux split-window -P -F "#{pane_id}" -t "$web_pane" -h -c "$ROOT_DIR")
tmux send-keys -t "$web_pane" "just dev-web" C-m
tmux send-keys -t "$mobile_pane" "just dev-mobile" C-m
tmux select-layout -t "$SESSION_NAME:clients" even-horizontal

tmux new-window -t "$SESSION_NAME" -n shell -c "$ROOT_DIR"
tmux send-keys -t "$SESSION_NAME:shell" "printf 'DocVault dev shell\n'; printf 'Session: %s\n' '$SESSION_NAME'; exec '$LOGIN_SHELL'" C-m

tmux set-option -t "$SESSION_NAME" mouse on
tmux select-window -t "$SESSION_NAME:services"
tmux select-pane -t "$backend_pane"

attach_session
