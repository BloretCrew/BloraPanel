# Thirteenth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`38cffea1df54ffad4d62f01080b285288d14b7ca679b639dd60c13632c012648`.
All **14 entries** in `SHA256SUMS.txt` match the complete, unique ZIP entry set;
the ZIP contains 15 entries including the manifest. Requested and tested ref:
**`c63bcb49a0af67aeead4e27bf070708d8b48b1e4`**.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-04 21:36:51–21:38:00 UTC+8** (13:36:51–13:38:00 UTC).
Schema 2, outcome **FAILED**, no fatal runner error;
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]`,
`browserServerMode=production-preview`, `browserBuildSelected=true`.
Private machine paths and uploaded raw logs are not published.

| Actual Windows selection | Result |
|---|---|
| Type-checked production build | PASS, **24.04s** |
| Matching WebKit installation | PASS, **2.18s**, cached engine |
| Browser stage | Startup failure, **2.39s** |
| Five scenarios, three executions each | **None executed**; all five titles lack PASS evidence |

Playwright reports **zero suites and zero executions**, with one global error:
its fresh server cannot start because `http://127.0.0.1:5173` is already used.
The JSON duration is **19.980ms**, not seconds. Aggregate expected/unexpected/
skipped/flaky counts are all zero because no test started; they do not mean the
15 required executions passed, failed their assertions, or were deliberately
skipped. No case diagnostic was produced. The runner correctly rejected the
missing title coverage and 15-execution requirement.

The port occupant's identity is unknown. The evidence does not establish that
it is an earlier validation run. This startup failure prevents assessment of
the [twelfth report's](windows-twelfth-report-2026-10-04.md) remaining task-window
failures and subsequent test-prerequisite/diagnostic corrections. Their Windows
effectiveness remains unverified. Go/native/race, SDK, Chromium and Firefox were
omitted, not newly passed. Prior evidence stays attached to its tested revision.

## Changes for the next run

The Windows runner now asks the OS for a currently free **loopback port**
immediately before each selected browser engine. It releases its temporary
listener before starting Vite, sets `BLORA_E2E_PORT`, and records the selected
port in console output and `report.json`. Both Playwright's base URL and its
development/compiled-preview server use that same port. Ordinary manual commands
still default to 5173; explicitly invalid port values fail before testing.

Fresh-server checks remain enabled in validation runs. Both server commands
use `--strictPort`; the runner does not terminate or reuse the program on 5173
and does not silently fall back to a different server. If another listener takes
the selected port between probing and startup, validation **fails explicitly**.
Port selection minimizes the collision but is not an atomic reservation lasting
through Vite startup.

Report finalization also retains the actual server URL and up to five global
Playwright error messages, each at most 2,000 characters after the existing
redaction step. Such errors mark browser evidence incomplete even if nominal
aggregate/title pass data is present. An actually failed command remains FAILED.
The report README includes these startup errors so an empty test report is
readable without digging into the raw JSON.

Only the test runner, browser-server configuration and verification/documentation
change. Product runtime, frozen UI, dependencies, browser cases and passive
diagnostics are unchanged. The [pinned followup](../../operations/WINDOWS_VALIDATION.md)
still requires a successful fresh production build and **the same five exact
titles, each three times: strictly 15/15**. Original **45s total** and **5s recovery**
limits, real queued timers, cancellation identity, retained cache, refresh,
worker/storage, ACK, no-write/input-replay and page-error assertions remain.

## Local validation

Private evidence is under `.local/evidence/windows-report13-2026-10-04/`.
The actual PowerShell 7.4 Linux-container function harness passed, including a
real listener occupying 5173, OS-selected free-port allocation, probe release,
bounded startup-error/URL/port reporting and all existing progress, argument,
failure/skip/timeout, title/execution/build, diagnostic, ZIP and recovery gates.
This is not a new Windows PowerShell 5.1 execution result.

The actual runner functions selected port **37557** while 5173 was occupied.
That selected number was then used by browser tests in their own isolated
container network namespaces, with an independently owned listener on 5173.
The official Playwright `v1.63.0-noble` image runs engines sequentially, one
worker per run. No product source, dependency or bundle content changed; compiled
checks reuse the preceding locally built bundle and do not claim a new local
production build. Windows's uploaded fresh build is a separate PASS above.

| Local check | Actual result |
|---|---|
| Occupied default-port startup | Reproduced before any test; zero suites, one global error |
| Selected port claimed before startup | Rejected before any test; no server reuse |
| Six invalid port inputs | Actual configuration load rejects all six |
| Compiled WebKit, exact Windows selection | **15/15**, **97.036s** |
| Compiled Firefox, all seven cases in the two files | **7/7**, **66.612s** |
| Development server, Chromium, same seven cases | **7/7**, **78.294s** |
| Type check, including changed browser configuration | PASS |
| Actual PowerShell 7.4 runner harness and OS-port selection | PASS |
| Actual local/returned JSON parsed by runner functions and returned report finalized | PASS; empty Windows execution rejected and startup error retained |

The two intentional startup failures remain separate evidence, not product
passes. For each browser, 5173 stays occupied throughout its positive executions;
the Playwright result records the selected URL and `reuseExistingServer=false`,
and the server command contains the exact selected port plus `--strictPort`.
The owned default-port fixture receives no application-test requests and is
still running after the tests. Only the fixture's own listeners are closed.
There are no retries, skips, flaky outcomes or global errors in these positive
runs. Existing loopback proxy errors during fixture teardown remain in private
logs and are not hidden or interpreted as Windows backend evidence.

Actual AST-extracted runner functions verify each positive JSON report's exact
required-title execution counts and server metadata. They reject the returned
Windows zero-execution report with all five titles still missing and no invented
case diagnostic. Executing the real report-finalization block with that returned
JSON preserves **FAILED**, the original startup error and server URL, and the
omitted native/Go/SDK scope; it creates a separate private evidence archive.
The synthetic harness also verifies that a global startup error prevents an
otherwise nominal 15-pass report from being accepted.

All owned test sessions and containers have ended, with no handles to take over.
Final diff checks passed; **1,145 local Markdown links** have no missing targets.
The remote branch was checked against the uploaded revision before publishing
the runner with these configuration/test/documentation sources in one commit.
No Windows effectiveness, Go/native/unit/SDK or E08 pass is inferred from local
browser results.

Reproduction uses `web/playwright.config.ts`, `BLORA_E2E_PORT=<free loopback port>`,
`BLORA_E2E_FRESH_SERVER=1`, `BLORA_BROWSER=<engine>` and
`node node_modules/@playwright/test/cli.js test
tests/browser/terminal.spec.ts tests/browser/query-navigation.spec.ts
--workers=1 --reporter=line,json`. Compiled checks set `BLORA_E2E_PREVIEW=1`.
WebKit adds the actual runner's anchored five-title pattern and `--repeat-each=3`;
Firefox/Chromium cover all seven cases in the two affected files. Keep a separate
owned listener on 5173 to reproduce this regression environment. Harness:
`pwsh -NoLogo -NoProfile -File scripts/test-windows-validation.ps1` in
`mcr.microsoft.com/powershell:7.4-ubuntu-22.04`.

Browser management transports are test doubles. They are not Windows backend/
native or hardware-compositor acceptance. Full Windows notification/IME,
authorized Engine/SCM/task/firewall lifecycle, physical power/cache loss,
cross-host combinations and the unchanged E08 p95 ≤50ms target remain open.
