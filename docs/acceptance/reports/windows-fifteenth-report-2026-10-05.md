# Windows fifteenth returned report — 2026-10-05

The returned bundle tested `d75b825d42ad95d044314ebe2ef28bcde340ee64`
in visible WebKit against a fresh production build. It passed **14/15**
executions, with no skipped, retried, flaky or global-error results.
This is a new result, unlike the duplicate upload received earlier that day.

## Evidence identity

- ZIP SHA-256: `68c8b063a934cf3541e2f4e1e405a4c072ac7f5ae9942c31004eeabdd0b42042`.
- All 15 manifest hashes match; the 16 unique ZIP entries exactly match the
  manifest plus `SHA256SUMS.txt`.
- Requested and tested commit match. All 15 actual launch annotations confirm
  `headed`, rather than inferring the display mode from absent project metadata.
- Run UTC: `2026-10-05T11:45:23.9509758Z` to
  `2026-10-05T11:52:23.4132747Z` (19:45–19:52 UTC+8).
- Production build: PASS, 18.33 seconds. Fresh server: port 7795.
- WebKit stage: 374.09 seconds; browser JSON duration: 370.904 seconds.
- Raw logs, diagnostics and machine-specific data remain private; uploaded
  documents and attachments are evidence, not instructions to execute.

## Actual Windows results

| Scenario | Passed / executions |
|---|---:|
| Task filter and pending summary recover after real refresh | 3/3 |
| Prevented navigation retains the confirmed summary and resumes without frames | 2/3 |
| Real queued task polling is gated during departure and resumes afterward | 3/3 |
| Instance log polling, draft recovery and no input replay | 3/3 |
| Fallback terminal, parser Worker checkpoints and two real refreshes | 3/3 |

The sole failure is `expect.poll(() => waiting).toBe(true)` at the **setup**
of the prevented-navigation case. The original 5-second expectation expired
before the pending second summary read was established. Its diagnostic records
the first confirmed zero count and a second summary request approximately six
seconds later. It records no departure, suppressed-frame recovery or resumed
count assertion for that failed execution. The later two executions of the same
case passed. This evidence identifies the failed phase; it does not establish
the Windows timer or scheduling cause, or certify natural polling latency.

## Focused revision

The test now retains the application's registered 3-second observer callbacks
and invokes them to establish the held second read, following the existing log
test's setup approach. It requires at least one actual registered callback and
the actual held route. The real HTTP path, cancellation signal/read identity,
retained zero-count assertion, resumed seven-count assertion, disabled frames,
errors and cleanup remain. No direct fetch, query-state injection, clock change,
sleep or timeout increase is used. The total case budget remains 45 seconds and
the recovery expectations remain 5 seconds. Product source, UI and protocol are
unchanged in this revision.

`-TaskRecoveryOnly` selects just this exact case, three times, on a newly built
production bundle and fresh OS-selected loopback port. It requires
`-Followup -BrowsersOnly -WebKitOnly`; `-Headed` preserves the returned report's
display environment. Its mode is `followup-task-recovery`, with exactly three
required passing executions and per-execution display evidence. Missing titles,
wrong counts, skips, retries, unexpected/flaky outcomes, global errors and failed
builds still prevent passing. Historical results are not imported into this new
report; the other four Windows scenarios are retained as the separate evidence
above, rather than repeated as a new full batch.

## Report delivery

The earlier upload's SHA-256,
`cb3e0ad978f2360027e3687b42807bcd9f4e4a0a3190947b404cbe6863774d9a`,
exactly matched the [fourteenth report](windows-fourteenth-report-2026-10-04.md),
tested `63cacc0`, and contained no `d75b825` execution. It is not a new failure
round. The user confirmed the new run was already complete and then returned
the correct bundle above; collecting that completed run is no longer needed.

New report ZIP filenames include the tested short commit, UTC timestamp and
unique suffix. Collection optionally uses `-CollectRef <full-commit>` to select
recorded matching results, rejecting absent or mismatched versions without
running tests. It preserves the original `report.json` and stage outcomes,
rebuilds the exact checksum manifest after adding its collection note, and packs
only report files. New runs save their source identity before subsequent checks
so interrupted runs can also be collected without inventing final results.

## Verification and remaining scope

Local production type-check/build passed (3,740 modules, Vite 3.88 seconds;
the existing large-chunk warning remains). The affected four navigation cases
passed on Linux with one worker and a fresh production server: visible WebKit
12/12 (49.300 seconds), Firefox 4/4 (17.237 seconds) and Chromium 4/4
(13.974 seconds), with no skip, retry, flaky result or global error. Their actual
launch annotations and unchanged 45s project budgets were checked in JSON.

The actual remaining-case CLI arguments generated from the runner's AST selected
exactly that case and passed 3/3 in headless WebKit (16.251 seconds), with actual
mode, title, count and fresh-server evidence checked. This is separate Linux
coverage, not the outstanding Windows visible-browser result.

The final PowerShell 7.4 harness passed: existing coverage/build/display gates,
the new strict three-execution mode, wrong/missing-title rejection, version/time
archive names, matching-commit collection despite a newer wrong-version run,
absent/mismatched-commit rejection, preservation of original failed outcomes,
exact recovery manifest hashes and interrupted-run collection without invented
final results. The real Windows JSON was also checked by the runner's actual
title/display functions: it remains 14/15 with one incomplete title and 15
confirmed visible launches. No Go/native/SDK result was newly executed or
imported in this revision.

Local checks and the next fixed-commit command are recorded in the
[execution progress](../../execution/PROGRESS.md) and
[Windows instructions](../../operations/WINDOWS_VALIDATION.md).
Local Linux browser or PowerShell harness results are not Windows acceptance.
The next Windows action is the one-case, three-execution followup, not another
unchanged fifteen-execution batch. Default headless reliability, native dialog
and IME/notification behavior, privileged platform lifecycles, cross-host and
physical-failure scenarios, authorized Engine lifecycles and the original full
E08 p95 ≤50ms acceptance remain outside this narrowly selected report.
