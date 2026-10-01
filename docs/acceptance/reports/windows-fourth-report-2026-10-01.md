# Fourth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`9199f2b17f927b533040a4ca89a86e9db6db55aaa12c6641ab4458845437b213`.
The exact tested checkout is `08db0995e6221933b77e202834d6a187a11649b1`,
in `retest` mode on Windows NT 10.0.26200 AMD64, 32 logical processors,
PowerShell 5.1.26100.9168. Run: **2026-10-01 15:41:49–16:11:22
Asia/Hong_Kong** (07:41:49–08:11:22 UTC). No fatal runner error.
Private machine paths and original raw logs are not published.

| Selection | Passed | Failed | Skipped |
|---|---:|---:|---:|
| Chromium, nine affected files | 32 | 0 | 0 |
| Firefox, nine affected files | 32 | 0 | 0 |
| WebKit, nine affected files | 29 | 3 | 0 |
| Go all-package race test events, including subtests/helpers | 356 | 2 | 14 |
| Named Go retest events, including subtests | 36 | 0 | 0 |
| Windows native events, including subtests/helpers | 25 | 0 | 0 |

All **27 required Go regressions** and **11 required native checks** passed.
Event totals are not counts of those required checks. The portable compiler
download, checksum validation, extraction and GCC runtime probe also passed.
The report outcome is **FAILED**, rather than full acceptance.

## Newly verified on Windows

The previously unverified metadata attributes/times, readonly/writable
directory behavior, real ConPTY command/Unicode execution and resize/Job close,
portable lifecycle/console/log/metrics/schedule fixtures, directory transfers,
source-daemon restart proof rebuilding and bridge transport-loss classification
now have actual Windows PASS evidence. SDK build and all three package fixtures
passed. Chromium and Firefox passed every selected browser scenario without
skips or retries.

This supersedes the *pending Windows retest* statements in the earlier
[fix report](windows-report-fixes-2026-10-01.md) for those particular cases;
it does not promote their entire feature/acceptance rows to completed.

## Remaining failures and corrections

- `TestCancelAcceptedUploadBeforeNodePreparation`: generic dispatch raced the
  bulk cleanup of a client upload. The daemon correctly had no task receipt,
  but its missing receipt turned the cancellation into `INTERRUPTED`.
  Unsent client-upload cancellations now remain owned by bulk cleanup. An
  already dispatched commit still uses normal daemon reconciliation. The test
  holds the cancel intent through several actual dispatcher ticks before
  completing cleanup with the original cancellation identity.
- `TestStdinSurvivesDaemonAndOutputEOF`: the log helper could sync completion
  and exit between reading its record and probing liveness. Read/Finish now
  reload the record after exit; only completion with the same birth identity
  and token counts. Missing completion, archive failure or replacement still
  fails. Deterministic interleaving covers all of these. Local repeats also
  exposed reading a still-being-written fixture result and Linux `/proc`
  returning `ESRCH` during exit; both were corrected without changing deadlines
  or production stdin delivery.
- WebKit account/permission recovery: the normal pointer confirmation waited
  for an enabled/stable button until the 45-second test limit. The fixture now
  brings the page forward, checks both filled credentials and scrolls the
  actual button before its unchanged trusted click. Password exclusion,
  immediate refresh and exact write counts remain asserted. The Windows
  pointer-stability cause is not yet proven fixed.
- WebKit's automatic/worker and fallback/worker terminal scenarios reported
  `Fetch API cannot load ... /tasks[/summary] due to access control checks`.
  Local repetition reproduced the same background-request diagnostic, plus
  occasional large-checkpoint reattachment delays. Browser transport doubles
  now include the independent task-notification stream, correctly shaped task
  summaries, and HTTP routes attached to the context rather than the old page.
  No page-error, checkpoint, credit or no-replay assertion is removed. Native
  exception/request/paint diagnostics are bundled when a case fails; Windows
  behavior still requires the next report.

## Local validation of these corrections

The final source passed the following checks. These are Linux-hosted results,
not a replacement for the next Windows run.

| Check | Result |
|---|---|
| Three named Go corrections, `-race -count=8` | Passed; Master 12.760s, runlog 1.783s |
| Complete `internal/master` and `internal/runlog`, `-race -count=1 -p=2` | Passed; Master 426.988s, runlog 2.542s |
| Final affected browser files, Chromium | 6/6 passed, 59.158s |
| Final affected browser files, Firefox | 6/6 passed, 66.321s |
| Final affected browser files, WebKit `--repeat-each=3` | 18/18 passed, 169.378s |
| Windows amd64 test cross-compilation, both changed Go packages | Passed; not Windows execution |
| Go vet for both changed packages and frontend type checking | Passed |
| PowerShell 7.4 function harness | Passed; includes followup counts, failure diagnostics, ZIP/recovery and broken-progress-host regression |

Go commands, from the project root:

```text
go test -race -count=8 -timeout=5m ./internal/runlog ./internal/master -run '^(TestFinishRechecksDurableRecordAfterHelperExit|TestStdinSurvivesDaemonAndOutputEOF|TestCancelAcceptedUploadBeforeNodePreparation)$'
go test -race -count=1 -p=2 -timeout=12m ./internal/runlog ./internal/master
go vet ./internal/master ./internal/runlog
```

Browser commands used the official Playwright `v1.63.0-noble` container, a fresh
Vite server, `BLORA_BROWSER` set to each engine, one worker, and
`tests/browser/accounts.spec.ts tests/browser/terminal.spec.ts`. WebKit added
`--repeat-each=3`; all final reports show zero skipped, unexpected or flaky cases.
`npm run check` ran in `web`. The script harness used the official PowerShell
`7.4-ubuntu-22.04` container with `scripts/test-windows-validation.ps1`.

The first concurrent local runs reported Firefox 5/6 and WebKit 16/18, including
the same background-request diagnostic and checkpoint delays. Those reports
were retained; they are not combined with subsequent passes into an invented
full-suite result. Final runs were separate, with context-level HTTP doubles.
Local JSON evidence is under `.local/evidence/windows-report4-2026-10-01/`.
Neither this container sequence nor cross-compilation establishes the cause
or resolution of the Windows pointer-stability failure.

## Boundaries

The 14 race skips include explicit Docker/Compose/host-Engine opt-ins and
Linux-only firewall/PTY cases. They remain gaps, not passes. Native IME,
privileged Windows service/task/firewall lifecycle, system notification-center
interaction, Windows full-stack browser acceptance, external Engine/network
scenarios, physical failures and strict E08 performance remain outside this
report's passing evidence. UI artwork and preferences are unchanged.

The next `-Followup` runner creates a fresh pinned checkout. It repeats the
three named Go corrections three times, reruns the eleven native checks,
executes complete race regression in Master/runlog and all six scenarios in
the two affected browser files (Chromium/Firefox once; WebKit three times).
It requires the exact execution counts and never imports previous passes.
See [Windows validation](../../operations/WINDOWS_VALIDATION.md).
