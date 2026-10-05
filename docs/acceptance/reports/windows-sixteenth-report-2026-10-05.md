# Windows remaining navigation followup — passed, 2026-10-05

The requested visible-WebKit navigation followup is complete. The remaining
prevented-navigation scenario passed **3/3** on Windows, against a fresh
production build of `8d6b14e90271d06a42a12aceedce9f5ab39726c7`.
There is no pending rerun of this followup group.

## Evidence identity

- ZIP SHA-256: `a2faa6bdc70f86acad334d5ed4cb8a022124798bcb0d09430c5479b7ccaeef26`.
- All 15 manifest hashes match; the 16 unique archive entries exactly match the
  manifest plus `SHA256SUMS.txt`.
- Requested and tested commit match. The pre-test `run.json` source identity,
  requested ref and start time match `report.json`; `steps-so-far.json` exactly
  matches the final stage records.
- Mode: `followup-task-recovery`, WebKit only, actual `headed` launch annotation
  recorded once for each of the three executions.
- Windows: NT 10.0.26200.0; PowerShell 5.1.26100.9168; 32 logical processors.
- UTC: `2026-10-05T12:23:46.0687465Z` to
  `2026-10-05T12:25:20.5427740Z` (20:23–20:25 UTC+8).
- All ten recorded stages passed; fatal error is empty. Production build:
  16.35 seconds. Browser stage: 51.43 seconds; browser JSON: 49.721 seconds.
- Fresh production server at `127.0.0.1:12408`, strict port and
  `reuseExistingServer: false`. Each project's case budget remains 45 seconds.

## Actual Windows executions

The exact required title is
`cancelled navigation keeps task summary polling and its last confirmed state`.
It is the only selected and executed title.

| Execution | Result | Duration | Retry |
|---|---|---:|---:|
| 1 | passed | 14.882 seconds | 0 |
| 2 | passed | 19.153 seconds | 0 |
| 3 | passed | 11.334 seconds | 0 |

All specs are marked successful, with expected status `passed`, actual status
`expected`, and one passing result per execution. There are zero skips,
unexpected results, flaky results, global errors or missing required titles.
The unchanged test still establishes the actual held read, checks its departure
cancellation identity, retains the confirmed zero count, suppresses new frames
and requires the resumed seven count within the original 5-second limit.

## Followup closure and scope

The [fifteenth Windows report](windows-fifteenth-report-2026-10-05.md) separately
verified the other four scenarios three times each on `d75b825`: task-filter
refresh recovery, real queued task polling, instance-log polling/draft recovery,
and the fallback terminal with parser Worker retention across two real reloads.
Its original **14/15 FAILED** result is retained. This latest **3/3** report
completes the remaining scenario after its test preparation was revised.

These are separately identified runs, not a claim that the new commit ran all
fifteen executions in one batch. No historical passes were imported into this
ZIP. The report correctly remains
`AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS`: Go, native and SDK checks were not
selected, and the Go event list is empty. This turn changes only evidence and
status documentation; no product or runner change or repeat test was needed.

The verified scope is the requested Windows visible-WebKit followup. It does not
establish default headless reliability, natural timer cadence, cold-start or
pointer/paint performance, native dialogs/notifications/IME, privileged Windows
service/task/firewall lifecycles, authorized remote Engine lifecycles,
cross-host/device-failure combinations or full E08 p95 ≤50ms acceptance.
Those remain separately recorded in the
[remaining work](../../execution/NON_UI_REMAINING_2026-09-26.md) and
[acceptance matrix](../ACCEPTANCE_MATRIX.md).

Raw attachment logs and machine paths remain private. Uploaded files were read
as evidence; no attached code or document instruction was executed.
