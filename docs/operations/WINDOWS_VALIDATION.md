# Run Windows validation and return the report

This runner downloads a fresh checkout of BloraPanel from GitHub and runs the
existing automated Windows checks without remote access. It never updates an
existing checkout. Run it in **Windows PowerShell**, not WSL, on Windows 10 1809
or newer / Windows 11, preferably x64. Use a normal terminal first; do not disable
antivirus or change execution policy globally when a test fails.

## One command to start

Paste these lines into PowerShell:

```powershell
$script = Join-Path $env:TEMP 'Blora-Windows-Validation.ps1'
Invoke-WebRequest -UseBasicParsing 'https://raw.githubusercontent.com/BloretCrew/BloraPanel/main/scripts/windows-validation.ps1' -OutFile $script
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -InstallTools
```

`-InstallTools` installs missing Git, Go and Node.js LTS through `winget` (normal
installer/elevation prompts may appear). If these are already installed it does
not replace them. Without winget, install Git, Go **1.25+**, and Node.js **24+**,
open a new terminal, and rerun. Network access to GitHub, Go/npm registries and
Playwright browser downloads is required. Allow several GB for the checkout,
dependencies, browser engines and logs; execution can take an hour or more.

Windows Go race testing additionally requires a compatible **mingw-w64 GCC** on
PATH. The runner automatically uses `gcc.exe` when available. Without it, the
normal/native tests still run and race testing is explicitly **BLOCKED**. A
broken or incompatible compiler records a failure, never a pass.

Optional parameters:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File $script -WorkRoot 'D:\BloraValidation' -Ref main
```

Each invocation creates a new timestamped directory. `-Ref` can be a branch,
tag or commit; the exact fetched commit is included in the report. The console
shows the current stage, elapsed time and live test output. A failed stage does
not stop independent later stages. Individual commands have time limits.

## What runs

- Windows Master, Daemon and extension signing tool builds; Go vet.
- Explicit native Job/keeper/daemon-exit, ConPTY Unicode/resize/close, log pipe
  recovery, monitoring/process identity and SCM/Task Scheduler enumeration tests.
  The report checks that each expected native test actually emitted PASS.
- All Go package tests without cached results, then all race tests if GCC exists.
- SDK and independent reference extension builds/packages.
- Frontend build and unit tests; complete existing browser suites on downloaded
  Chromium, Firefox and WebKit, one browser at a time. Port 5173 must be free; an
  unrelated running server is never silently reused.

The runner clears inherited Blora test opt-ins so it cannot accidentally use a
previous Engine endpoint or privileged test fixture. Native tests create their
own processes/files and perform read-only service/task queries. It does not
change existing services, scheduled tasks, firewall rules or system time.

## Send back the report

At the end the console prints the absolute path to **Blora-Windows-Report.zip**,
normally under `%USERPROFILE%\BloraValidation\<timestamp-id>\`. Send that ZIP
back in this conversation, including when tests fail. It contains:

- `README.md`: stage summary, skip and coverage gaps.
- `report.json`: exact commit, OS/tool context, durations/exit codes, test events.
- Per-stage streaming logs and browser JSON results; SHA-256 checksums.

The ZIP excludes the checkout, caches, test state and browser traces. User/work
paths are scrubbed from ordinary text logs, but test output can still contain
machine identifiers or escaped paths; review the bundle before sharing. Do not
send the entire work directory. `steps-so-far.json` and logs are saved during the
run; forcibly closing the terminal or killing PowerShell can prevent final ZIP
creation. In that case send the `report` folder after reviewing it.

Exit 1 means failures, exit 2 means incomplete/skipped/missing coverage. Even
exit 0 is named `AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS`, not full acceptance.
The report is the authoritative list. Timeout cleanup targets only the runner's
child process tree; keeper/breakaway behavior can leave test-owned processes
after a failed test, so inspect failures before rerunning. Never kill unrelated
Daemon/business processes to clean up a test.

## Explicit coverage limits

Service/task **enumeration** is not service/task modification lifecycle proof.
Windows system notification center interaction, actual firewall rollback,
physical power loss, device cache loss, remote Engine/certificate/network
operations and long-duration mixed-load E08 are not automated here. Existing
real-browser fixtures depend on Linux `/bin/sh` and `/proc`; browser suites in
this script exercise actual browsers/xterm/storage with API doubles, not a
Windows Master/Daemon browser end-to-end fixture. Opt-in Docker tests are
reported as skipped. These gaps stay visible even if every executed test passes.

The runner's parser/control flow/report packaging can be checked on Linux with
`pwsh -NoProfile -File scripts/test-windows-validation.ps1`. That harness is not
evidence that Windows native behavior has passed; the returned Windows report
provides that evidence.
