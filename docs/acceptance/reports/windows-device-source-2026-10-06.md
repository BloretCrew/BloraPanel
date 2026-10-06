# Windows device source acquisition — 2026-10-06

The user supplied console output from fixed runner/source `2bfac47830c8e54ad122014e0fc4d821116f619d`.
Stage `clone` failed after 30.3 seconds with reported exit code -1. The checkout,
builds and fourteen real-device checks had **not started**. No ZIP contents were
supplied in this message, so the actual Git/network failure cannot be established
from the console excerpt alone. This is not fourteen failed product tests.

The wrapper now uses structured child arguments, explicit UTF-8 output, bounded
stream reads and child exception logging. Native stderr diagnostics/progress do
not trigger a terminating Windows PowerShell error before the real exit code is
collected. Startup failures use a missing exit code rather than a fabricated -1;
actual negative child exit codes remain failures. Stage detail is kept in logs
and report JSON. This improves compatibility and diagnosis; it does not prove
the original machine's cause without its log.

Source download can use `-ArchiveOnly`, obtaining the HTTPS GitHub codeload ZIP
for the same full fixed commit without Git. The default Git path also falls back
to that archive after a clone/checkout failure. The original failure is retained
and any recovered source acquisition is explicit in report metadata. All
fourteen real-device requirements remain unchanged.

Archive extraction validates the fixed commit root, complete path set,
case-insensitive duplicate destinations, traversal, Windows unsafe path forms,
symlinks and bounded archive/expanded sizes before writing source. It requires a
new task-owned source directory. Both source methods still compare the runner's
SHA-256 with the exact source script before executing build/test commands.

Local verification:

- `scripts/test-windows-device-validation.ps1` passed in the existing isolated
  PowerShell 7.4 Linux container. Successful native stderr, Unicode/literal
  arguments, actual exit 7, missing-command error/exit 1, report checksums and
  private-file exclusion were checked. Valid fixed-root ZIP extraction passed;
  traversal, wrong commit, duplicate destinations and symlink ZIPs were rejected
  before creating source files.
- The actual GitHub HTTPS ZIP for `2bfac47830c8e54ad122014e0fc4d821116f619d`
  downloaded successfully and safely extracted **1,218 source files** using the
  revised PowerShell extraction function. No archive code was executed.
- `git diff --check` passed. No product/UI/Go checker changes or release builds
  were made. Prior real-backend 14/14 evidence is unchanged, not rerun or relabeled.

Windows execution of the revised wrapper remains unverified. The next ordinary
device command uses the newly delivered fixed version with `-ArchiveOnly` to
bypass the already failed Git path. The original failed report remains historical
evidence; E08 performance and deferred release work remain separate.

## Follow-up: identical archive-stage exit

The user's console for fixed `ba68e5d444e6a5c1307a3928252a1042a3d994cc`
with `-ArchiveOnly` reports the download stage exiting `-1` after 30.2 seconds,
with no child diagnostics shown. This is not a timeout under the script's
150-second deadline. No ZIP from that run has been received; the network,
child startup or external termination cause cannot be certified from console.
None of the fourteen product checks started.

Both stage and downloader now use plain task-owned `.ps1` files rather than
nested encoded PowerShell commands. Stage arguments are read from UTF-8 JSON
without shell interpolation. Logs include the child PID and explicit stage
startup/download-start/download-complete markers. Helper/spec files remain
outside the report archive; all original requirements and failure retention
remain intact.

The updated portable harness passed in PowerShell 7.4's isolated Linux container,
including paths containing spaces, Unicode/literal arguments, normal stderr,
exit 7, missing-command exit 1, startup marker, report checksums, private-file
exclusion and unsafe ZIP rejection. It rejects any reintroduced encoded launcher.
The exact new downloader, invoked through the new stage launcher, downloaded
the actual fixed `ba68e5d` GitHub source archive in 4.0 seconds and safely
extracted 1,219 files. Task-owned containers and temporary files were cleaned.
These are local wrapper results, not Windows acceptance or proof of the user's
root cause. The next action is one fixed-version Windows run with `-ArchiveOnly`.
