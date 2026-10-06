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
