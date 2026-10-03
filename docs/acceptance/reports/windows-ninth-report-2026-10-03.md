# Ninth returned Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`58fd4638977129a2f428bf143980ffb8afec156e594d40b0a26a9a009f2ad237`.
All **14 entries** in `SHA256SUMS.txt` match; no listed file is missing.
The ZIP has 15 entries including the manifest. Tested checkout and requested ref:
**`3e6c31c75227b182a76945b02f2591c5dcdd705e`**,
`runMode=followup-webkit`, `selectedBrowsers=["webkit"]`.
Windows NT 10.0.26200 AMD64, 32 logical processors, PowerShell 5.1.26100.9168.
Run: **2026-10-03 21:42:58–21:48:45 UTC+8**
(13:42:58–13:48:45 UTC). Original outcome **FAILED**, no fatal runner error.
Private machine paths and raw uploaded logs are not published.

| Actual Windows selection | Passed | Timed out | Skipped |
|---|---:|---:|---:|
| Pending task-summary cancellation and restored task filter | 3 | 0 | 0 |
| Continued task-summary polling after prevented navigation, with RAF suppressed | 3 | 0 | 0 |
| Real queued task polling during departure | 3 | 0 | 0 |
| Independent instance-log polling and retained command drafts | 3 | 0 | 0 |
| Terminal fallback renderer / worker checkpoints | 2 | 1 | 0 |
| Total | **14** | **1** | **0** |

Playwright JSON duration: **322.971s**; browser stage: **324.750s**.
No retries or flaky outcomes. Both eighth-report failing scenarios now have
three actual Windows passes at this revision, including the stronger no-frame
recovery assertion. The report still correctly rejects the incomplete terminal
title and aggregate. Go/native/race, SDK, Chromium and Firefox were omitted;
empty arrays are not new passing results. Earlier evidence remains scoped to
its tested revision.

## Remaining terminal timeout

The first fallback-renderer/worker execution exhausted its unchanged **45s**
test limit (**45.200s** including teardown). The later two full executions took
**18.810s** and **18.762s**, both passed. The failure recorded only the total
test timeout, without a failing assertion location or terminal phase timings.
There was no recorded native JavaScript error, unhandled rejection, Playwright
page error or failed request. The surface probe was visible, focused and painted,
with no dialog.

Relative to the diagnostic's first navigation, the initial `pageshow` was at
**4.326s**, the first log read at **30.986s**, and the first `beforeunload` at
**32.102s**. `pagehide` followed 9ms later, and the refreshed page showed at
**32.440s**. This establishes that most of the total case budget was consumed
before the later scrollback, window move, burst, second refresh and storage
failure checks could finish. It does not locate the specific slow UI action or
prove a Windows GPU/driver, network, timer or compiler cause. Development-server
lazy transforms and dependency optimization are a hypothesis to investigate,
not a confirmed explanation of the uploaded timeout.

## Test the compiled frontend and retain stage timing

The narrow Windows followup now type-checks and builds the production frontend
once in a separate **`web-browser-build`** stage, before starting its browser
suite. Playwright serves this freshly compiled bundle on loopback using
`BLORA_E2E_PREVIEW=1`. This removes on-demand development compilation from the
runtime case budget; it does not warm up a terminal, pre-open a resource or skip
the first UI navigation. Each case still starts with a fresh browser context
and carries out its original real UI actions and refreshes.

The build stage must pass; a missing or failed build cannot become successful
browser evidence, even if aggregate/title data claims passes. The checkout and
server are fresh. Port 5173 must be available; preview cannot silently switch
to another port or use a pre-existing server. The report explicitly records
`browserServerMode=production-preview` and `browserBuildSelected=true`.
Other runner modes retain development serving for suites that import `/src`.

Vite documents [previewing a compiled build locally](https://vite.dev/guide/static-deploy#testing-the-app-locally)
and [the strict-port and inherited proxy options](https://vite.dev/config/preview-options).
The preview server is local test infrastructure, not a production deployment.

Terminal checkpoints now log static phase labels and elapsed milliseconds as
the actual case progresses: desktop and resource navigation, replay ACK,
input, both refreshes, takeover, large scrollback, original-view migration,
burst/resize, output gap, and simultaneous storage failures. A failed case's
existing bounded diagnostic JSON includes the phases and twenty slowest static
asset request durations, along with server mode and existing lifecycle/error
signals. No request bodies, field contents, terminal input text, storage dumps
or raw traces are added to the diagnostic bundle.

The **45s test limit** and **5s recovery assertions** stay unchanged. Real
xterm, worker/main checkpoints, IndexedDB, binary protocol, parsed-byte credit,
view identity, exact input, zero automatic replay and empty page-error assertions
remain active. There is no force click, settling sleep, retry, skip or exception
filter. No product UI, theme, material or production behavior was changed.

## Local validation

Evidence: `.local/evidence/windows-report9-2026-10-03/`. Linux results cannot
establish resolution on Windows. Browser engines ran sequentially using the
official Playwright `v1.63.0-noble` image, one worker and a fresh local server.

| Check | Actual result |
|---|---|
| Fresh development-server fallback/worker terminal baseline | 1/1 passed; case **21.625s**, full run 25.764s; final phase 20.906s after fixture start |
| Same terminal case against a fresh compiled preview, without preceding UI/terminal cases | 1/1 passed; case **14.766s**, full run 17.023s; final phase 14.133s after fixture start |
| Type check and production build | Passed; 3,740 modules, Vite **3.10s**; pre-existing large-chunk warning retained |
| Compiled WebKit: all twelve account/terminal/navigation/log scenarios, three executions each | **36/36** passed, **181.604s**, no skip/retry/flaky |
| Compiled Firefox: same twelve scenarios | **12/12** passed, **67.804s**, no skip/retry/flaky |
| Compiled Chromium: same twelve scenarios | **12/12** passed, **55.533s**, no skip/retry/flaky |
| Intentionally failed browser diagnostic probe | **0/1**, **6.369s**, deliberate assertion failure; two phases and five completed asset timings survive in a 1,401-byte diagnostic copied by the real runner |
| Final PowerShell 7.4 isolated runner harness | Passed; missing/failed production-build gates, server-mode metadata, phase/asset retention, original completeness checks, progress-host faults, ZIP/privacy and recovery |
| Runner functions with actual local and uploaded Windows JSON | Passed; required title counts verified, returned 14/15 rejected with only terminal missing its third pass, and actual failure diagnostics decoded/copied |
| Runner-generated selection using the actual Playwright CLI `--list` | Exactly five titles each three times: **15 executions in two files**; selection only, not another executed suite |

The single cold-run comparison is local evidence of reduced runtime test
overhead, not a benchmark establishing the Windows cause or E08 acceptance.
The 36-case WebKit run independently includes three complete fallback/worker
terminal passes of 8.902s, 8.935s and 8.850s, with every original assertion active.

The first local diagnostic-output probe failed before starting tests: its
nested, local-only configuration used that config directory as the server cwd.
`diagnostic-probe.json` retains the setup failure. Only the probe configuration
was corrected to use the frontend root; this is independent of the main runner
configuration and is not counted as a product test pass.

Commands:

```text
npm run build -- --outDir .local/browser-test-build
# Set CI=1, BLORA_E2E_FRESH_SERVER=1 and the selected BLORA_BROWSER.
# Compiled runs additionally set BLORA_E2E_PREVIEW=1.
node node_modules/@playwright/test/cli.js test tests/browser/accounts.spec.ts tests/browser/terminal.spec.ts tests/browser/query-navigation.spec.ts tests/browser/logs.spec.ts --workers=1 --reporter=line,json
# WebKit additionally uses --repeat-each=3.
node node_modules/@playwright/test/cli.js test tests/browser/terminal.spec.ts --grep "fallback renderer, worker checkpoints" --workers=1 --reporter=line,json
pwsh -NoProfile -File scripts/test-windows-validation.ps1
```

Preview's absent-backend/teardown proxy warnings remain visible in the mocked
log suites. Those suites do not establish full backend/platform acceptance.
All owned test sessions and containers have ended; there is no runtime handle
to resume. No production service or UI/theme setting was changed.

## Next Windows run

Use the new pinned source and runner revision with
**`-Followup -BrowsersOnly -WebKitOnly`**. The existing five-title selection
remains required three times each: **15/15**, with no retry or skip. The four
Windows-passed polling scenarios guard the change to compiled serving alongside
the remaining terminal case. Return the report ZIP even if a stage fails.

Full Windows lifecycle, notification-center behavior, native IME coverage,
strict mixed-load E08 performance, actual Docker/cross-host acceptance and
physical power/cache-loss testing remain separate gaps. A successful narrow
run will not establish full platform or project acceptance.
