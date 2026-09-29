#!/bin/sh
set -eu
# Invoke inside a disposable container with xvfb-run and dbus-run-session.
test -f /.dockerenv
test "${BLORA_NATIVE_NOTIFICATION_E2E:-}" = 1
test -n "${DBUS_SESSION_BUS_ADDRESS:-}"
test -n "${DISPLAY:-}"
dunst -conf /workspace/scripts/native-notifications.dunstrc &
blora_notification_pid=$!
trap 'kill "$blora_notification_pid" 2>/dev/null || true; wait "$blora_notification_pid" 2>/dev/null || true' EXIT
cd /workspace/web
node node_modules/@playwright/test/cli.js test --config playwright.notifications.config.ts
