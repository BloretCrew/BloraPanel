# Windows report fixes and targeted retest

The user requested corrections for the failures in the
[third returned report](windows-third-report-2026-09-30.md), followed by a new
Windows runner. Changes start from `d7e5ef712c9326af3a602b8a7350af0b1aae823d`.
The UI, wallpapers, icon artwork and theme settings are unchanged.

## Corrections

| Finding | Change | Validation boundary |
|---|---|---|
| Directory metadata returns Access denied | Reopen the retained directory with `NtCreateFile`, explicit directory/attribute access and an empty relative name. Files still use `ReOpenFile`. No path lookup can redirect the already validated object. Add readonly/writable, timestamp, identity and retained-content checks. | Windows tests compile; actual Windows execution awaits the next report. |
| ConPTY receives redirected parent standard handles | Set `STARTF_USESTDHANDLES` with null standard handles for the pseudoconsole child. The test now requires both an executed command's file and a computed Unicode marker that is absent from the input, plus resize and Job close. | Windows tests compile; actual Windows execution awaits the next report. |
| Master fixtures require `/bin/sh` | Use a real owned Go executable for stop policy, stdin delivery, continuous logs, descendant monitoring, restart, maintenance and schedules on both platforms. | Linux real-process and all-package race tests pass; Windows execution awaits retest. |
| Race-instrumented fixture exits exceed the production stop deadline | Set only the child fixture's `GORACE=atexit_sleep_ms=0`. Race detection stays enabled and the production one-second stop deadline is unchanged. | Settings/transfer/bridge regressions pass in three consecutive race executions. |
| Unix permission and rename assumptions fail on Windows | Check Windows readonly/writable semantics. For retained-root rename denial, verify unchanged object identity and that releasing the owned root handle permits the rename. Existing-directory merge preserves the actual platform baseline. | Linux tests pass; Windows assertions compile and await execution. |
| ZIP bomb fixture is below the production ratio on this Go version | Use explicit best compression and first assert the fixture exceeds the unchanged 1000:1 production bound. | Filesystem tests and full race run pass. |
| Protocol deadline fails during unrelated TLS setup | Consume the initial token-bucket burst directly, then exercise the actual writer's unchanged 20 ms rate-wait deadline. | Protocol tests and full race run pass. |
| Firefox identity edits truncate redo / replay groups miss the saved cursor | Ignore semantic identity replacements when journaling; preserve a stack boundary at the saved history cursor, including older records. | Browser undo/redo, reload, move, Unicode and history-budget assertions pass. |
| WebKit editor shortcuts use Windows bindings for a Macintosh UA | Match test key bindings to Monaco's browser UA, including redo and document start/end. | All affected editor tests execute the intended commands; no production keymap override. |
| Extension note test fills an iframe before it receives real focus | Wait for real mouse hit testing and click the input; retain the immediate reload without waiting for a checkpoint ACK. The example sends each checkpoint in its input event rather than queuing behind an earlier round trip. | Firefox and WebKit note tests each pass three consecutive runs, including stable task identity and no replay. |
| WebKit reports `Load failed` instead of `fetch` | Expose a stable `NETWORK_ERROR` with explicitly uncertain outcome; preserve caller cancellation and never automatically retry. Only the iframe's parent may deliver bridge results. | Four transport unit cases, extension receipt tests and affected browser scenarios pass. |
| Abrupt node disconnect can fail a resumable transfer | Bridge send failures retain their transport cause and identify a closed management connection before the bridge reader finishes closing. Remote file-operation errors remain distinct. No RPC is automatically repeated. | New TLS bridge tests and real source-daemon restart/hash verification pass; all-package race regression passes. |
| Unicode undo test reads a stale WebKit frame | Confirm undo changes the painted text before capturing its expected result, and confirm redo restores it. Initial Unicode insertion still reloads immediately. | WebKit and Firefox each pass three subsequent runs; Chromium final verification is recorded below. |
| Delayed find-field focus steals fast replacement input | Focus after Vue has created the input, instead of a delayed animation frame; respect focus the user has already moved elsewhere and avoid focusing an unmounted/closed view. Verify both find and replacement field values before selecting matches. | Reproduced in Chromium as replacement text entering the find field with zero matches. After correction each of Chromium/Firefox/WebKit passes 3/3. |

## Local results

Commands run in `/data/instances/blora-panel`, or its indicated subdirectory.

- `go test -race -count=1 -p=2 -timeout=12m ./...`: **exit 0**. Master package
  296.382 s; extensions 173.993 s. Windows-only and opt-in environment checks
  are not Linux runtime evidence.
- `go test -race -count=3 -timeout=4m ./internal/bridge ./internal/master -run
  '^(TestSendPreservesClosedTransportIdentity|TestRemoteOperationErrorIsNotTransportLoss|TestInstanceSettingsAreVersionedAndRootsRemainSeparate|TestTransferSourceDaemonRestartRebuildsProof)$'`:
  **exit 0**, bridge 1.088 s / Master 38.544 s.
- `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c` for filesystem,
  terminal, Master and bridge: **exit 0**, using a writable temporary Go cache.
  This proves compilation, not Windows behavior.
- Frontend `npm run build` (includes type checking) and `npm test`:
  **exit 0**, **80/80 unit tests**, repeated after the final find-focus product
  correction (Vite build 3.55 s, Vitest 1.33 s).
- SDK and independent reference extension builds, default package and both
  versioned fixture packages: **exit 0**.
- Official `mcr.microsoft.com/playwright:v1.63.0-noble` container, nine affected
  browser files, one worker and a fresh server: Chromium **33/33** (4.2 min);
  Firefox **32 pass / 1 native IME skip** (197.227 s). WebKit's first expanded
  run was **31 pass / 1 fail / 1 native IME skip**; the remaining failure was
  the stale undo-frame assertion above, preserved in the original JSON. The
  corrected case subsequently passed **3/3** (34.6 s); Firefox also **3/3**
  (31.5 s). These are separate runs, not an invented clean combined run.
- A subsequent Chromium Unicode repeat found the separate find-focus race:
  **2 pass / 1 fail** (35.1 s), retained in `chromium-unicode.json`. With the
  focus correction and explicit input-value assertions it passed **3/3**
  (34.1 s), recorded in `chromium-unicode-final.json`. The same final source
  passed Firefox **3/3** (41.6 s) and WebKit **3/3** (43.5 s), with separate
  `*-unicode-final.json` files. These repeats have no retry allowance.
- Extension note immediate-reload cases independently: Firefox **3/3**
  (48.8 s), WebKit **3/3** (54.0 s).
- `scripts/test-windows-validation.ps1` in the official PowerShell 7.4 Linux
  container: **exit 0**. Checks broken progress rendering, literal arguments,
  child progress suppression, failures/skips/timeouts, report coverage gates,
  checksum/traversal-safe compiler extraction, ZIP boundaries and old-run
  recovery. This is not a Windows PowerShell 5.1 runtime result.
- `go vet ./internal/bridge ./internal/filesystem ./internal/master
  ./internal/protocol ./internal/terminal` and `git diff --check`: **exit 0**.
  New/updated local documentation links resolve.

The earlier targeted race failure remains recorded: a fixture's one-second
race exit sleep caused stop timeout, and raw connection EOF interrupted source
proof rebuilding. Neither was hidden by increasing a production deadline or
skipping the real restart scenario.

Local browser JSON evidence is under `.local/evidence/windows-fixes-2026-10-01/`.
Original Windows ZIPs and private machine paths are not published.

## Next Windows run

Use [the runner](../../../scripts/windows-validation.ps1) with `-Retest
-InstallRaceCompiler`, downloaded from the same exact commit as `-Ref`.
It makes a fresh checkout, generates the three required SDK packages, and runs:

- 27 explicitly named Go regressions, checking actual PASS by package and name.
- 11 required native checks, including metadata and real ConPTY execution.
- All-package race tests, previously blocked by a missing compiler.
- 32 scenarios in nine affected files on each of Chromium, Firefox and WebKit.
  The previously passed Chromium-specific native IME injection is intentionally
  omitted; Firefox/WebKit native IME remains unverified. A real `--list` run
  confirms the final exclusion selects **32 tests in 9 files**.

The optional compiler is pinned to an official WinLibs release and verified
against its fixed SHA-256 before extraction or execution. It is confined to
this new run directory; PATH changes apply only to the validation process.
No administrative access, persistent compiler installation or global execution
policy change is needed for this compiler.

Missing named tests, wrong-package matches, skips and successful retries after
failure cannot count as the required regression passing. A final report ZIP is
attempted even when a stage fails; plain-text progress avoids the previously
broken progress renderer. Prior report results are not imported as new passes.

Windows directory metadata, ConPTY and the new portable fixtures are still
**pending real Windows execution**. SCM/Task Scheduler lifecycle changes,
firewall rollback, native notification interaction, full-stack Windows browser
acceptance, strict E08 performance, external Engine/network environments and
physical failure scenarios retain their existing gaps. This change is not full
platform or whole-project acceptance.
