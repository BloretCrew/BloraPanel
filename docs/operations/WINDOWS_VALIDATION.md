# Run Windows validation and return the report

This runner downloads a fresh checkout of BloraPanel from GitHub and runs the
existing automated Windows checks without remote access. It never updates an
existing checkout. Run it in **Windows PowerShell**, not WSL, on Windows 10 1809
or newer / Windows 11, preferably x64. Use a normal terminal first; do not disable
antivirus or change execution policy globally when a test fails.

## One command to start

Paste these lines into PowerShell:

```powershell
$script = Join-Path $env:TEMP 'Blora-Windows-Validation.ps1'
Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/BloretCrew/BloraPanel/main/scripts/windows-validation.ps1' -OutFile $script
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -InstallTools
```

`-InstallTools` installs missing Git, Go and Node.js LTS through `winget` (normal
installer/elevation prompts may appear). If these are already installed it does
not replace them. Without winget, install Git, Go **1.25+**, and Node.js **24+**,
open a new terminal, and rerun. Network access to GitHub, Go/npm registries and
Playwright browser downloads is required. Allow several GB for the checkout,
dependencies, browser engines and logs; execution can take an hour or more.

Windows Go race testing additionally requires a compatible **mingw-w64 GCC**.
Pass `-InstallRaceCompiler` to download a pinned, SHA-256-verified portable
WinLibs compiler into this run's directory on Windows x64 (about 261 MB download;
allow additional extraction space). It does not install a system compiler,
change the registry, or persist PATH changes. Otherwise the runner uses an
existing `gcc.exe` on PATH. A missing compiler is **BLOCKED**, and a broken or
incompatible compiler records a failure, never a pass. The archive version and
official release digest are pinned in `scripts/windows-race-compiler.ps1`.

Optional parameters:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -WorkRoot 'D:\BloraValidation' -Ref main
```

Each invocation creates a new timestamped directory. `-Ref` can be a branch,
tag or commit; the exact fetched commit is included in the report. The console
shows the current stage, elapsed time and live test output. A failed stage does
not stop independent later stages. Individual commands have time limits.
Progress uses plain text (including a 15-second heartbeat), not the Windows
PowerShell progress renderer. Child-shell progress is suppressed to avoid CLIXML
progress noise; ZIP creation uses .NET directly rather than `Compress-Archive`.

## Follow up the ninth returned report: remaining WebKit checks

The [ninth returned report](../acceptance/reports/windows-ninth-report-2026-10-03.md)
passed WebKit **14/15**. All four navigation/log scenarios passed three times,
including both eighth-report failing cases. The first fallback-renderer/worker
terminal execution exhausted the original 45s test budget without a recorded
page error; the later two full terminal executions passed in about 19s each.

The narrow followup now type-checks and builds the production frontend in a
separate logged stage, then tests that compiled bundle on a fresh local preview
server. This removes on-demand development compilation from the case budget.
It does not pre-open or warm up the terminal, change the 45s limit, add retries,
or omit any original UI action, refresh, worker, recovery, ACK or no-replay check.
The uploaded timing does not establish the Windows cause; the new run checks
the compiled application and provides more precise diagnostic evidence.

Download the runner from the **same pinned commit** supplied in the conversation,
then use this narrower selection:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -Ref <commit> -Followup -BrowsersOnly -WebKitOnly
```

It needs Git and Node.js, creates a fresh checkout, installs only `web`
dependencies, builds the production frontend, installs the matching WebKit
engine, and selects these five cases:

- Fallback-renderer/worker terminal checkpoint recovery, parsed-byte ACKs and
  no input replay, retained to guard the previously corrected native refresh.
- Pending task-summary cancellation before a real refresh, with the latest
  task filter restored.
- Last confirmed task count and continued polling after prevented navigation,
  deliberately suppressing new frame callbacks until the unchanged 5s assertion passes.
- Real queued QueryObserver polling callbacks cannot start task-list/summary
  reads during departure; a surviving document resumes them.
- Independent instance-log polling cancels pending reads, holds queued
  callbacks, resumes its run list, and restores command drafts without input replay.

Every case runs **three times**, requiring **15/15** actual passing executions.
The file/title selection is anchored; the report checks every required title
and aggregate count. Retries, skips, expected failures and missing results
cannot count as passes. Original terminal/page-error assertions remain active.
The queued-callback observation covers the navigation event turn and its
microtasks, independently of paint. The bounded guard and controlled
prevented-navigation checks do not prove every slow navigation, native dialog
or bfcache sequence. Linux results do not establish resolution on Windows.

Chromium/Firefox, account scenarios, the other terminal configurations,
Go/native/race, compiler downloads and SDK packages are deliberately omitted.
No Go installation or `-InstallRaceCompiler` is needed. The report records
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]` and the three
Go/native/SDK selection flags as `false`, plus
`browserServerMode=production-preview` and `browserBuildSelected=true`.
The **`web-browser-build`** stage must pass; a failed or omitted build prevents
success and never falls back to stale output or development serving. The
preview binds only to loopback and fails if port 5173 is occupied.
It does not import earlier passes or
require JSON reports from omitted engines. `-WebKitOnly` requires both
`-Followup` and `-BrowsersOnly`. Progress, deadlines, failure diagnostics and
ZIP creation remain enabled; return `Blora-Windows-Report.zip` even on failure.

Terminal tests display static operation labels and elapsed milliseconds as they
progress. If a case fails, the existing bounded diagnostic JSON now also carries
these phases and up to twenty slowest static asset request durations. Field
contents, terminal input text, request bodies, storage dumps and browser traces
remain excluded from the diagnostic bundle.

## Broader browser-only followup

The [fifth returned report](../acceptance/reports/windows-fifth-report-2026-10-01.md)
confirmed all eleven native checks and the three named Go regressions on Windows,
with no Master/runlog race failures. Two WebKit terminal executions still failed
the unchanged page-error assertions. To rerun all affected engines, download the
runner from the **same pinned commit** supplied in the conversation, then use:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -Ref <commit> -Followup -BrowsersOnly
```

This mode needs Git and Node.js, makes a fresh checkout, installs `web`
dependencies and the matching browsers, and runs all ten current scenarios in
`accounts.spec.ts`, `terminal.spec.ts` and `query-navigation.spec.ts`.
Chromium and Firefox require **10/10** each; WebKit repeats every scenario three
times and requires **30/30**. The original `f98b078` selection required 8/8/24;
`8ad386a` added queued task polling (9/9/27), and the current version adds
independent instance-log polling. Actual counts and titles
must pass, without retries, skips or expected failures. The new controlled
checks cover aborting a pending task summary before a real refresh while
restoring the latest filter, and continuing polling after a prevented
`beforeunload` event. The latter is not native confirmation-dialog or bfcache
acceptance. Original password exclusion, checkpoint, ACK/credit, no-replay and
page-error assertions remain active.

Go/native/race, compiler downloads, SDK packages and unrelated browser files
are deliberately omitted. No Go installation or `-InstallRaceCompiler` is needed
for this selection. The report records `runMode=followup-browser` and
`goChecksSelected`, `nativeChecksSelected`, `sdkChecksSelected` as `false`;
it does not import earlier passes as new evidence. `-BrowsersOnly` requires
`-Followup`. Progress, stage deadlines, failure diagnostics and report ZIP
creation still work in this mode; return `Blora-Windows-Report.zip` even if a
test fails. Local passes do not establish resolution of the Windows diagnostic.

## Retest the returned Windows failures

After the third returned report, use the updated runner with:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -Retest -InstallRaceCompiler
```

This mode creates a fresh checkout and runs:

- The 27 named Go regressions: all 22 failures from that report, a new Windows
  directory-attribute check, the previously Unix-only two-node lifecycle, and
  three connection-loss / transfer-resumption checks.
- Eleven required native checks, including ConPTY command execution and the
  new file/directory readonly attributes and timestamp check.
- All-package Go race tests, previously blocked by the missing compiler.
- All scenarios in nine affected browser files on Chromium, Firefox and
  WebKit, including editor history/save/split, account recovery, reduced motion,
  extensions and terminal recovery. The unchanged Chromium-only native IME
  injection test is excluded from this targeted run; cross-browser committed
  Unicode and multicursor tests remain included.
- SDK builds and all three independent reference packages needed by these tests.

Each required Go case must emit PASS in its correct package. Each browser must
pass every named required regression without retries, skips or expected-failure
substitution. A nonempty but incomplete selection is **INCOMPLETE**. The report
records `runMode=retest`, the missing case names, and every intentional omission.
Previous results are not imported as new passes. This is a focused regression
run, not complete Windows platform acceptance. Full race testing may still
report legitimate opt-in/environment skips, which are kept visible.

Use `-Ref <commit>` to pin the source revision supplied in the conversation;
download the runner from that same revision. `-InstallTools` is optional if Git,
Go or Node.js is missing. Normal test objects are isolated in the new work root.

## Follow up the fourth returned report

After the [fourth returned report](../acceptance/reports/windows-fourth-report-2026-10-01.md),
use the revision supplied in the conversation and **`-Followup
-InstallRaceCompiler`** instead of repeating the earlier 27-case/all-package
run:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -Ref <commit> -Followup -InstallRaceCompiler
```

Followup makes a fresh checkout and builds the SDK packages. It requires three
executions of each of the two new Go failures and the new durable-finish
interleaving regression, all eleven native checks, and full race regression
of **Master and runlog**. The original `3c7abb3` runner selected six browser
scenarios (6 / 6 / 18). The `f98b078` runner added two query-navigation cases
(8 / 8 / 24). The `8ad386a` runner added queued task polling (9 / 9 / 27).
The current runner also adds independent instance-log polling, so it
selects ten scenarios in `accounts.spec.ts`, `terminal.spec.ts` and
`query-navigation.spec.ts`: once each on Chromium/Firefox, three times on WebKit
(10 / 10 / 30 passing executions required). These browser tests still use real
trusted actions, actual xterm/Monaco/storage and the original strict recovery,
password exclusion, ACK/credit and no-replay assertions. They use transport
doubles, including the independent notification stream, not a Windows
full-stack backend.

The report records `runMode=followup` and the exact intentional omissions.
Other packages/browser files, standalone builds/vet, production web/unit checks
and native IME are not rerun or imported as passes. Existing opt-in skips are
visible and can leave the overall outcome **INCOMPLETE** even when this
selection has no failure. `-Followup` and `-Retest` are mutually exclusive.

When a selected browser case fails, the ZIP also includes bounded
`browser-<engine>-diagnostic-<number>.json` files extracted from its actual
Playwright attachments. They record exception/request paths, page visibility,
paint readiness and control geometry, plus bounded task-request and page
lifecycle timing; no input contents, storage, credentials or traces.
JSON-escaped run/profile paths are redacted too. The Linux fixture
corrections do not themselves prove the Windows-specific WebKit cause fixed;
return the ZIP even if the same test still fails.

## Remaining mode and interrupted-run recovery

For the next run after the first returned report, download the updated script
and use `-Remaining`:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -Remaining
```

This makes a fresh checkout and repeats SDK packaging, native checks, full Go
regression (including known unresolved failures), race checks when GCC is
available, and all three browser suites. It omits standalone Go builds/vet and
web build/unit tests that already passed. The report records `runMode=remaining`
and explicitly lists omissions; previous passes are not imported into this run.
This is a diagnostic/retest run, not a claim that all outstanding platform
acceptance is automated. Report every failure in the returned ZIP.

Download the current script using the first two lines above, then run:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -CollectLatest
```

This only packages existing evidence from the most recent run under the default
work root; it does not repeat tests or infer that unfinished tests passed. It
prints a new `Blora-Windows-Recovered-<id>.zip` path. For a custom location use
`-WorkRoot 'D:\BloraValidation' -CollectLatest`, or select the exact old run with
`-CollectRun 'D:\BloraValidation\<timestamp-id>'`. Keep the original logs and send
the recovered ZIP before rerunning. The recovery marker explicitly records that
the run may have been interrupted. Do not collect a run that is still active.

## What runs

- Windows Master, Daemon and extension signing tool builds; Go vet.
- Explicit native Job/keeper/daemon-exit, ConPTY Unicode/resize/close, log pipe
  recovery, monitoring/process identity and SCM/Task Scheduler enumeration tests.
  File/directory readonly attributes, timestamps and object identity are checked.
  The report checks that each expected native test actually emitted PASS.
- All Go package tests without cached results, then all race tests if GCC exists.
- SDK and independent reference extension builds/packages.
- Frontend build and unit tests; complete existing browser suites on downloaded
  Chromium, Firefox and WebKit, one browser at a time. Port 5173 must be free; an
  unrelated running server is never silently reused.

The runner clears inherited Blora test opt-ins so it cannot accidentally use a
previous Engine endpoint or privileged test fixture. Native tests create their
own processes/files and perform read-only service/task queries. It does not
change existing services, scheduled tasks, firewall rules or system time.

## Send back the report

At the end the console prints the absolute path to **Blora-Windows-Report.zip**,
normally under `%USERPROFILE%\BloraValidation\<timestamp-id>\`. Send that ZIP
back in this conversation, including when tests fail. It contains:

- `README.md`: stage summary, skip and coverage gaps.
- `report.json`: exact commit, OS/tool context, durations/exit codes, test events.
- Per-stage streaming logs and browser JSON results; SHA-256 checksums.

The ZIP excludes the checkout, caches, test state and browser traces. User/work
paths are scrubbed from ordinary text logs, but test output can still contain
machine identifiers or escaped paths; review the bundle before sharing. Do not
send the entire work directory. `steps-so-far.json` and logs are saved during the
run; forcibly closing the terminal or killing PowerShell can prevent final ZIP
creation. In that case send the `report` folder after reviewing it.

Exit 1 means failures, exit 2 means incomplete/skipped/missing coverage. Even
exit 0 is named `AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS`, not full acceptance.
The report is the authoritative list. Timeout cleanup targets only the runner's
child process tree; keeper/breakaway behavior can leave test-owned processes
after a failed test, so inspect failures before rerunning. Never kill unrelated
Daemon/business processes to clean up a test.

## Explicit coverage limits

Service/task **enumeration** is not service/task modification lifecycle proof.
Windows system notification center interaction, actual firewall rollback,
physical power loss, device cache loss, remote Engine/certificate/network
operations and long-duration mixed-load E08 are not automated here. Existing
real-browser fixtures depend on Linux `/bin/sh` and `/proc`; browser suites in
this script exercise actual browsers/xterm/storage with API doubles, not a
Windows Master/Daemon browser end-to-end fixture. Opt-in Docker tests are
reported as skipped. These gaps stay visible even if every executed test passes.

The runner's parser/control flow/report packaging can be checked on Linux with
`pwsh -NoProfile -File scripts/test-windows-validation.ps1`. That harness is not
evidence that Windows native behavior has passed; the returned Windows report
provides that evidence.
