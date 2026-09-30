# First returned Windows report

User-supplied `Blora-Windows-Recovered-4850cc32.zip`, SHA-256
`8cd7cbf2a63212d8d5bc918395389ff45ff256b4eacbb1fe97b57ab96d892a17`.
Tested checkout: `3943dca51636da322b9e3bb6727bbd54561c032f`.
Reported Windows NT 10.0.26200, amd64, Go 1.27.0, Node 22.17.1.
The private original is retained as the conversation attachment; logs contain
machine paths and are not copied into public documentation.

## Actual outcomes

- Native selection: 21 passing test events and one failure. Of the ten mandatory
  named native checks, nine passed. Job keeper/descendants/daemon exit recovery,
  log pipe survival, native process metrics, service and task enumeration passed.
- ConPTY Unicode/resize/close test failed in both native and all-package runs.
  Captured output contained terminal setup but not the expected Unicode marker;
  console prompt output also appeared outside the captured stream. Cause remains
  unproven; do not change timeouts or delete this assertion to obtain a pass.
- Full Go run: 316 pass / 26 fail / 14 skip events (including subtests, not a count
  of distinct top-level cases). Container package additionally terminated with
  `runtime.semawakeup`, `setevent failed; errno=6`.
- Builds for Master/Daemon/signing tool and Go vet passed. SDK TypeScript build
  passed, but extension packaging failed on `/C:/...` filesystem paths.
- Web production build and all 76 unit tests passed. Browser tests never ran:
  the old runner's progress host failed after Chromium download. No browser
  runtime result is inferred. Race testing lacked GCC and was BLOCKED.

## Corrections from this evidence

1. Reference package converts file URLs using `fileURLToPath`, for both esbuild
   paths and Go WASI compiler arguments; `.pathname` is not a Windows path.
2. Runner generates the default reference package as well as v2/v3 before Go
   integration tests. Previously Go ran before packaging, and the later fixture
   command generated only v2/v3. Missing-package failures were real setup errors.
3. Windows Compose CLI deferred CloseHandle captured each original handle by
   value, then closed the same handles explicitly after process creation. Deferred
   cleanup now checks the mutable slot. This removes a definite double-close
   that could close a reused runtime event; it is not yet Windows confirmation
   that every runtime crash has been resolved.
4. Windows metadata handle now requests READ_ATTRIBUTES as well as
   WRITE_ATTRIBUTES. Go's handle-based Chmod reads existing attributes first.
5. Container log archive directory sync is platform-specific. Linux retains
   directory fsync; Windows retains file FlushFileBuffers before rename without
   claiming unsupported directory/power-loss durability, consistent with other
   project Windows archive adapters.

The progress renderer and ZIP recovery fix was already delivered as `2bf08c5`;
the returned report predates it.

## Local verification of corrections

`go test ./internal/daemon ./internal/filesystem ./internal/containers -run
'TestContainerLogArchive|TestMetadataVersion|TestUploadMetadata|TestComposeOutput'
-count=1` passed on Linux (0.135 / 0.024 / 2.643 seconds). Windows amd64 test
binaries for all three changed packages compiled with `CGO_ENABLED=0`.
`npm run package` in the independent reference extension built the actual
frontend/WASI/default package. PowerShell isolated harness passed all outcomes,
progress suppression and interrupted-report recovery. These checks do not
replace another Windows execution.

## Still open after this patch

- ConPTY missing input/output marker: investigate console/std-handle ownership
  and startup sequencing using a bounded native probe, not a blanket timeout.
- Several Master integration fixtures execute `/bin/sh`; they need equivalent
  Windows test processes. They are not proof that native Job execution failed.
- Filesystem mode assertions assume Unix 0710/0750/0700. Windows READONLY and
  directory-lock semantics require explicit equivalent assertions, not silent
  skipping or a claim of POSIX ACL preservation.
- Root-renaming test met Windows open-handle sharing denial; archive expansion
  fixture did not trip its expected limit. Check each fixture and production
  contract separately; neither has been declared solved.
- Bulk write deadline test failed during its initial 600KiB send (20ms budget),
  before the intended rate-limit wait. Separate setup from the property under
  test without weakening the configured timeout behavior.
- Windows re-execution of the corrected metadata/log/CLI paths, race compiler,
  three browsers, and the previously documented external/platform gaps remain.

The earlier statement that no definite local fixes remained is superseded by
this first real-device evidence. Full Windows acceptance remains incomplete.
