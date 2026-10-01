# Sixth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`3f42d2492cea0b700c1922a4b42ba9d2d0fd976779d1b9044a83d8e99d3b638a`.
All **20 entries** in `SHA256SUMS.txt` match their actual contents; no listed
file is missing. The ZIP contains 21 entries including the manifest. The exact
tested checkout and requested ref are
**`f98b078c63df8f7ba3f5b9ce2cc3a84018c3fb65`**, `runMode=followup-browser`.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-01 18:45:21–18:56:39 Asia/Shanghai**
(10:45:21–10:56:39 UTC). No fatal runner error; original outcome **FAILED**.
Private machine paths and raw report logs are not published.

| Actual Windows browser selection | Passed | Failed | Skipped | JSON duration |
|---|---:|---:|---:|---:|
| Chromium, eight affected scenarios | 8 | 0 | 0 | 61.122s |
| Firefox, eight affected scenarios | 8 | 0 | 0 | 64.570s |
| WebKit, all eight repeated three times | 23 | 1 | 0 | 489.305s |

There were no retries or flaky outcomes. Both query-navigation regressions from
the [fifth report](windows-fifth-report-2026-10-01.md) passed on all engines,
including three executions each on Windows WebKit. All three account scenarios
also passed three times on WebKit. The browser-only selection explicitly omitted
Go/native/race/GCC and SDK checks. Their empty result arrays are omissions, not
new passes. Prior Windows Go/native verification remains scoped to its own
revision and the fifth report.

## The remaining failure and its timing

The first fallback-renderer/worker terminal execution failed the unchanged
empty-page-error assertion at `terminal.spec.ts:84` after 22.218s. The other two
executions of that configuration passed; automatic-renderer/worker and
fallback-renderer/main passed all three each.

The single bundled runtime diagnostic reports a task-list fetch access-control
error on the local Vite origin, from TanStack's query polling path. Relative to
the initial page navigation timestamp:

| Offset | Recorded event |
|---|---|
| +4,153ms / +4,185ms | Task-summary GET starts / finishes |
| +4,474ms / +4,487ms | Task-list GET starts / finishes |
| +6,492ms | `beforeunload` |
| +6,500ms | Playwright task-list fetch access-control diagnostic |
| +6,508ms | `pagehide` |
| +6,521ms | New document navigation |
| +6,930ms | New document `pageshow` |
| +7,107ms / +7,132ms | New document summary GET starts / finishes |
| +7,224ms / +7,255ms | New document list GET starts / finishes |

There is no corresponding new HTTP request event for the failing attempt, no
recorded native `window.error` or `unhandledrejection`, and no open dialog. The
visible, focused page passed the paint probe at 1440×960. These facts identify a
narrow departure window; they do not establish that the diagnostic is harmless
or prove its resolution on Windows.

## Guarding reads started after cancellation

Cancelling existing queries during `beforeunload` did not prevent an already
queued QueryObserver polling callback from starting a new read afterward.
A controlled test retains the real task summary/list interval callbacks,
dispatches a prevented `beforeunload`, and invokes those callbacks before the
next frame. The unchanged production baseline starts both real fetches in that
window and fails the new assertion. No artificial JavaScript error is injected.

Task queries now use `readTaskQuery` to check cancellation before calling the
existing API. `beforeunload` pauses new reads before cancelling current ones.
Pending reads resume after two animation-frame callbacks in a surviving
document; `pageshow` also releases the gate. `pagehide` cancels waiting queries
when the document actually leaves. A generation check prevents an earlier
frame callback from releasing a later pause, and signals are checked again
before a resumed read can call fetch.

The same guard covers task summary, shared lists, task detail/history, stage
history, firewall-task status and read-only Compose output queries. The firewall
status query joins the `['tasks']` cancellation family; its mutation and lease
confirmation remain unchanged. The guard does not cancel backend tasks,
mutations, uploads, terminal input or Docker operations. The last confirmed
query state is preserved, and ordinary polling resumes after prevented
navigation. No arbitrary settling sleep, force click, relaxed page-error
assertion, error filter, retry, skipped test or extended deadline was added.
Existing recovery and credential-exclusion assertions remain active.

The controlled event/timer replay is not native confirmation-dialog or bfcache
acceptance. Animation frames indicate a surviving document can render; they do
not prove every possible slow-navigation or background-tab sequence. The original
terminal refresh scenario remains in the Windows retest to check the actual
reported failure. Relevant contracts: [TanStack query cancellation](https://tanstack.com/query/latest/docs/framework/vue/guides/query-cancellation)
and [animation-frame callbacks](https://developer.mozilla.org/en-US/docs/Web/API/Window/requestAnimationFrame).

## Local validation

Evidence: `.local/evidence/windows-report6-2026-10-01/`. These are Linux-hosted
checks and are separate from the returned Windows results:

| Check | Actual result |
|---|---|
| Queued-polling baseline before the new guard | 0/1 passed; both summary and list fetches attempted, 12.295s |
| Corrected queued-polling case, WebKit repeated three times | 3/3 passed, 21.333s |
| Final combined WebKit, nine cases repeated three times | 27/27 passed, 267.153s |
| Combined Chromium | 9/9 passed, 78.202s |
| Combined Firefox | 9/9 passed, 100.685s |
| Shared lists: partial batch acceptance and upload recovery, Chromium | 2/2 passed, 28.094s; no remote operation replay |
| SystemApp firewall-task status/confirmation, refresh and unavailable controls, WebKit | 2/2 passed, 33.824s; API doubles, no host modifications |
| Frontend units | 80/80 in 16 files, 1.78s |
| Final frontend type check and production build | Passed; 3,739 modules, Vite 5.83s; existing large-chunk warning retained |
| Final PowerShell 7.4 runner harness | Passed; exact file/title selection and missing/incomplete WebKit-only results rejected |
| Runner functions with actual combined browser JSON | Passed; exact 9/9/27 title counts, all four narrow cases at three each, baseline failure diagnostic decoding |
| PowerShell-generated selection with the actual Playwright CLI `--list` | Exactly four required titles at three executions each (12); selection only, no tests run |

Final browser commands use the official Playwright `v1.63.0-noble` container,
fresh Vite (`BLORA_E2E_FRESH_SERVER=1`), one worker and the selected engine:

```text
node node_modules/@playwright/test/cli.js test tests/browser/accounts.spec.ts tests/browser/terminal.spec.ts tests/browser/query-navigation.spec.ts --workers=1 --reporter=line,json
```

WebKit adds `--repeat-each=3`. Known local Vite proxy warnings at navigation or
fixture teardown remain in the logs; a passing browser run is not full-stack
or Windows evidence, and those warnings are not hidden.

The combined runs validate the common guard and its existing consumers.
SystemApp's additional firewall-task status wiring is checked separately after
those runs; it does not change the shared guard or combined test consumers.

## Focused Windows retest

Use the pinned revision supplied in the conversation:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -Ref <commit> -Followup -BrowsersOnly -WebKitOnly
```

This selects the remaining fallback-renderer/worker terminal case plus all
three task-query navigation cases. Each must pass three executions: **12/12**
required. The runner installs only `web` dependencies and WebKit; Go/native/race,
compiler downloads, SDK packages, Chromium/Firefox, accounts and the other two
terminal configurations are deliberately omitted. Their prior passes are not
imported into the new result. The report records `runMode=followup-webkit`,
`selectedBrowsers=["webkit"]`, and false Go/native/SDK selection flags.

File/title matching accounts for Playwright's [full grep title](https://playwright.dev/docs/api/class-testconfig#test-config-grep),
including the file name. Per-title checks require three actual passes for each
selected case, plus an aggregate count of twelve, with no failure, skip, retry,
flaky or expected-failure substitution. Missing results from an omitted engine
do not become blockers or new passes. Original diagnostics and ZIP generation
remain enabled. The broader `-Followup -BrowsersOnly` remains available and now
requires **9/9/27**. See [Windows validation](../../operations/WINDOWS_VALIDATION.md).

The Windows diagnostic resolution still needs the next actual report. Full
Windows acceptance, the four opt-in Docker/PTY checks, native IME, Windows
service/task/firewall changes, system notification-center interaction,
Windows full-stack testing, cross-host/physical failures and strict E08 remain
unverified or incomplete as previously recorded. UI styles, icons, wallpaper
and preferences remain unchanged.
