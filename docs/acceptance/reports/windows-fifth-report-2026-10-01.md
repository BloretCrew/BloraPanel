# Fifth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`bae0894c4730ae6e76c0502d75fb7fc82f196d19d772bbb680640fb54c31b810`.
All **36 entries** in `SHA256SUMS.txt` match their actual ZIP contents; no
listed file is missing. The archive has 37 entries including the manifest.
The tested checkout is **`3c7abb3ef89f30d4bebd366019a72fe63c7f7eea`**,
`runMode=followup`, on Windows NT 10.0.26200 AMD64, 32 logical processors,
PowerShell 5.1.26100.9168. Run: **2026-10-01 17:36:06–17:55:00 Asia/Shanghai**
(09:36:06–09:55:00 UTC). No fatal runner error; original outcome **FAILED**.
Private machine paths and raw report logs are not published.

| Actual Windows selection | Passed | Failed | Skipped |
|---|---:|---:|---:|
| Eleven required native checks | 11 | 0 | 0 |
| Three named Go regressions, three executions each | 9 | 0 | 0 |
| Native events including subtests/helpers | 25 | 0 | 0 |
| Named Go retest events including subtests | 24 | 0 | 0 |
| Master/runlog race events including subtests/helpers | 87 | 0 | 4 |
| Chromium, accounts and terminal files | 6 | 0 | 0 |
| Firefox, accounts and terminal files | 6 | 0 | 0 |
| WebKit, both files repeated three times | 16 | 2 | 0 |

Event totals are separate from required case counts. Compiler download/checksum,
extraction/runtime probe, SDK builds and all three reference packages passed.
The Windows race stage took 436.63s; package durations were Master 405.848s and
runlog 5.495s. Browser JSON durations were Chromium 45.575s, Firefox 54.705s,
WebKit 356.358s. No browser retries or flaky outcomes were accepted.

## Verified corrections and remaining failures

Each of `TestCancelAcceptedUploadBeforeNodePreparation`,
`TestStdinSurvivesDaemonAndOutputEOF` and
`TestFinishRechecksDurableRecordAfterHelperExit` passed three executions in its
correct package. These are now actual Windows verification of the prior
upload-dispatch and durable-log completion corrections, superseding their
pending-retest statements in the [fourth report](windows-fourth-report-2026-10-01.md).
The WebKit account/permission confirmation scenario also passed all three
executions, together with the other two account scenarios: **9/9**.

WebKit terminal results were automatic-renderer/worker **2/3**,
fallback-renderer/worker **2/3**, fallback-renderer/main **3/3**. Both failures
occurred in the third repetition at the unchanged empty-page-error assertions
(`terminal.spec.ts` lines 84 and 109). The message is a Fetch API access-control
diagnostic for `/api/v1/tasks/summary` on the local Vite origin.
Both bundled diagnostics show a Playwright `pageerror` but no recorded native
`window.error` or `unhandledrejection`; the page was visible/focused with a
successful paint probe and a 1440×960 viewport. This does not establish that the
diagnostic is harmless or identify its Windows cause. WebKit maps JavaScript
console errors into Playwright page errors in the installed browser adapter;
the two channels therefore remain useful separate evidence.

Four opt-in race skips remain explicit gaps:

- `TestComposeProjectSaveAcrossDaemonRestartAndExplicitApply`
- `TestContainerNodeAPIRealEngineLifecycle`
- `TestHostDockerTerminalAPIActorBindingAndRefresh`
- `TestTerminalRealPTYRefreshLeaseAndRevocation`

## Request lifecycle correction

A controlled local test held the second actual task-summary request open,
changed the task filter and immediately refreshed. Before the correction the
original request did not receive an abort signal; the regression failed.
Afterward its real `AbortSignal` fires during `beforeunload`, the latest filter
restores, the new document reads a fresh count, and no write request is replayed.

Task summary, list/detail/history and stage queries now consume TanStack's
query cancellation signal. The shared list consumers in Instances, InstanceBatch
and Files, plus the read-only Compose output query, use the same signal.
`App.vue` cancels only the `['tasks']` query family at `beforeunload` and
`pagehide`. This cancels client reads, not backend tasks, task mutations, upload
submissions, terminal input or Docker operations. Existing recovery confirmation
and persistence remain intact. See the [TanStack cancellation contract](https://tanstack.com/query/latest/docs/framework/vue/guides/query-cancellation).

A second controlled test dispatches a prevented `beforeunload` in a live
document, verifies cancellation preserves the last confirmed count, and waits
for ordinary polling to show the next real response. Polling is not permanently
disabled when navigation is cancelled. This synthetic event is not native
confirmation-dialog or bfcache acceptance. There is no new settling delay,
error filter, force click, retry, skip, extended deadline or relaxed assertion.

Failure diagnostics now retain bounded timestamps, task request IDs and
start/finish/failure paths, navigation and page lifecycle signals. They retain
no field contents, request bodies, query parameters, storage dumps or traces.
These additions are diagnostic only and do not influence pass/fail.

## Local validation

Evidence is under `.local/evidence/windows-report5-2026-10-01/`. These are
Linux-hosted checks, separate from the returned Windows evidence:

| Check | Actual result |
|---|---|
| Controlled pending-read refresh before the correction | 0 passed / 1 failed; no abort recorded, 17.106s |
| Same corrected test, WebKit `--repeat-each=3` | 3/3 passed, 33.050s |
| Initial combined accounts/terminal/navigation WebKit run | 24/24 passed, 243.606s |
| Initial combined Chromium run before shared-list consumers were updated | 8/8 passed, 81.748s |
| Final combined Chromium run | 8/8 passed, 79.608s |
| Final combined Firefox run | 8/8 passed, 115.302s |
| Final combined WebKit run | 24/24 passed, 263.222s |
| Final frontend unit tests | 80/80 passed in 16 files, 3.26s |
| Final frontend type checking and production build | Passed; 3,738 modules, Vite build 8.95s; existing large-chunk warning retained |
| PowerShell 7.4 runner harness | Passed; browser-only omitted checks and incomplete-count rejection included |
| Runner functions parsing actual final browser JSON and baseline failure attachment | Passed; exact 8/8/24 titles and real diagnostic decoding verified |
| Shared list consumers: partial batch acceptance and checkpointed file upload, Chromium | 2/2 passed, 25.541s; no remote operation replay after refresh |

Final browser commands use the official Playwright `v1.63.0-noble` container,
fresh Vite (`BLORA_E2E_FRESH_SERVER=1`), one worker and `BLORA_BROWSER` set to
the engine. Selection:

```text
node node_modules/@playwright/test/cli.js test tests/browser/accounts.spec.ts tests/browser/terminal.spec.ts tests/browser/query-navigation.spec.ts --workers=1 --reporter=line,json
```

All final browser reports have zero failures, skipped cases, retries or flaky
outcomes. WebKit adds `--repeat-each=3`. Frontend checks are `npm test` and `npm run build`
from `web`; build includes `vue-tsc --noEmit`. The script harness runs
`scripts/test-windows-validation.ps1` in the official PowerShell
`7.4-ubuntu-22.04` container. Baseline failures and earlier successes remain
separate; they are not combined into an invented full-suite result.

The two shared-consumer regressions ran separately in Chromium with
`tests/browser/instance-batch.spec.ts tests/browser/files.spec.ts --grep
'per-view filters|upload reload uses'`. They use their existing isolated HTTP
doubles and are not part of the next eight-case Windows browser selection.
Their unchanged legacy notification transport still logs connection-refused
proxy warnings; these are not full-stack backend checks.

Initial and final local runs still showed Vite proxy warnings at teardown. Passing
the controlled cancellation regression and Linux suites does not prove that the
Windows access-control diagnostic has been resolved. The new Windows run
must confirm that independently.

## Next Windows selection and boundaries

Use the next pinned revision with **`-Followup -BrowsersOnly`**. This downloads
a fresh checkout, installs only web dependencies and matching browser engines,
and requires all **8 Chromium / 8 Firefox / 24 WebKit** executions. Go/native
checks, GCC downloads, SDK packages, unrelated browser files and standalone
builds are deliberately not rerun. `runMode=followup-browser` and three explicit
selection flags distinguish omitted checks from verified passes. Prior Windows
passes are not imported. The report ZIP and failure diagnostics remain enabled.
See [Windows validation](../../operations/WINDOWS_VALIDATION.md).

This report closes the three named Windows Go corrections and provides new
account recovery evidence. It does not close full Windows/platform acceptance,
the four opt-in checks, native IME, Windows service/task/firewall modification,
system notification-center interaction, Windows full-stack browser testing,
external Engine/network scenarios, physical failures or strict E08. UI styling,
icons, wallpaper and preferences remain unchanged.
