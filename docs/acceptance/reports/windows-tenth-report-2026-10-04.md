# Tenth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`054fd7092544e6d16da6af32bebb847909b02ce11df001eb9f1f35c56952eb66`.
All **17 entries** in `SHA256SUMS.txt` match. The ZIP has 18 entries including
the manifest. Tested checkout and requested ref:
**`3ef289aec026bb07db3d83bc546abffc927883cf`**,
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]`,
`browserServerMode=production-preview`, `browserBuildSelected=true`.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-04 11:13:10–11:22:09 UTC+8**
(03:13:10–03:22:09 UTC). Original outcome **FAILED**, no fatal runner error.
Private machine paths and raw uploaded logs are not published.

| Actual Windows selection | Passed | Failed / timed out | Skipped |
|---|---:|---:|---:|
| Pending task-summary cancellation and restored task filter | 3 | 0 | 0 |
| Continued task-summary polling after prevented navigation, with RAF suppressed | 2 | 1 | 0 |
| Real queued task polling during departure | 3 | 0 | 0 |
| Independent instance-log polling and retained command drafts | 1 | 2 | 0 |
| Terminal fallback renderer / worker checkpoints | 3 | 0 | 0 |
| Total | **12** | **3** | **0** |

The separate production build passed in **17.42s**. Browser JSON duration:
**484.807s**; browser stage: **486.63s**. No retries or flaky outcomes.
All three terminal executions passed (**31.485s**, **33.691s**, **34.322s**),
including the original refresh, worker/storage, parsed-byte ACK and no-input-replay
assertions. This is actual evidence for the tested revision, not proof that the
previous timeout's Windows cause was development compilation.

The runner correctly rejects both incomplete navigation/log titles and the
aggregate. Go/native/race, SDK, Chromium and Firefox were omitted, not newly
passed. Earlier evidence remains scoped to its tested revision. This report
does not supersede full Windows, native lifecycle or mixed-load E08 gaps.

## What the three failures establish

Both log cases exhausted the unchanged **45s** test limit (**45.182s** and
**45.160s**) at the actual **console button's pointer click**, waiting for
visibility, enabled state and stability. No instance-log HTTP request or
`beforeunload` had occurred. Their diagnostics show a visible, focused document,
with painting observed by the later surface probe. The test had not reached its
log cancellation or recovery assertions. Asset requests took roughly 3.2–3.3s;
periodic task queries continued, with one approximately 17s gap in each case.
This locates the stalled action but does not establish a GPU/driver, frame,
layout or network cause. A later painted probe cannot prove frames were running
during the stalled click.

The second prevented-navigation execution failed its original **5s** visible
task-count assertion, total **20.856s**. Its second summary request was cancelled
2ms after `beforeunload`. The third request finished about **3.019s** after that
event, and a fourth also finished. No page error, unhandled rejection or native
exception was recorded. A successful HTTP completion does not prove its body was
consumed, query data published or DOM assertion observed. The original diagnostic
did not record the displayed count or intermediate query phases. The Windows
response-publication cause therefore remains unresolved.

## Changes for the next check

Task queries cancelled at departure now **revalidate immediately when the safe
document-read gate resumes**, using active query observers and
`cancelRefetch:false`. They retain their confirmed cache while cancelled, do not
replace a queued read that already resumed, and do not activate disabled queries
or replay mutations/input. The existing two-frame / **250ms** bounded resume
guard remains; `pagehide` invalidates it and `pageshow` resumes a surviving
document. The subscription is removed with the application scope. This removes
the avoidable wait for another 3s poll, without claiming a confirmed explanation
for the uploaded count timeout.

A controlled real QueryClient/QueryObserver regression with animation callbacks
suppressed fails without the resume subscription (**1 failed**, old behavior
still has two reads at 250ms), and passes with it. It verifies the last confirmed
count, immediate new data, no disabled/unrelated query reads, no stale deadline
after `pagehide`, `pageshow` resumption and subscription disposal. The baseline
temporarily disables only the new subscription; it is not a rerun of the uploaded
Windows failure or evidence about the Windows rendering cause.

The log HTTP lifecycle case now opens the same real application, resource and
console through **native button Enter-key activation**, preceded by visible
control assertions. It still fills the actual command field, captures actual
registered interval callbacks, cancels the held log read, checks queued reads
with and without an old pending request, reloads the real page and verifies the
same draft/run, zero writes and zero page errors. There is no force click,
dispatched click, internal-store navigation, prewarming, sleep, retry or skip.
This tests keyboard setup and log lifecycle; it no longer claims pointer
stability coverage for those three setup actions. The terminal and other browser
cases keep their pointer actions. It does not resolve overall pointer/paint
performance acceptance.

Playwright documents [pointer actionability and stable-frame checks](https://playwright.dev/docs/actionability)
and [native keyboard actions](https://playwright.dev/docs/api/class-locator#locator-press).
These explain the distinct setup requirements; they do not establish the
machine-specific cause.

The retained-count test holds the first resumed mock response until its cache
assertion completes, then releases it before the original 5s new-count assertion.
This makes the mock's state transitions explicit, avoiding a race with faster
revalidation. All four task/log tests now emit static phase timing, alongside
terminal phases. Failed diagnostics also record only the numeric displayed task
count and console button geometry/selection/focus. No bodies, user field contents,
storage dumps or raw traces are bundled.

The next Windows selection remains five required titles, each three times,
**strictly 15/15**, served from a freshly type-checked production build. The **45s
case limit**, **5s recovery assertion**, account separation and original protocol,
ACK, identity, storage, error and no-replay criteria remain. Product visuals and
the UI design are unchanged.

## Local validation

Evidence: `.local/evidence/windows-report10-2026-10-04/`. All product suites below
have no skips, retries or flaky outcomes. Linux tests cannot establish Windows
resolution. Each engine uses one worker and a fresh compiled preview server in
the official Playwright `v1.63.0-noble` image. The Firefox run's beginning overlaps
the WebKit run's final approximately 21 seconds; these are correctness checks,
not isolated performance comparisons or E08 evidence.

| Check | Actual result |
|---|---|
| Resume-subscription regression without the new subscription | **0/1**, missing immediate third read at 250ms; baseline only, not Windows reproduction |
| Final web unit suite | **90/90**, 17 files, **1.960s**; includes actual QueryObserver resume/cache/hidden-document/disposal regression |
| Type check and production build | Passed; 3,740 modules, Vite **6.23s**; existing large-chunk warning retained |
| Compiled WebKit, twelve account/terminal/navigation/log cases each three times | **36/36**, **210.400s** |
| Compiled Firefox, same twelve cases | **12/12**, **82.874s** |
| Compiled Chromium, same twelve cases | **12/12**, **71.432s** |
| PowerShell 7.4 isolated runner harness | Passed; strict required titles/counts, missing/failed production-build gates, progress-host faults, diagnostic boundaries/retention, ZIP/privacy and recovery |
| Isolated intentionally failed diagnostic probe | **0/1**, **2.967s** deliberate assertion failure; real runner copies the **857-byte** JSON with two phases, displayed count 7 and selected console geometry |
| Actual local and uploaded JSON processed by real runner functions | Passed; every local required title count verified, uploaded 12/15 rejected with both incomplete titles, all three uploaded failure diagnostics decoded |
| Runner-generated selection checked by actual Playwright CLI `--list` | Exactly five titles each three times, **15 executions in two files**; selection only, not another executed suite |
| Documentation and diff checks | 1,129 local links checked, none missing; `git diff --check` passed |

The no-frame summary case's new-count phase follows release of its resumed
response by **330ms**, **340ms**, **337ms** across the three WebKit executions.
The entire log lifecycle cases take **3.203s**, **2.391s**, **2.931s** locally.
These timings include test transport doubles and do not establish a Windows or
backend latency target. Existing loopback proxy/teardown warnings in the standalone
log suites remain recorded; page-error and write/replay assertions stay active.

A separate exploratory probe paused the application clock to attempt a pointer
stability stall. Its actual click **succeeded**, so its predicted failure assertion
failed. It did **not** reproduce the Windows stall or validate a frame-cause claim.
That initial probe run is preserved separately (`probes.json`, two failures:
the incorrect hypothesis and the intentional diagnostic assertion). The bounded
diagnostic was then checked alone with the unchanged intentional assertion and
real runner decoder; neither probe is counted as a product-suite pass. Playwright
[clock control](https://playwright.dev/docs/clock) must not be assumed to suspend
the automation's native actionability checks.

All owned test containers and sessions have ended; no running handle needs
handover. No product visual or Go/native code changed. The next product validation
action is the pinned Windows followup, not another unchanged local full-suite run.

## Next Windows run and remaining boundaries

Use the pinned script and checkout provided in the conversation with
`-Followup -BrowsersOnly -WebKitOnly`. Return the newly generated ZIP.
The runner checks a fresh production build and every title/execution; omitted
engines and Go/SDK checks are not imported from older reports. Live phase output
and failure diagnostics will locate any further stalled action or count update.

Management HTTP/WebSocket endpoints remain transport doubles in these browser
checks. They use real DOM, xterm, worker, storage and binary processing, but do not
establish Windows native notification/IME, privileged SCM/Task Scheduler/firewall
rollback, authorized real Engine lifecycle, full Windows backend-to-node behavior,
power/device-cache loss, cross-host endurance or E08 <=50ms performance. Native
dialogs, bfcache and arbitrary slow navigation are not proven by controlled
prevented-navigation events.
