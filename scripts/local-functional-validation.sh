#!/usr/bin/env bash
# Current-source Linux functional regression; no release packaging or publishing.
set -Eeuo pipefail

if [[ $(uname -s) != Linux ]]; then
  printf '%s\n' 'This runner requires Linux /proc and the local Docker Engine.' >&2
  exit 2
fi
for command in go npm docker python3; do command -v "$command" >/dev/null; done
task_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$task_root"
mkdir -p .local
task_run=$(mktemp -d "$task_root/.local/functional-$(date -u +%Y%m%dT%H%M%S)-XXXXXX")
task_container="blora-functional-$(basename "$task_run")"
task_fixture_pid=''
task_heartbeat_pid=''
task_cleanup() {
  local result=$?
  trap - EXIT INT TERM
  if [[ -n $task_heartbeat_pid ]]; then kill "$task_heartbeat_pid" 2>/dev/null || true; wait "$task_heartbeat_pid" 2>/dev/null || true; fi
  # Exact names/PIDs created by this invocation only; no global container cleanup.
  if docker inspect "$task_container" >/dev/null 2>&1; then docker stop --time 10 "$task_container" >/dev/null 2>&1 || true; fi
  if [[ -n $task_fixture_pid ]]; then
    kill -TERM "$task_fixture_pid" 2>/dev/null || true
    local cleanup_result=0
    wait "$task_fixture_pid" || cleanup_result=$?
    printf '[cleanup] fixture exit=%s\n' "$cleanup_result"
    if [[ $result == 0 && $cleanup_result != 0 ]]; then result=$cleanup_result; fi
  fi
  printf 'Private run evidence: %s\n' "$task_run"
  printf '%s\n' 'Do not publish credentials, state, raw logs or browser snapshots.'
  exit "$result"
}
trap task_cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

printf '%s\n' '[build] locked dependencies, SDK/reference packages, frontend, Master, Daemon and fixture'
npm --prefix web ci >"$task_run/build.log" 2>&1
make sdk build >>"$task_run/build.log" 2>&1
go build -trimpath -o dist/blora-devfixture ./cmd/devfixture >>"$task_run/build.log" 2>&1
(cd web && npm run build) >>"$task_run/build.log" 2>&1
docker info --format '{{.ServerVersion}}' >"$task_run/docker-version.txt"
printf '%s\n' '[fixture] isolated real HTTPS Master, two Daemons and owned Docker tests'
# SIGTERM is handled by the fixture for confirmed cleanup and also terminates
# early initialization before its signal context has been registered.
./dist/blora-devfixture --performance --listen 127.0.0.1:9443 --state-parent "$task_run" --docker-endpoint unix:///var/run/docker.sock >"$task_run/fixture.log" 2>&1 &
task_fixture_pid=$!
task_credentials=''
for ((attempt=0; attempt<120; attempt++)); do
  task_credentials=$(python3 - "$task_run" <<'PY'
from pathlib import Path
import sys
files=list(Path(sys.argv[1]).glob('fixture-*/browser-credentials.json'))
if len(files)==1: print(files[0])
PY
)
  if [[ -n $task_credentials ]]; then break; fi
  if ! kill -0 "$task_fixture_pid" 2>/dev/null; then printf '%s\n' 'Fixture exited before readiness.' >&2; exit 1; fi
  sleep 0.5
done
if [[ -z $task_credentials ]]; then printf '%s\n' 'Fixture readiness timed out.' >&2; exit 1; fi

printf '%s\n' '[regression] all real functional scenarios, one worker, real Docker and 121-point monitoring'
printf '%s\n' '[scope] E08 is separately measured and closed under the user-approved performance scope.'
(while sleep 30; do printf '%s\n' '[regression] still running; individual results are in the private run log'; done) &
task_heartbeat_pid=$!
task_result=0
# The host PID namespace and identical absolute mount are necessary for tests
# that identify and pause only their own Daemon or restart their own fixture.
docker run --name "$task_container" --rm --network host --ipc host --pid host --user 0 \
  -v "$task_root:$task_root" -w "$task_root/web" \
  -e "BLORA_E2E_CREDENTIALS=$task_credentials" -e BLORA_BROWSER=chromium \
  -e BLORA_CHROMIUM=/ms-playwright/chromium-1243/chrome-linux64/chrome \
  -e BLORA_E01_HISTORY_SOAK=1 -e BLORA_E01_SAMPLE_INTERVAL_MS=1000 \
  -e PLAYWRIGHT_NO_COPY_PROMPT=1 -e "PLAYWRIGHT_JSON_OUTPUT_FILE=$task_run/real-functional.json" \
  mcr.microsoft.com/playwright:v1.63.0-noble \
  npx playwright test --config=playwright.real.config.ts \
  --grep-invert 'real eight-window mixed load' --workers=1 --reporter=list,json \
  "--output=$task_run/browser-output" "$@" >"$task_run/real-functional.log" 2>&1 || task_result=$?
python3 - "$task_run/real-functional.json" <<'PY'
import json,sys
from pathlib import Path
p=Path(sys.argv[1])
if not p.exists():
    print('[result] no structured browser report was produced')
    sys.exit(1)
r=json.loads(p.read_text());s=r['stats']
print('[result] '+json.dumps({k:s[k] for k in ['expected','unexpected','skipped','flaky','duration']}))
if s['unexpected'] or s['skipped'] or s['flaky'] or s['expected']==0: sys.exit(1)
PY
exit "$task_result"
