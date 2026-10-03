# Eighth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`92da4d5f6239347e11e8e13fdeb0ffec61de5c1e842a0a8eb2ea38d4f2eae467`.
All **15 entries** in `SHA256SUMS.txt` match; no listed file is missing.
The ZIP has 16 entries including the manifest. Tested checkout and requested ref:
**`57d4fb6e18977f78c92015a00df45cdf0c92e612`**,
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]`.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-03 20:59:26–21:05:00 UTC+8**
(12:59:26–13:05:00 UTC). Original outcome **FAILED**, no fatal runner error.
Private machine paths and raw uploaded logs are not published.

| Actual Windows selection | Passed | Failed | Skipped |
|---|---:|---:|---:|
| Pending task-summary cancellation and restored task filter | 3 | 0 | 0 |
| Continued task-summary polling after prevented navigation | 2 | 1 | 0 |
| Queued task polling during departure | 3 | 0 | 0 |
| Independent instance-log polling and retained command drafts | 2 | 1 | 0 |
| Terminal fallback renderer / worker checkpoints | 3 | 0 | 0 |
| Total | **13** | **2** | **0** |

Playwright JSON duration: **310.499s**; browser stage: **312.290s**.
No retries or flaky outcomes. The original terminal case now has three actual
Windows passes with its page-error, ACK, recovery and no-input-replay assertions
unchanged. The runner correctly rejects the two incomplete required titles and
the aggregate shortfall. Go/native/race, SDK, Chromium and Firefox were omitted;
empty arrays are not new passing results. Previous results remain scoped to
their tested revisions.

## Two distinct timeouts

The first task-summary execution failed its original **5s** visible-count
assertion (`query-navigation.spec.ts:82`) after **23.388s**. The confirmed count
was retained, and the held read was cancelled 2ms after `beforeunload`. No new
summary request followed in the diagnostic. Its surface probe did not finish
within 3s, so the upload alone cannot establish whether frames, JavaScript
execution or another browser subsystem stalled.

The first log-poll execution failed while waiting to establish its second,
held log read (`query-navigation.spec.ts:187`) after **44.405s**, within the
unchanged 45s test limit. It never reached its controlled navigation event.
The first log request started **36.371s** after initial navigation and completed
6ms later; the setup had already consumed most of the test budget. The visible,
focused 1440×960 page had no dialog and failed its paint probe. Neither failure
recorded a JavaScript/page error. These are not the former terminal fetch error,
nor proof that a particular Windows GPU or driver caused the failures.

## Restore reads independently of paint

A controlled browser case now suppresses new animation-frame callbacks exactly
at the prevented-navigation boundary. With the previous production code it
fails the same unchanged 5s count assertion: **0/1 passed, 14.543s**. Two
API tests independently fail on the previous code when no frame is delivered.
An initial unit-test setup also accidentally faked RAF; it was corrected to
fake only timeout functions before recording the meaningful failing baseline.

The shared safe-read gate retains early two-frame resumption and adds a
**250ms timer fallback** for a surviving document. `pagehide` cancels the timer
and invalidates both timer and frame callbacks until `pageshow`. Another
`beforeunload` starts a new generation and deadline, so an older callback cannot
release a newer pause. Tests check pending and queued read resumption without
frames, no duplicate reads after later paint, stale deadlines, and continued
pagehide blocking after both frames and the fallback deadline have elapsed.

Only document-navigation cancellation may restart a safe read. Owner/account
cancellation, response-body cancellation and real network failures retain their
existing behavior; ordinary mutations, saves, terminal input and backend tasks
are never automatically replayed. No visual or theme change was made.

The log case establishes its held read by invoking the real registered
application interval callbacks after normal UI setup. It no longer needs to
wait for the next wall-clock tick. Its pending-read cancellation, repeated
queued-callback boundary, restored run, draft recovery, zero mutation and empty
page-error assertions remain active. Queued-read observations now cover the
navigation event turn and its microtasks, ending at the next task independently
of RAF; waiting for a paint to end that observation would falsely label a
correct no-frame recovery as departure-time traffic.

No force click, settling sleep, retry, skip, error filter or extended deadline
was added. The strengthened no-frame browser assertion still requires recovery
within its original 5s. A 250ms deadline is a bounded guard, not an authoritative
native-navigation cancellation signal: a busy or suspended JavaScript loop can
delay it, and an unusually slow accepted navigation can outlive it. The
controlled cases do not prove every native confirmation-dialog or bfcache
sequence. Relevant browser contract:
[animation frames may pause independently of timers](https://developer.mozilla.org/en-US/docs/Web/API/Window/requestAnimationFrame).

## Local validation

Evidence: `.local/evidence/windows-report8-2026-10-03/`. Local checks run on
Linux and do not establish that the Windows timeouts are resolved.

| Check | Result |
|---|---|
| Controlled no-frame task-summary baseline | 0/1 passed, 14.543s; original 5s assertion fails |
| Corrected API-unit baseline before production fix | 11 passed / 2 failed; both no-frame cases time out at the unchanged 5s unit limit |
| Frontend units after fix | 89/89 in 16 files, 1.320s elapsed |
| Type check and production build | Passed; 3,740 modules, Vite 3.22s; pre-existing large-chunk warning retained |
| WebKit: accounts, all terminal configurations, navigation and log suites, each three times | 36/36 passed, 226.946s; no skip/retry/flaky |
| Firefox: same twelve affected scenarios once each | 12/12 passed, 87.737s; no skip/retry/flaky |
| Chromium: affected scenarios plus Docker and files | 19/19 passed, 103.257s; no skip/retry/flaky |
| Additional Chromium batch, editor-save and system scenarios | Separate 4/4 pass, 33.451s; no skip/retry/flaky; initial selection had three unmatched filename patterns, so only these omissions were rerun |
| PowerShell 7.4 isolated runner harness | Passed; failure/timeout cases intentionally remain failures, progress host faults, coverage gates, ZIP/privacy and recovery checked |
| Runner functions with actual local and returned Windows JSON | Passed; exact ten-title local subsets, five-title WebKit subset each three times, returned 13/15 rejected and both returned/baseline diagnostics decoded |
| Runner-generated selection with the actual Playwright CLI `--list` | Exactly five required titles each three times: 15 executions in two files; selection only, not another executed suite |

Commands use the official Playwright `v1.63.0-noble` container, a fresh Vite
server (`BLORA_E2E_FRESH_SERVER=1`), the selected engine and one worker:

```text
node node_modules/@playwright/test/cli.js test tests/browser/accounts.spec.ts tests/browser/terminal.spec.ts tests/browser/query-navigation.spec.ts tests/browser/logs.spec.ts --workers=1 --reporter=line,json
# WebKit additionally uses --repeat-each=3.
# The 19-case Chromium selection additionally matched docker.spec.ts and files.spec.ts.
node node_modules/@playwright/test/cli.js test tests/browser/instance-batch.spec.ts tests/browser/editor-save.spec.ts tests/browser/system-management.spec.ts --workers=1 --reporter=line,json
npm test -- --reporter=json --outputFile=<evidence>/units-after.json
npm run build
pwsh -NoProfile -File scripts/test-windows-validation.ps1
```

Vite's absent-backend/teardown proxy warnings remain visible in the mocked
browser suites; those suites do not claim full backend/platform acceptance.
All owned test sessions and containers have ended. No production service or
UI/theme setting was changed.

## Next Windows run

Use the same newly pinned source and runner revision with
**`-Followup -BrowsersOnly -WebKitOnly`**. The existing five-title selection is
retained: both corrected polling cases plus queued reads, real refresh recovery
and the original fallback/worker terminal guard. Each must pass three times,
requiring **15/15**, with no imported old pass, skip or retry. The runner checks
each exact title as well as the aggregate. Go/GCC/SDK and the other engines are
deliberately not rerun. Progress and failure-safe report ZIP creation remain
enabled; return the ZIP even on failure.

The broader selection remains ten cases once on Chromium/Firefox and three
times on WebKit. Windows correction efficacy, full native-platform coverage,
strict E08 latency and other external-environment gaps remain open. See the
[runner instructions](../../operations/WINDOWS_VALIDATION.md) and
[remaining-work ledger](../../execution/NON_UI_REMAINING_2026-09-26.md).
