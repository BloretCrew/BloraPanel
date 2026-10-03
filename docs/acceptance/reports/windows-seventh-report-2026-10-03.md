# Seventh returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`aa2456fc3474ad6edc1ba58c0b147c2f80b5d1e5300bcfd2e503c695dbc8e3b5`.
All **14 entries** in `SHA256SUMS.txt` match; no listed file is missing.
The ZIP has 15 entries including the manifest. Tested checkout and requested ref:
**`8ad386a97cf214caa9a5eafa75d7a194e136ed87`**,
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]`.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-03 18:43:33–18:47:00 UTC+8**
(10:43:33–10:47:00 UTC). Original outcome **FAILED**, no fatal runner error.
Private machine paths and raw uploaded logs are not published.

| Actual Windows selection | Passed | Failed | Skipped |
|---|---:|---:|---:|
| Pending task-summary cancellation and restored task filter | 3 | 0 | 0 |
| Continued task-summary polling after prevented navigation | 3 | 0 | 0 |
| Queued task polling during departure | 3 | 0 | 0 |
| Terminal fallback renderer / worker checkpoints | 2 | 1 | 0 |
| Total | **11** | **1** | **0** |

Playwright JSON duration: **181.593s**; browser stage: **183.360s**.
No retries or flaky outcomes. The three task-query corrections have actual
Windows evidence at this revision. Go/native/race, SDK, Chromium and Firefox
were deliberately omitted; empty arrays are not new passing results. Prior
native and Go passes remain scoped to the [fifth report](windows-fifth-report-2026-10-01.md).

## An independent polling path

The remaining terminal execution failed the unchanged empty-page-error assertion
at `terminal.spec.ts:109` after 21.517s. Its fetch diagnostic is for the instance
**log run list**, not the task list or summary. The stack identifies
`InstanceConsole.refreshRuns`, invoked by its independent three-second timer.
Around the second real refresh, the bundled diagnostic records:

| Relative to second `beforeunload` | Event |
|---|---|
| 0ms | `beforeunload` |
| +9ms | Playwright instance-log fetch access-control diagnostic |
| +10ms | `pagehide` |
| +29ms | New document navigation |
| +364ms | New document `pageshow` |

There is no native `window.error` or `unhandledrejection`, and the visible,
focused 1440×960 page passes its paint probe with no open dialog. The old
diagnostic tracks task HTTP timing, so it cannot establish whether the log
request started before or after `beforeunload`. It identifies the failing
polling path, not a harmless browser warning or a resolved Windows failure.

A controlled browser test holds a real pending log-list read and invokes the
actual interval callbacks after a prevented `beforeunload`. The unchanged
production baseline starts a new log fetch in that window and fails:
**0/1 passed, 24.702s**. No fake page error is injected. The corrected test also
replays the callbacks after the pending read has completed, preventing the
component's overlap guard from concealing an unguarded new request.

## Shared safe-read lifecycle

The document gate now lives at the API read boundary, covering independent
console, editor, upload, Compose and terminal-status polling as well as query
observers. GET/HEAD and explicitly read-only Docker queries share this gate;
ordinary POST/PUT/DELETE mutations keep their original behavior and identities.

`beforeunload` pauses new safe reads and aborts existing ones while the document
can still cancel fetches. Only cancellation issued by this document lifecycle
can restart a safe read after a surviving document paints or receives
`pageshow`. Real network failures and owner cancellation still propagate.
`pagehide` invalidates older resume frames and holds reads until `pageshow`.
Task-query cancellation also preserves its last confirmed cache state.

The read signal remains active through response-body parsing, so cancellation
cannot become an empty successful response. Deferred reads check their original
account before fetching. InstanceConsole owns an abort signal for its reads,
cancels it on destruction, prevents overlapping poll requests, and discards
results after destruction. Command submissions, upload chunks, save commits,
terminal input and backend tasks are never automatically replayed by this gate.

The original terminal/page-error, ACK, checkpoint, no-replay and credential
assertions remain unchanged. No settling sleep, force click, error filter,
skipped test, retry or extended deadline was added. Failure-only diagnostics
now include bounded log request timings as well as task timings, without bodies,
field values, query strings or recovery-storage dumps.

The controlled prevented-navigation tests do not establish native confirmation
dialog or bfcache acceptance. Paint-based resumption does not prove every slow
navigation/background-tab sequence. Relevant browser contracts:
[abort includes response-body consumption](https://developer.mozilla.org/en-US/docs/Web/API/AbortController/abort),
[pagehide](https://developer.mozilla.org/en-US/docs/Web/API/Window/pagehide_event),
[animation-frame callbacks](https://developer.mozilla.org/en-US/docs/Web/API/Window/requestAnimationFrame).

## Local validation

Evidence: `.local/evidence/windows-report7-2026-10-03/`. Local checks run on
Linux and do not establish that the Windows diagnostic is resolved.

| Check | Result |
|---|---|
| Controlled log-polling baseline | 0/1 passed, 24.702s; new departure-time fetch reproduced |
| Initial corrected log-polling case, three WebKit executions | 3/3 passed, 39.659s; later strengthened with a second queued-callback boundary |
| Frontend units | 87/87 in 16 files, 1.11s |
| Frontend type check and production build | Passed, 3,740 modules, Vite 5.36s; existing large-chunk warning retained |
| PowerShell 7.4 runner harness | Final pass; first run caught an unsynchronized broad-selection count, corrected to ten cases |
| Final WebKit combination: accounts, three terminal configurations, four navigation cases and two log/input cases, each repeated three times | 36/36 passed, 244.049s |
| Final Firefox, same twelve cases once each | 12/12 passed, 89.141s |
| Final Chromium, same twelve plus Docker, batch, files, remote editor save and system-task cases | 23/23 passed, 142.687s |
| Runner functions with actual browser JSON | Passed; broad ten-title subsets each pass once on Chromium/Firefox and three times on WebKit; all five narrow titles pass three times each in the WebKit combination |
| Runner-generated selection with the actual Playwright CLI `--list` | Exactly five required titles three times each, 15 executions in two files; selection only, not another executed suite |

The runner harness rejects missing titles, a missing third execution, incomplete
aggregate counts and omitted-engine reports being imported as passes. It retains
the broken-progress-host check, stage failure/timeout behavior, compiler digest
and traversal checks, diagnostic privacy boundaries and .NET ZIP recovery.

Commands use the official Playwright `v1.63.0-noble` container, fresh Vite
(`BLORA_E2E_FRESH_SERVER=1`), the selected engine and one worker:

```text
node node_modules/@playwright/test/cli.js test tests/browser/accounts.spec.ts tests/browser/terminal.spec.ts tests/browser/query-navigation.spec.ts tests/browser/logs.spec.ts --workers=1 --reporter=line,json
# WebKit additionally uses --repeat-each=3.
# Chromium additionally includes docker.spec.ts, instance-batch.spec.ts,
# files.spec.ts, editor-save.spec.ts and system-management.spec.ts.
npm run test
npm run build
pwsh -NoProfile -File scripts/test-windows-validation.ps1
```

All final browser runs have zero skips, retries, flaky results or failures.
Vite proxy connection-refused warnings remain visible in transport-double and
teardown paths; these checks are not real-backend or Windows host-lifecycle
evidence. UI/assets, Go/native code and existing release artifacts were not
changed by this correction. The original failed Windows report is not rewritten
as passing evidence.

## Next Windows run

Use the updated runner and source from the same pinned commit supplied in the
conversation:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -Ref <commit> -Followup -BrowsersOnly -WebKitOnly
```

Five cases, each run three times: the remaining terminal configuration, the
three previously confirmed task-navigation cases, and the new log-polling case.
Required outcome: **15/15 actual passing executions**, no skips, retries,
expected failures or missing results. This does not repeat Go/GCC/SDK or the
other browser engines, and does not import previous passes into the new report.
The report still records its omissions and coverage gaps; return the ZIP even
if a test fails. Full Windows/native lifecycle acceptance, strict E08, opt-in
Docker/PTY checks, physical faults and external environments remain open.
