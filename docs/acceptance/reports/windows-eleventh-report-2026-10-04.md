# Eleventh returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`37539b0e931f2fec161bfc36789299f5cc520e10f6a66faddd25c8bac4482e65`.
All **16 entries** in `SHA256SUMS.txt` match; the ZIP has 17 entries including
the manifest. Tested checkout and requested ref:
**`74268013bd3e4a4893147c26f29e019612af2d3e`**.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-04 12:08:45–12:15:35 UTC+8** (04:08:45–04:15:35 UTC).
Schema 2, outcome **FAILED**, no fatal runner error;
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]`,
`browserServerMode=production-preview`, `browserBuildSelected=true`.
Private machine paths and uploaded raw logs are not published.

| Actual Windows selection | Passed | Failed / timed out | Skipped |
|---|---:|---:|---:|
| Pending summary cancellation and restored task filter | 3 | 0 | 0 |
| Confirmed count and polling after prevented navigation without RAF | 2 | 1 | 0 |
| Real queued task polling at departure | 3 | 0 | 0 |
| Instance-log polling, retained command draft and no input replay | 3 | 0 | 0 |
| Terminal fallback renderer / worker checkpoints | 2 | 1 | 0 |
| Total | **13** | **2** | **0** |

The production build passed in **17.67s**. Browser stage **363.86s**, JSON
duration **361.995s**. No retries, skips or flaky outcomes. All three log cases
passed (**17.140s**, **8.304s**, **8.388s**) using the native-key setup added in
the preceding revision. This is actual evidence for that setup and lifecycle,
not evidence that the original Windows pointer-stability cause was resolved.

The runner correctly rejects both incomplete titles and the aggregate. Go,
native/race, SDK, Chromium and Firefox were omitted, not newly passed. Prior
passes remain evidence for their respective tested revisions.

## What the two failures establish

The third prevented-navigation execution failed its **initial** `0 项后台任务`
visible assertion, before establishing the pending second request or suppressing
RAF/dispatching `beforeunload`. Total **16.236s**, assertion limit **5s**.
The only phase is opening the desktop. Its first summary request began about
**9.065s after navigation** and finished 11ms later; the later bounded diagnostic
shows confirmed count **0**. Static assets completed in 12–22ms; no page error,
request failure or departure event was recorded. This differs from the previous
report's post-navigation count assertion. It does not show a failure of resumed
polling or establish why cold initialization took that long.

The first terminal execution timed out after **45.152s**, before attachment or
any terminal protocol, refresh, worker, storage or ACK assertions. Phase times:
navigation **31ms**, opening instance center **6,423ms**, opening instance
**37,932ms**, opening console **44,378ms**. The stalled console action was waiting
for native pointer actionability/stability. Later diagnostics show a visible,
enabled console button, focused visible document and a painted probe, with no
page/network error. Those later observations do not prove paint was available
throughout the earlier click. The other two terminal executions passed in
**32.734s** and **31.809s**. No Windows GPU/layout/frame cause is established.

## Changes for the next run

Only browser-test setup and runner/documentation descriptions change. Product
code and frozen UI remain unchanged.

Task-case setup now registers a first-summary response wait **before** navigation,
then confirms status 200 and successful body completion before starting the
original 5s visible-count assertion. The session/workspace initialization remains
inside the **same 45s total test budget**. It uses real HTTP completion, with no
sleep, prewarming or injected state. The original 5s post-navigation recovery
assertion remains unchanged. This explicitly distinguishes cold bootstrap from
rendering confirmed data; it does **not** prove cold bootstrap completes in 5s.

Terminal setup now checks visible, enabled buttons and uses **native Enter-key
activation** for the real application, resource, console, terminal manager and
session attachment. It keeps the same UI and transport route, real xterm/worker/
IndexedDB, parsed-byte ACK accounting, input delivery/no replay, two refreshes,
session/cursor identity, view migration, output-gap retention and storage-failure
credit checks. Later takeover, migration and remount pointer actions remain.
These five setup actions no longer claim pointer-stability coverage. There is no
force click, dispatched click, internal-store navigation, retry, skip or enlarged
limit. Pointer/paint performance remains a separate unclosed acceptance gap.

The first local Firefox check also exposed a test-observation assumption: a slow
real refresh recorded **two** aborts, and its trace shows a distinct third read
before the new document's session request. This is consistent with another
read resumed by the bounded guard; the old recorder had no request identity.
Its aggregate abort-count assertion failed, without establishing duplicate
cancellation of the held read. The recorder now
assigns a per-document read index and resets its departure flag in each document.
It requires **exactly one departure cancellation of held read 2**, and checks
that every recorded abort occurs after departure. Extra distinct read cancellations
are separately identified, rather than conflated with read 2. The separate
queued-fetch prohibition, retained count/filter, no-write/replay and page-error
assertions remain. This does not make the bounded guard proof of arbitrary slow
navigation or native navigation cancellation.

The next [Windows runner](../../operations/WINDOWS_VALIDATION.md) still requires
a fresh type-checked production build and the **same five exact titles, each
three times: strictly 15/15**. Live phase and bounded failure diagnostics remain;
no earlier passes are imported and no user field/body/storage data is bundled.

## Local validation

Evidence: `.local/evidence/windows-report11-2026-10-04/`. Final browser checks
use the official Playwright `v1.63.0-noble` image, one worker, a fresh compiled
preview server, and run engines sequentially. They cover the **two affected
files' seven scenarios**. All final executions have no skips, retries or flaky
outcomes. Local results cannot establish Windows resolution or E08 performance.

| Check | Actual result |
|---|---|
| Final compiled WebKit, seven scenarios each three times | **21/21**, **182.344s** |
| Final compiled Firefox, same seven scenarios | **7/7**, **65.855s** |
| Final compiled Chromium, same seven scenarios | **7/7**, **49.663s** |
| Type check, including modified browser tests | Passed after the final recorder change |
| PowerShell 7.4 isolated real-runner harness | Passed; failure/skip/timeout, strict title/count and production-build gates, diagnostics, ZIP/privacy and old-run recovery |
| Real runner functions processing local and uploaded JSON | Passed; every required title count verified, uploaded **13/15** rejected with both incomplete titles, both bounded diagnostics actually decoded and copied |
| Runner-generated selection checked by actual Playwright CLI `--list` | Exactly **five titles each three times, 15 executions in two files**; selection only, not another executed suite |
| Documentation and diff checks | **1,133** relative local links checked, none missing; `git diff --check` passed |

The final no-frame count case's release-to-visible phase intervals are **272ms**,
**241ms**, **298ms** on local WebKit. Fallback-renderer/worker terminal cases take
**14.514s**, **13.496s**, **11.584s**. These include test transport doubles and
are not Windows/backend latency targets or pointer-performance evidence.

Initial WebKit **36/36**, **240.615s**, including account and standalone log
checks before the recorder refinement, is preserved in `webkit-initial.json`.
Initial Firefox **6/7**, **79.796s**, with the aggregate abort-count assertion
above, is preserved in `firefox-initial.json` with its original trace. It is not
counted as a passing suite. The final affected files were rerun after correction.
Existing loopback proxy/teardown warnings remain recorded; page-error and
write/replay assertions stay active.

Product source, dependencies and build configuration are unchanged; local
preview reuses the preceding compiled bundle without claiming a new build.
The Windows runner always builds a fresh bundle before testing. Go/SDK/unit
implementation is unchanged and not rerun or claimed as newly passed. All owned
containers and sessions have ended, with no handle requiring handover.

## Remaining boundaries

These browser suites replace management HTTP/WebSocket transports while using
real DOM, xterm, workers, storage and binary processing. They do not establish
Windows full-stack/native notification/IME, privileged SCM/Task Scheduler/firewall
rollback, authorized real Engine lifecycle, power/device-cache loss, cross-host
endurance or full mixed-load E08 <=50ms. Native dialogs, bfcache, arbitrary slow
navigation, cold startup and setup pointer/paint performance remain unproven.
