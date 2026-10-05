# Local remaining work — 2026-10-05

The user requested completion of locally feasible remaining work and explicitly
excluded release work. This report supplements the
[scope review](scope-review-2026-10-05.md); it does not change historical failed
results or certify a Windows run from Linux evidence.

## Firewall capability gate

`SystemApp.vue` now enables rule preview/application only when the selected
node reports both `platform: linux` and `firewall: firewalld`. Status querying
remains independent. Windows `netsh`, macOS `pf`, and other Linux readers must
not offer an operation their backend does not implement.

Local verification:

- `cd web && npm run check`: passed.
- `cd web && npx playwright test tests/browser/system-management.spec.ts --workers=1 --reporter=line`:
  **5/5 passed**, approximately 1.1 minutes, Linux Chromium.
- Three new cases verify readable status, disabled mutation controls, persistence
  after refresh, and zero POST requests for `netsh`, `pf`, and `nftables`.
- Existing Linux firewalld preview/application/task confirmation and unavailable
  capability scenarios also passed. These use browser API fixtures and are
  frontend gate regression evidence, not Windows firewall execution evidence.

The first browser launch was blocked by the local execution sandbox's loopback
listener restriction before any test ran. The authorized local test launch
completed successfully; it did not modify host firewall rules or services.

Release packaging, release rollback, and publishing are deferred by the user's
latest instruction. Ordinary test builds and stopped-state restart checks remain
within the current local validation scope.

## Frozen-material performance investigation

A new layout/style isolation candidate was measured in an eight-window
diagnostic. Captured screenshots were pixel-identical, but candidate p95 was
952.5/922.7 ms against a final baseline of 1005.3 ms, without the full E08
PTY/transfer workload. It was not adopted: these measurements do not establish
a robust production gain and remain far above 50 ms. See the
[diagnostic and its limits](render-layout-diagnostic-2026-10-05.md).

The original complete-load performance failures remain failures. This local
investigation neither weakens the requirement nor asks the user to supply
another special environment.

## Ordinary-device real backend entry point

Added `cmd/device-validation`, `web/scripts/device-browser.mjs`, and
`scripts/windows-device-validation.ps1`. The checker starts actual built Master
and Daemon executables in fresh private state with a new local TLS certificate,
then uses their real HTTPS API and terminal WSS. The browser logs in to that
same server and uses the real Vue/Monaco interface; no API doubles are used.

The checker verifies initialization/login/enrollment, a disposable stopped
instance, member permission denial, UTF-8 file write/read, backup and restore,
fresh monitoring, extension install/upgrade/remove, native instance start/stop,
terminal execution with an actual file-write result, cursor-based same-session
reconnection, editor draft restoration on refresh and server save, and stopped
resource process restart preserving identities and file content. Cleanup is an
explicit required check. Failed or unconfirmed cleanup prevents success and
retains private recovery state. The disposable worker has a nine-minute maximum
lifetime as a last safety bound; tests still require immediate confirmed stop.

The final real-backend run on Linux/amd64 passed **14/14**, exit 0, from
`2026-10-05T14:43:54.192920188Z` to `2026-10-05T14:44:05.580580274Z`,
approximately 11.388 seconds. It also confirmed the same persisted SUCCEEDED
file task receipt after restarting Master/Daemon.
The Chromium UI stage used the matching Playwright 1.63.0 container and passed
login, resource/file creation, unsaved Chinese draft refresh, server readback,
and a second refresh. This run is **Linux evidence, not Windows execution**.
The private sanitized result was `/tmp/blora-device-local-final.json`; credentials,
database/state, raw process logs and browser snapshots were not copied here.

The latest production frontend build (including the firewall gate) passed,
3,740 modules, Vite build 7.64 seconds. `go test ./cmd/device-validation` passed
two checker safety tests; `go vet ./cmd/device-validation` and
`env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/blora-device-validation.exe ./cmd/device-validation`
exited 0. `node --check web/scripts/device-browser.mjs` passed. These establish local code/entry-point
readiness, not execution of the Windows binary. The checker safety tests verify
that an interrupted task cannot pass and that cleanup of an unconfirmed live
resource fails while preserving private state, even if the original workflow
context has already been cancelled.

The new PowerShell wrapper requires a full fixed commit, clones into a fresh
directory, compares its SHA-256 with the checked-out script, builds normal test
executables, and installs Chromium only. It requires exactly one PASS for every
named device check, including the actual browser and cleanup. Non-Windows,
missing, duplicate and failed cleanup evidence is rejected. Console progress and
15-second heartbeats use plain text. .NET ZIP creation avoids the host progress
renderer failure observed in the original Windows script.

`scripts/test-windows-device-validation.ps1` passed in the existing isolated
PowerShell 7.4 Linux container: literal argument preservation, failed child
retention, required-check rejection, report source identity, exclusion of private
files, and full ZIP checksum coverage. Initial harness-only invocation/variable
scope mistakes were corrected before this pass; they were not product failures
or Windows execution evidence. Runtime failures still produce the unique ZIP.

One ordinary Windows run of this new entry point remains necessary to obtain
Windows real-backend evidence. It does not rerun the completed WebKit followup
or the old broad native/race suite, and does not require administrative service,
firewall, VM, disk-exhaustion or release testing. The command and scope are in
[Windows validation](../../operations/WINDOWS_VALIDATION.md).

All owned fixture, terminal, worker, browser and temporary test-container
sessions from the final run were closed. No release packages were generated.
