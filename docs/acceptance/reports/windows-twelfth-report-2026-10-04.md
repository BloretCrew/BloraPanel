# Twelfth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`7bc7c8f265e38f3838f3772496e90b109b153fa371af72fa0f8614270304d31b`.
All **17 entries** in `SHA256SUMS.txt` match the complete, unique ZIP entry set;
the ZIP contains 18 entries including the manifest. Requested and tested ref:
**`90658d02095721ea904c099ec00b92ff11343f67`**.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-04 16:25:38–16:33:28 UTC+8** (08:25:38–08:33:28 UTC).
Schema 2, outcome **FAILED**, no fatal runner error;
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]`,
`browserServerMode=production-preview`, `browserBuildSelected=true`.
Private machine paths and uploaded raw logs are not published.

| Actual Windows selection | Passed | Failed / timed out | Skipped |
|---|---:|---:|---:|
| Pending summary cancellation and restored task filter | 2 | 1 | 0 |
| Confirmed count and polling after prevented navigation without RAF | 3 | 0 | 0 |
| Real queued task polling at departure | 1 | 2 | 0 |
| Instance-log polling, retained command draft and no input replay | 3 | 0 | 0 |
| Terminal fallback renderer / worker checkpoints | 3 | 0 | 0 |
| Total | **12** | **3** | **0** |

The type-checked production build passed in **22.71s**. Browser stage
**404.70s**, JSON duration **401.704s**. No retries, skips or flaky outcomes.
Terminal executions passed in **34.762s**, **38.298s**, **38.435s**; log cases
in **12.993s**, **23.858s**, **29.463s**; no-frame recovery cases in **21.426s**,
**17.703s**, **25.794s**. Those are actual Windows successes for this revision.
They do not establish pointer/paint performance or every cold-start scenario.

The runner rejects the aggregate and both incomplete task titles. Go/native/
race, SDK, Chromium and Firefox were omitted, not newly passed. Prior evidence
remains associated with its tested revision. The one queued-poll pass also has
the fixture-transition limitation described below.

## What remains unresolved

The first pending-summary/filter case exhausted the **45s total budget** at
real refresh (**45.497s**). Its initial summary was confirmed at **9.193s**;
the next phase, after task-app activation and pending-read setup, was only at
**34.686s**. Refresh began at **38.682s**, leaving about 6.3s. Static task-view
modules took about 3.9s, but the old phases cannot isolate how much of the other
setup time was pointer actionability, activation, polling or automation delay.
The failure attachment lacks the final refresh phase that appears in stdout and
records no departure event. This partial diagnostic is not proof that departure
never occurred or that the refresh lifecycle itself caused the timeout. The other
two executions passed in **20.927s** and **22.419s**.

The first queued-poll failure (**20.279s**) occurred **before departure**, at a
5s assertion expecting the first task-list request. Task-view modules took about
**7.4s**; the first list request did eventually finish. Clicking an async app did
not establish its list/query readiness. A request counter alone also did not
confirm the response body or filter controls. The later diagnostic count was
already 7 because the old mock changed count on its second ordinary summary
poll, even before the tested departure. Consequently the old queued case's
positive count assertion could be satisfied by ordinary pre-boundary polling;
its single reported pass cannot independently certify post-boundary recovery.

The second queued-poll failure (**21.739s**) reached departure and failed the
original **5s** visible-count assertion. Task-list and summary HTTP requests
finished about **3.022s after departure**, followed by further polls. No native
JavaScript error or request failure was recorded. The final surface probe itself
exceeded its 3s limit. HTTP completion does not prove the client consumed the
JSON, published query data or changed the DOM in time. These observations do
not identify the Windows data/render/automation cause; this remains a real
unresolved acceptance failure. The third execution passed in **19.205s**, with
the earlier fixture limitation retained.

## Changes for the next run

Only browser-test prerequisites, passive diagnostics and runner/documentation
descriptions change. Product runtime and the frozen UI remain unchanged.

Both task-window cases activate the **visible, enabled application with native
Enter** through its normal UI path, then confirm the actual first task-list
HTTP 200 response and successful body completion. They require a visible,
enabled status filter before departure. New static phases separate activation,
first list response and filter readiness. Cold session/workspace/module setup
still consumes the **original 45s case budget**. There is no prewarming, internal
store navigation, forced/dispatched click, sleep, retry, skip or increased limit.
These setup actions do not certify pointer-stability or cold startup within 5s.

The queued fixture retains count 0 until the browser **actually dispatches**
`beforeunload`, observed by a test-only binding. It does not change count before
an automation round trip to that event. A visible baseline 0 is required after
the real task list is ready; subsequent list reads must exceed the captured
baseline, rather than assume only one ordinary poll has happened. Actual queued
QueryObserver timer callbacks still run in the same departure event turn, with
the microtask fetch prohibition, surviving-document count 7, original **5s**
recovery assertion, no writes and no page errors. The no-frame/cache-retention
case, held-second-read cancellation, immediate real refresh and terminal/log
worker/storage/ACK/input/identity checks remain.

Passive failure diagnostics now record browser timestamps when the application
consumes its own summary JSON and when the taskbar's numeric accessible count
changes. A separate browser timestamp identifies departure. They do not read
a response clone, set query data, modify the UI or await the reporting binding.
Numbers, per-document read indices and fixed lifecycle kinds are whitelisted;
response bodies, labels, user field contents and storage dumps are excluded.
Signals remain capped at 120. DOM publication is **not proof of painted pixels**;
the timestamps distinguish data consumption/DOM changes from delivery of their
diagnostic messages and a later unavailable surface probe.

The [Windows runner](../../operations/WINDOWS_VALIDATION.md) still requires a
fresh production build and the **same five exact titles, each three times:
strictly 15/15**, without importing earlier passes. Real Windows effectiveness
must be established by the next pinned report.

## Local validation

Evidence: `.local/evidence/windows-report12-2026-10-04/`. Final browser checks use
the official Playwright `v1.63.0-noble` image, one worker, independent fresh
compiled preview servers, and run engines sequentially. The affected two files
have seven scenarios, including the unselected automatic/main terminal variants.
All final executions have no retries, skips or flaky outcomes.

| Local check | Actual result |
|---|---|
| Final compiled WebKit, seven scenarios each three times | **21/21**, **114.264s** |
| Final compiled Firefox, same seven scenarios | **7/7**, **42.794s** |
| Final compiled Chromium, same seven scenarios | **7/7**, **41.167s** |
| Controlled WebKit cold task-module gate, real queued-poll case | **1/1**, **9.729s**, details below |
| Type check and fresh test-mode production build | PASS; **3,740 modules**, Vite **5.70s** |
| Final type recheck, including browser tests | PASS |
| Actual PowerShell 7.4 Linux-container runner harness | PASS |

The controlled cold-view check gates the actual production task module until
the **third real ordinary summary poll**, with no sleeps or query/DOM injection.
It releases at **6.453s**, confirms the real task-list response at **6.538s**,
and starts departure at **6.590s** with count still 0. The original queued-poll
assertions finish at **6.748s**; the case remains inside 45s and its original 5s
recovery limit. This validates slow-view prerequisites and the corrected fixture
transition, not the unresolved Windows post-boundary cause. The initial WebKit
run before the final browser-observed boundary binding also passed **21/21**,
**113.237s**, and remains separate preliminary evidence.

The runner harness exercises real AST-extracted functions and report finalization:
progress-host/child faults, literal arguments, failures/skips/timeouts, title/
execution/build gates, numeric body/DOM timestamp copying, arbitrary attachment
exclusion, ZIP boundaries and interrupted-run recovery. It is Linux PowerShell
evidence, not a new Windows result.

An independent intentional browser failure verifies actual summary-body
consumption, DOM counts 0/7 and browser departure timestamps in a **2,823-byte**
diagnostic. The real runner copies it correctly and excludes extra arbitrary
fields and unrecognized event kinds. Its JSON is **0 passed / 1 failed**, total
**7.096s**, exit 1 at the deliberate assertion, and remains separate from product
passes. The copied evidence is in `probe-diagnostic-check/`; no trace, body,
field values or storage contents are published.

Actual AST-extracted runner functions also verify all three local JSON reports'
exact required-title execution counts, reject the returned Windows **12/15** and
both incomplete titles, and copy all three returned diagnostics (including the
unavailable third surface). Actual CLI `--list` confirms **five exact titles ×
three = 15 tests in two files**; that is selection verification, not an additional
executed pass. Evidence includes `webkit-final.json`, `firefox.json`,
`chromium.json`, `delayed-task.json`, `diagnostic-probe.json`,
`webkit-selection.json`, and `returned-diagnostic-check/`.

Reproduction uses `web/playwright.config.ts`,
`BLORA_E2E_FRESH_SERVER=1`, `BLORA_E2E_PREVIEW=1`, `BLORA_BROWSER=<engine>`,
and `node node_modules/@playwright/test/cli.js test
tests/browser/terminal.spec.ts tests/browser/query-navigation.spec.ts
--workers=1 --reporter=line,json` (WebKit adds `--repeat-each=3`). Build command:
`node node_modules/vue-tsc/bin/vue-tsc.js --noEmit` followed by
`node node_modules/vite/bin/vite.js build --mode test --outDir .local/browser-test-build`.
Runner harness: `pwsh -NoLogo -NoProfile -File scripts/test-windows-validation.ps1`
in `mcr.microsoft.com/powershell:7.4-ubuntu-22.04`.

All owned browser/container sessions have ended; there are no running handles
to take over. The existing large-chunk warning remains. No production runtime,
dependencies, Go or SDK code changed; no new Go/native/unit result is claimed.
Final diff checks passed; **1,137 local Markdown links** were checked with no
missing targets. The pinned runner is published with these test sources, not
downloaded independently from a mutable branch.

Local browser transports are test doubles, not Windows backend/native acceptance.
E08, physical power/cache loss, authorized
Engine/SCM/firewall operations, native notifications/IME and cross-host failures
remain outside this narrow selection.
