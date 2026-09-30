# Second returned Windows report

Source: user-supplied `Blora-Windows-Report.zip`, SHA-256
`94e981d1a5731be07348e2a846cae218eb8df519461aa3984a3ea9a1f4033ae7`.
Checkout `e5b940541dca461edf26d044c3c2b550e75b8115`, remaining mode,
Windows NT 10.0.26200 amd64, PowerShell 5.1.26100.9168.
Private paths and raw logs are not published.

## Evidence

The runner completed and packed the report. No fatal progress/ZIP failure.
Go was not found on PATH: **no native or all-package Go tests ran**. Previous
native passes are not new passes, and previous code fixes remain unverified on
Windows. SDK TypeScript builds passed, but WASI packaging failed with
`spawnSync go ENOENT`; generated extension package was absent in every browser.
The former `/C:/` path error did not recur before this missing-tool failure.

| Browser | Passed | Failed | Skipped |
|---|---:|---:|---:|
| Chromium | 118 | 1 | 0 |
| Firefox | 115 | 4 | 0 |
| WebKit | 111 | 8 | 0 |

All three browser extension-package failures have the missing Go/package
prerequisite. Chromium's only failure was this case. Firefox additionally had
14.2969px versus 14.3px serialization, editor view recovery, and a test calling
Chromium-only CDP. WebKit additionally had CDP, editor undo/reload, history budget,
multicursor click stability, wallpaper-settings click stability, instance draft
restoration, and a terminal-test task-request access-control error.

## Response

- Runner refreshes registered Machine/User PATH before probing tools and checks
  common Go install locations. It still marks missing tools BLOCKED and does not
  silently install without `-InstallTools`. Missing Go explicitly blocks WASI
  packaging instead of invoking a known missing compiler.
- Font scale assertion allows less than 0.01 CSS pixel serialization difference,
  preserving the actual 14.3px target. Linux Firefox targeted test passed 1/1;
  this is not Windows revalidation.
- Native Chromium composition injection is separated from the cross-browser
  committed Unicode/multicursor/find-state test. Firefox/WebKit native IME is
  explicitly skipped/unverified, never passed by substitute input. The runner
  continues to flag skips as incomplete coverage.
- PowerShell isolated harness passed. Actual Windows PATH discovery still needs
  a new device run.

Remaining editor persistence, WebKit interactions/request errors, ConPTY and
first-report Go fixtures remain open. Do not ask for another unchanged full run
or claim that all browser failures are environment errors.

The separated Firefox Unicode/multicursor case was then executed locally:
one failed, native-CDP case explicitly skipped. After the second refresh the
editor reports that draft history and body differ and refuses restoration.
This reproduces a real recovery inconsistency on Linux Firefox too; it is not
resolved by the test capability split. Evidence is in the local Playwright
`desktop-committed-Unicode--2d2d5-te-survive-immediate-reload` trace/context.
