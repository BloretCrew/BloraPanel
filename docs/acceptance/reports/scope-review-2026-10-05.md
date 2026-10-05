# Acceptance scope review — 2026-10-05

The user asked for several delegated code reviews and removal of impractical
environment tests, including real disk exhaustion and disposable virtual
machines. Three read-only reviews covered storage/backup/transfer/scheduling,
Windows platform capabilities, and terminal recovery/performance/compatibility.
The parent checked the findings against source and existing reports.
Reviewed baseline: `4eb5eeef6bfc460cf78d816683bdd5ead6e2bea8`.
No product files changed and no tests, services or destructive experiments ran.

This changes the **required acceptance scope for this delivery**. It does not
convert unexecuted tests into PASS, erase old failures, or remove error handling,
authorization, durable task identity, source-data protection or ordinary tests.
The frozen UI and the E08 p95 ≤50ms target remain unchanged.

## Environment checks no longer required for this delivery

| Check | Accepted review/existing evidence | Remaining limitation |
|---|---|---|
| Physical power loss, device write-cache loss, many physical filesystem combinations | Commit intent, object/hash reconciliation, atomic checkpoints, SQLite WAL/FULL and real process-crash recovery | Physical-device durability was not tested; Windows directory sync is explicitly omitted |
| Actual host/NTFS disk exhaustion, including Docker Engine's storage device | Existing private 1MiB Linux ENOSPC tests, pre-mutation Unknown records, error propagation and interrupted reconciliation | Simultaneous state/log/cleanup write failures on a full physical device remain untested |
| Actual browser quota threshold and storage eviction | Three-engine recovery failure/transaction tests and bounded recovery-state handling | No guarantee of recovering locally erased/evicted content |
| Independent remote Engine/registry hosts, public-network long faults and real certificate-operation deployments | Real local Engine HTTPS/Compose tests, controlled network faults, two-version registry and CA-rotation integration | These environments are deployment compatibility checks, not newly verified platforms |
| Runs beyond the completed 24h soak and every desktop/platform combination | Existing 86,415.480-second real soak, 17,262 samples, two fault stages and backup restore | Longer or different combinations were not run |
| Additional Windows headless WebKit, natural timer cadence, cold-start/pointer/paint combinations, native dialogs/bfcache and system IME matrices | Existing visible Windows followup plus local lifecycle/terminal review | The untested combinations and old headless failures remain recorded; no repeat of the completed followup is required |
| Windows notification-center rendering and privileged service/task modification lifecycle | Existing notification implementation, native enumeration evidence and review of fixed-target, bounded platform actions | No actual Windows notification-center or privileged modification PASS is claimed |
| Windows-specific container combinations and unavailable graphics hardware | Capability-dependent platform contract and existing Linux Engine/native evidence | Windows container support and hardware E08 performance are not certified |

Existing safe ENOSPC, SIGKILL, permission, cancellation, schema migration and
ordinary success/failure tests remain in the repository. In particular, the
[real ENOSPC report](enospc-2026-09-19.md) already verifies move-source protection,
no incomplete backup publication and preservation of an existing restore target;
the user does not need to fill a disk or supply a VM to repeat those checks.
The [24h report](local-endurance-24h-2026-09-27.md),
[Engine HTTPS report](docker-https-transport-2026-09-21.md),
[network extension report](netem-extension-task-2026-09-22.md),
[CA rotation report](remote-catalog-root-rotation-2026-09-21.md) and
[browser recovery report](recovery-engines-2026-09-26.md) retain their original
source, scope and results.

## Code review findings

### Storage, publication and remote effects

- [Upload checkpoints](../../../internal/filesystem/transfer.go): data is synced
  before the confirmed offset is persisted; commit identity precedes rename,
  and a durability error requires reconciliation rather than success.
- [Backup publication](../../../internal/backup/create.go) and
  [restore](../../../internal/backup/restore.go): pending archive identity/hash
  is checked; interrupted restores inspect the existing upload and retain
  cleanup/unknown results instead of blindly repeating an overwrite.
- [Scheduling](../../../internal/backup/scheduler.go): stable occurrence, fire,
  task and request identities are persisted through preparation/acceptance.
- [Engine mutation](../../../internal/containers/operations.go) and
  [execution completion](../../../internal/containers/manager.go): Unknown is
  persisted before effects; uncertain remote results and failed final persistence
  do not become success or unconditional replay.
- [Backup directory sync](../../../internal/backup/io.go) explicitly limits the
  Windows guarantee; review is not proof of physical power-loss durability.
- [Remote registry transport](../../../internal/extensions/catalog.go) requires
  HTTPS, rejects redirects and bounds response size/time; package authenticity
  and capability checks still apply separately at installation.

No additional confirmed defect was found in these reviewed paths. This is a
bounded static review, not a claim that all code or environments are bug-free.

### Terminal and historical intermittent failures

[Terminal attachment](../../../web/src/apps/TerminalApp.vue) checks connection
generation/disposal around input, restore completion, parsing, errors and close.
[Terminal models](../../../web/src/services/terminals.ts) suppress replay/read-only
upstream input; [server grants](../../../internal/master/server.go) revoke active
streams. [Read lifecycle](../../../web/src/services/read-lifecycle.ts) invalidates
old frame/deadline resumes at pagehide and retries only its own cancelled safe
reads, with a bounded surviving-document fallback.

The old Vim extra-byte failure and single revocation timeout become
**non-blocking historical observations**, given subsequent unchanged assertions
and passing runs. Their root causes are not claimed fixed. Reopen on new failure
evidence; do not repeatedly run the same passing group to pursue an unbounded
absence-of-failure guarantee. See the
[diagnostics](local-diagnostics-2026-09-29.md) and
[RC34 functional report](terminal-generation-2026-09-29.md).

### Confirmed capability-gating defect — still open

[SystemApp](../../../web/src/apps/SystemApp.vue) treats any firewall backend
other than unavailable/unsupported as preview/apply-capable. Windows
[capability detection](../../../internal/systeminfo/capabilities.go) returns
`netsh`, which only supports
[status reads](../../../internal/systeminfo/firewall.go). Thus the Windows
“calculate difference” button can be enabled, but
[preview](../../../internal/systeminfo/firewall_diff.go) explicitly rejects
non-Linux platforms. The error is shown; this is not false success or an
authorization bypass. Separate read availability from Linux firewalld mutation
availability and verify the gate locally.

Windows firewall rule mutation/rollback is **outside the currently supported
platform subset**, not an implemented feature awaiting a dangerous test. Do not
add a Windows firewall implementation merely to satisfy the old test list.

## Remaining blocking work

| ID | Work | Execution location |
|---|---|---|
| R1 | Resolve/measure the known E08 latency failure while preserving UI and p95 ≤50ms | Local optimization/measurement; hardware-specific comparisons optional |
| R2 | Fix firewall read-versus-mutation capability gating | Local source and targeted verification |
| R3 | Add one non-destructive Windows real-backend and release-recovery runner | Local implementation, builds and runner checks |
| R4 | Execute that runner once against Windows Master/Daemon and the final release | User's ordinary Windows device; no VM, full disk, host firewall/service changes or remote access required |
| R5 | Final source-matched release verification and acceptance/documentation closeout | Local, after relevant fixes; preserve all historical results |

R3/R4 cover an owned temporary workspace: initialization/login, actual node
connection, a disposable native instance, file edit/save, ConPTY/terminal,
confirmed tasks and refresh recovery, representative backup/restore/monitoring
and extension paths, then stopped-state recovery and supported compatible
rollback. Consolidate these into one report rather than repeat old browser-only
followups. Product Master/Daemon entrypoints already exist; the current
[devfixture](../../../cmd/devfixture/main.go) and
[release smoke runner](../../../scripts/package-smoke.py) contain Linux-specific
commands/artifacts and cannot stand in for Windows end-to-end evidence.

E08 has actual complete-material failures (1003.9/1108/545ms in three engines).
Review does not change them into PASS. Obtaining a GPU, testing every cold-start
combination or supplying a remote host is not a new user prerequisite.

The [sixteenth Windows report](windows-sixteenth-report-2026-10-05.md) closes the
requested visible-WebKit followup. All waived/optional rows above cease to be
blocking tasks; only R1–R5 form the current remaining list. Ordinary defects
found during those tasks must still be fixed. Full original all-environment
certification is not claimed.
