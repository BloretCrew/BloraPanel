# Fourteenth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`cb3e0ad978f2360027e3687b42807bcd9f4e4a0a3190947b404cbe6863774d9a`.
All **17 entries** in `SHA256SUMS.txt` match the complete, unique ZIP entry set;
the ZIP contains 18 entries including the manifest. Requested and tested ref:
**`63cacc0ef92a25dd84f4d534e1f767323d8b3965`**.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-04 21:59:04–22:06:45 UTC+8** (13:59:04–14:06:45 UTC).
Schema 2, outcome **FAILED**, no fatal runner error;
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]`,
`browserServerMode=production-preview`, `browserBuildSelected=true`.
Private machine paths, screenshots, traces and uploaded raw logs are not published.

| Actual Windows selection | Passed | Failed / timed out | Skipped |
|---|---:|---:|---:|
| Pending summary cancellation and restored task filter | 1 | 2 | 0 |
| Confirmed count and polling after prevented navigation without RAF | 3 | 0 | 0 |
| Real queued task polling at departure | 3 | 0 | 0 |
| Instance-log polling, retained command draft and no input replay | 3 | 0 | 0 |
| Terminal fallback renderer / worker checkpoints | 2 | 1 | 0 |
| Total | **12** | **3** | **0** |

The fresh type-checked production build passed in **24.36s**. WebKit installation
passed in **2.42s**. Browser stage **408.15s**, JSON duration **405.638s**.
There are no retries, skips, flaky outcomes or global startup errors.

The OS-selected port **3570** is recorded in the runner and actual Playwright
configuration. The fresh compiled server started at `http://127.0.0.1:3570`, with
`--strictPort` and `reuseExistingServer=false`. The
[thirteenth report's](windows-thirteenth-report-2026-10-04.md) startup collision
is resolved for this actual Windows run; it no longer prevents test execution.

The corrected queued-poll fixture now obtains **3/3 Windows passes** with the
real browser departure boundary and an initial count of zero, rather than a
pre-boundary ordinary poll satisfying the assertion. No-frame recovery and
independent log lifecycle also pass three times. These are revision-specific
functional results, not paint/performance or full Windows acceptance.
Go/native/race, SDK, Chromium and Firefox were omitted, not newly passed.

## What remains unresolved

The first pending-summary/filter execution passed in **37.502s**. The second
failed in **36.229s**: after real reload, the original **5s** value assertion
could not find the task-state select. This is a missing control at the deadline,
not evidence of a wrong restored value. Its phases show initial summary at
**6.324s**, confirmed task view at **6.504s**, and refresh starting at **9.431s**.
The held read is cancelled at departure. In the new document, actual summary JSON
consumption and DOM count 7 are recorded promptly; the task-list request is only
observed about **18.8s after the new navigation**. The later surface probe reports
`paint=false`. Event delivery times, a later paint probe and network completion
cannot identify the cause of delayed application mounting or prove that all
frames were absent during the failed assertion. No native JavaScript failure
is recorded. Filter recovery within the original limit remains unverified.

The third pending-summary execution failed in **26.054s**, **before departure**:
the original 5s preparation poll did not see the deliberately held summary read
in time. The task view and filter were already ready. A later summary-request
event is present; natural polling and automation were delayed. This is not a
demonstrated cancellation or restored-filter failure, and its Windows cause
is not established. Fixture setup itself consumed about **13.2s** before the
first recorded operation in this execution.

The terminal's first/third executions passed in **28.095s** and **23.188s**.
Its second execution exhausted **45s**, actual result **45.473s**, at the native
pointer action on the control-takeover button. Initial desktop navigation and
readiness consumed about **35.2s**; first refresh/readonly state succeeded and
takeover began at **41.667s**, leaving only about 3.3s. Initial scripts/styles
include observed resource durations around **10.8s**; the later surface probe
has `paint=false`. These observations do not establish whether the initial delay
was the browser, host scheduling, asset serving, security software or product
initialization. There is no justification to change terminal protocol/ACK/input
logic from this evidence. Pointer readiness and the total deadline remain real
unresolved failures.

## Terminal repair and next display comparison

The first local headed comparison retained the original Windows selection and
failed **12/15**, in **175.945s**. Task/log cases each passed three times, but
all three terminals stalled at the second reload's 5s readonly-state assertion.
These failures, JSON, diagnostics, screenshots and traces remain separately
stored; they are not relabeled as passes. This is a different failure from the
uploaded Windows takeover/initialization timeout.

A private same-case Worker probe reproduced the stall twice (both **0/1**).
The large serialized screen (**155,971 characters**) and color-write responses
completed; the next control-state message carried a Vue reactive proxy and
`Worker.postMessage` threw **DataCloneError**. Worker termination returned
immediately; the application dropped its checkpoint mirror. This evidence
identifies the cloning defect, not every part of the subsequent restoration
delay or the uploaded Windows failures' cause.

Restore now copies checkpoint controls and OSC-link metadata to plain JSON data
before crossing the Worker boundary. Validation, browser parsing, sequential
output replay, budgets, protocol, durable ACK and no-input-replay behavior stay
intact. Existing terminal cases additionally require the expected Worker/main
checkpoint mode after both real reloads. Control-sequence and link-continuation
cases now require that the real Worker remains present after remount; silent
fallback cannot conceal this regression. No journal batching, timeout increase,
UI change or new transport behavior is introduced.

The runner adds optional **`-Headed`**, using Playwright's actual `--headed`
argument. It opens visible browser windows in the signed-in Windows desktop.
It does not use `--debug`, skip cases, retry, prewarm, force clicks, change UI or
raise any deadline. Default invocation remains the existing non-headed mode.
This comparison tests a possible environment effect; it is **not a claimed
explanation of the Windows failures**. The independent cloning repair above is
not evidence that those three uploaded failures have been resolved.
`-Headed` is limited to `-Followup -BrowsersOnly -WebKitOnly`. Inherited `PWDEBUG`
is isolated and restored afterward to keep debug settings from changing limits.

Both modes use the same argument builder. Removing the sole `--headed` argument
produces exactly the original narrow CLI selection, including the anchored five
titles, one worker, three executions per case and unchanged artifact handling.
The fresh build/server/port gates and **45s total / 5s assertions** remain.
`browserDisplayMode` records the requested mode in `report.json` and its README.
For headed runs, each execution records one static annotation from Playwright's
resolved `headless` launch fixture. JSON project metadata omits `use.headless`.
Missing, duplicate, mismatched or malformed per-execution mode evidence is
BLOCKED even if all nominal cases pass. Resolved modes remain in the browser
summary for review. This passive evidence does not change browser/test behavior.

The next [pinned Windows run](../../operations/WINDOWS_VALIDATION.md) uses
`-Followup -BrowsersOnly -WebKitOnly -Headed`, with the desktop unlocked and the
automated windows allowed to run. It still requires a fresh build PASS and
**five exact titles × three executions = 15/15**. A visible-window pass would
establish that tested environment only; it would not erase this failed default
run or certify headless reliability, cold initialization, pointer latency or E08.

## Local validation

Evidence: `.local/evidence/windows-report14-2026-10-04/`.
Actual PowerShell 7.4 Linux-container runner harness passed, including the new
argument parity, requested/actual display-mode recording and mode-mismatch
rejection, with all existing port/global-error, build/title/execution, progress,
literal-argument, timeout/skip, diagnostic, ZIP and interrupted-run gates intact.
Actual runner functions generate the arguments used by browser checks.

The pinned official Playwright `v1.63.0-noble` image provides `--headed` and
`xvfb-run`; Linux's isolated virtual display verifies the GUI code path, not
Windows hardware or a physical display. Following the cloning repair, type
checking and a **new** production test bundle build passed (**3,740 modules**,
Vite **7.73s**); the existing large-chunk warning remains. Final source type
checking passed, and **90 unit tests / 17 files** passed in **1.850s**.

With the same fresh selected port and one worker, in separate sequential runs:

| Actual local execution | Passed | Failed / skipped | JSON duration |
|---|---:|---:|---:|
| Repaired single headed terminal, including retained Worker after both reloads | 1 | 0 | 43.532s |
| Exact five-case Windows selection, headed, three executions each | 15 | 0 | 181.922s |
| Same selection, default headless, three executions each | 15 | 0 | 91.029s |

Both 15-execution JSON reports contain exactly one matching resolved display
annotation per execution, no retries/flaky/global errors, and the selected fresh
server URL with `reuseExistingServer=false`. Headed mode is not a local speedup
claim: the visible Linux run is slower. The environment comparison on Windows
remains necessary and uncertain.

Actual PowerShell functions validate both real successful JSONs and the uploaded
12/15 failure: its two incomplete title counts and three bounded failure
diagnostics remain; Go/native/SDK and omitted engine passes are not invented.
The real finalizer preserves **FAILED** and makes a private evidence ZIP.
It accepts the actual headed 15/15 JSON, then rejects the same JSON when one
launch annotation is removed, despite nominal passes. Extended terminal
continuation on six source-import browser files passed **62/62 WebKit** in
**231.936s**, including the restored Worker assertions, CSI/OSC/charset/cursor/
mouse/palette/link continuation, UTF-8 tails, resize/scrollback, stale snapshots,
killed-Worker fallback and connection generation. This run uses a fresh development
server because those suites import the real service modules, not the production
preview bundle. Firefox's two-file production regression passed **7/7** in
**66.054s**, covering all three renderer/checkpoint variants and the four task/log
cases. Chromium's same production regression passed **7/7** in **51.943s**.
All three engine regressions use one worker, fresh isolated servers and no
retries/skips/flaky/global errors, with the engines run sequentially. All owned
local test containers and sessions have ended; no running session needs takeover.
Final PowerShell harness passed, including missing/duplicate/mismatched launch
evidence rejection and the original coverage/port/build/progress/ZIP gates.
Existing test-server teardown proxy warnings do not certify any real
daemon or performance result.

Reproduction commands, from the repository's `web/` directory (browser commands
ran in the pinned Linux Playwright image; `--headed` used `xvfb-run -a`):

```text
npm run build -- --mode test --outDir .local/browser-test-build
npm run check
npm test
```

The exact narrow command is generated by `Get-BrowserArguments` with
`Get-WebKitFollowupPlan` and its anchored `Get-BrowserTitlePattern`, then run with
`BLORA_BROWSER=webkit`, `BLORA_E2E_PORT=43301`, `BLORA_E2E_PREVIEW=1`,
`BLORA_E2E_FRESH_SERVER=1`. Both modes use `--workers=1 --repeat-each=3
--reporter=line,json`; only the visible command adds `--headed`.
The six-file terminal continuation command uses `--workers=1 --reporter=line,json`
and `terminal-parser-recovery`, `terminal-links-recovery`,
`terminal-theme-recovery`, `terminal-snapshot`, `terminal-mouse-recovery`,
`terminal-generation` under `tests/browser/`, with development serving for
actual source imports. Firefox/Chromium run `query-navigation.spec.ts` and
`terminal.spec.ts` against production preview, once each.

`pwsh -NoProfile -File scripts/test-windows-validation.ps1` ran in the pinned
PowerShell 7.4 Ubuntu 22.04 image from the repository root. The private
`browser-comparison.mjs`, `terminal-regressions.mjs`, JSON/log artifacts and
`notes.md` retain actual arguments, run isolation and failed-versus-passed
boundaries; the private PowerShell JSON check applies the runner's real AST
functions and finalizer to uploaded/local JSON, not a reimplementation.

Final difference checks passed; **1,151 local Markdown links** were checked with
no missing targets. Raw uploaded logs and local failed-test artifacts remain
private; only this sanitized report, code/tests, runner and acceptance updates
are committed for delivery.

Full Windows notification/IME, authorized Engine/SCM/task/firewall lifecycle,
physical power/cache loss, cross-host combinations and the unchanged E08
p95 ≤50ms target remain open. Linux browser passes do not close these gaps.
