# Non-release acceptance reconciliation review — 2026-10-08

The initial work was a read-only review of source, acceptance records and test
configuration. A subsequent parent assignment authorized repair and targeted
browser verification of the recovery-writer finding below. The parent remains
responsible for the current-source complete regression, Chromium screenshot
investigation and final integration results. This report does not predeclare
those results; no release, commit or push was performed by this reviewer.

The user requested completion of three remaining non-release tasks: current
source regression, code/documentation closeout and commit/push, and resolution
of the unstable Chromium screenshot comparison. Final release remains deferred.
The earlier E08 optimization scope is not reopened.

## Effective scope and evidence

- The [October 5 scope review](scope-review-2026-10-05.md) removes specified
  impractical environment checks from this delivery. It does not change their
  unexecuted status to PASS or remove error handling and ordinary safe tests.
- [Ordinary Windows validation](windows-device-passed-2026-10-06.md) closes R4
  with fourteen named real-backend checks. The report also records its fixed
  source, exact scope and earlier failures. It does not certify privileged
  Windows host mutation, all container combinations or physical failures.
- The [E08 closeout](e08-bounded-closeout-2026-10-08.md) records the user's
  acceptance of feedback around 100ms under the exceptional mixed workload.
  Chromium reached the original 50ms target; the final Firefox/WebKit runs did
  not. WebKit's 89–132ms variation remains visible. No universal 50ms or stable
  100ms guarantee follows from this scope decision.
- Existing Linux systemd/cgroup, firewalld, Docker/Compose, notification,
  backup/process-crash, ENOSPC, extension and 24-hour reports remain evidence
  for their stated source and environment. A current-source regression must
  identify what it rechecks rather than silently relabel old runs as new ones.

Windows `netsh` firewall status is supported; Windows rule mutation/rollback
is outside the declared platform subset. The corrected frontend gate separates
read availability from Linux firewalld mutation availability. See the
[local R2/R3 report](local-remaining-2026-10-05.md).

## Reconcile current table rows without rewriting history

The matrix's main F/A/E tables still contain older implementation states and
phrases such as “complete scenario not passed,” “Windows pending” and “physical
power loss pending.” Their latest top section overrides them, but a reader who
looks directly at a numbered row receives an inconsistent current status.

After the parent completes the current-source checks, update implementation
and verification separately. Suitable verification wording is **“Verified in
the current delivery scope; limitations retained”**. “Scope waived,” “not
supported on this platform,” “historical PASS,” and “new current-source PASS”
must remain distinct.

| Rows | Recommended reconciliation | Evidence and limitations to retain |
| --- | --- | --- |
| F01–F09 | Mark implementation complete where no new regression remains; replace obsolete blanket incomplete assertions with scoped verification | Current-source functional report plus existing A01–A17 reports and ordinary Windows report |
| F10, E01 | Mark implemented and verified for the supported monitoring/process subset | Real bounded history, 24-hour soak and Windows sampling; extra desktop/physical combinations are not newly tested |
| F11, E02 | Mark implemented and verified for the declared Engine/Compose subset | Real Linux Engine/Compose, HTTPS and safe ENOSPC; remote hosts and unavailable Windows container combinations retain scope limits |
| F12, E03–E04 | Mark implemented and verified in the current delivery scope | Real backup/restore and scheduling/crash evidence, ordinary Windows backup/restore; physical power/write-cache loss remains untested |
| F13, E05 | Mark implemented and verified for the explicitly supported platform subset | Linux private systemd/firewalld evidence, corrected capability gate; Windows firewall mutation is unsupported, not a missing dangerous test |
| F14, A13, E06–E07 | Mark implemented and verified in the current delivery scope | SDK/independent package, real lifecycle/permission/migration/network evidence and ordinary Windows lifecycle; public registry operations and all remote deployments are not certified |
| A01–A08, A10–A12, A14–A17 | Keep implementation complete and original results; add current-source regression linkage | Do not replace source-specific historical results with current-run claims; A11/A12 safe injected/ENOSPC checks remain useful |
| A09 | Mark implementation complete and scoped verification after current relevant guards pass | Existing bounded slow-consumer/control/transfer evidence and E08 protected output/data checks; optional longer physical environments are not blockers |
| E08 | Keep “closed under the user's adjusted scope” | Original 50ms failures, measurements, variability and unchanged workload remain recorded; no renewed automatic optimization |
| E09 | Split non-release readiness from release verification | Supported initialization/runtime/build/development evidence is retained; final artifact matching, packaged compatibility rollback and publication remain deferred |

The detailed historical evidence paragraphs can remain. Prepend a short current
disposition and links to the scope review, Windows evidence and new integrated
regression report. Do not remove failed reports or obsolete-source evidence.

The progress file's older B03–B09 “in progress” rows, B04 “Docker/cgroup/Windows
not run” wording and “current 0/14, 0/17, 0/9” statistic are historical. Label
that table explicitly as historical; add a concise current B00–B09 disposition
and a B10 disposition distinguishing non-release closeout from deferred release.

The remaining-work file's old final continuation paragraph still directs R1
optimization using 923–953ms figures. Label it historical or replace its current
instruction with the accepted E08 closeout and the actual current task. The
October 5 report's R1–R5 table is a dated record, so do not rewrite its original
results; link forward from current documents instead.

## Preserve rejected prototypes while keeping normal checks clear

The existing separation is sound:

- `web/playwright.config.ts` selects `tests/browser` for ordinary checks.
- `web/playwright.experiments.config.ts` selects `tests/experiments` and describes
  rejected/unadopted prototypes as executable diagnostics.
- `web/playwright.real.config.ts` selects `tests/real`; performance diagnostics
  require explicit environment options and default to the baseline. The final
  E08 report states that its formal runs used no rendering/positioning controls.
- A source search found no production import of the experiment helpers and no
  production `BLORA_PERF_*` branch. The private native-row adapter is a test
  helper, not a deployed terminal implementation.

Keep strict failures in the experiment directory and their recorded reports.
Their names and opt-in command should be documented; they must not be presented
as default product regressions or as adopted optimizations. Shared helpers that
support an ordinary regression or the explicit real diagnostic should remain
available. `terminal-native-rows.spec.ts` exercises a test-only adapter; its
result proves that adapter's comparison, not product installation of it.

The ordinary `window-position.spec.ts` screenshot comparison covers the
retained software positioning implementation. It must stay in ordinary checks
and be investigated rather than moved to experiments, skipped or given looser
pixel tolerances solely to make a green result. If its setup is corrected,
record the original failure, explanation and unchanged functional/pixel
assertions in the new report.

Generated `.local` browser/fixture evidence is intentionally ignored. Public
documentation should carry sanitized commands, numeric outcomes, source/bundle
identities, environment and limitations. Do not force-add private fixture state,
credentials, real browser snapshots, raw process logs or source directories just
to make historical links downloadable.

## Completion wording

Once the parent has verified the final source, resolved the screenshot issue
and completed the authorized repository handoff, say **“The three non-release
tasks are complete; final release remains deferred.”** Include actual test
counts, final commit/push result and screenshot findings from that work. Do not
say “all original environments passed” or “all browsers meet 50ms.”

No additional Windows rerun, full disk, VM, GPU purchase or privileged host
operation is a prerequisite introduced by this review. Reopen a covered area
only for a new relevant code change or failure, not for an unbounded search for
absence of all intermittent behavior.

## Follow-up code review: terminal transport and recovery

The parent requested an additional bounded read-only review of the grouped
terminal transport and background recovery path. Reviewed production files
include `internal/master/terminal_frames.go`, its stream caller,
`web/src/services/terminal-events.ts`, the TerminalApp/TerminalModel consumers,
and the recovery background writer, native snapshot transaction and terminal
delta restore code. The new grouped PTY integration and existing worker/delta
guards were read without execution at the review stage. The subsequent
targeted repair verification is recorded separately below.

### Confirmed finding: initialized recovery worker cannot fall back after failure

**P2 — confirmed, reported and repaired in the authorized follow-up.** The next
paragraphs describe the pre-repair behavior and its impact.

`BackgroundSnapshot` permanently closes itself on `onerror`, `onmessageerror`
or its request timeout. Pending and subsequent requests reject. In contrast,
`RecoveryService.flush()` catches this failure only by setting the protection
warning; it leaves `background` and `mirrored` installed. Future writes still
use the closed worker, while `enableBackgroundPersistence()` returns early
because a background object already exists. `persistenceMode` still reports
`worker`. Initialization failure has a foreground fallback, but failure after
initialization does not.

The consequence is not a premature durable database confirmation: failed
batches do not prune their recovery tail. New commits can still be protected
by the immediate sessionStorage journal until its 256KiB limit is reached.
However, with an otherwise healthy IndexedDB database, the durable head can
never advance again without a full page/service restart, and terminal output
eventually encounters the unprotected-tail/ACK guard. This is a newly introduced
availability regression in the optional background persistence path.

A local repair can distinguish worker transport/unavailability failures from
actual native storage failures. For a permanently unavailable worker, dispose
and clear its mirror/durable assumptions, then write the current complete
workspace through the existing foreground native transaction. An IDB
abort/quota/storage failure must retain its existing failure handling and old
head; it is not evidence that the worker is unavailable. A targeted browser
case should fail a worker after successful initialization and verify foreground
convergence and the latest recovery journal. No special external environment
is needed.

The current worker guards cover a blocked constructor and a native worker
transaction abort/retry. They do not cover a successfully initialized worker
that becomes permanently closed. TerminalModel's separate checkpoint worker
already catches mirror failures and downgrades to the native main-thread
serializer; this finding concerns the recovery IDB writer, not that existing
terminal fallback.

### Reviewed invariants with no additional confirmed defect

| Boundary | Source observation |
| --- | --- |
| Authorization | Stream handshake retains TLS/Origin/resource/owner checks. Current-user permission checks remain in the output loop and input/lease path. Grant mutation cancels registered user streams. Grouping itself performs no authorization or privileged action. |
| Archive identity | Groups contain only already-read archive events, retain their bytes and event order, and identify the final complete archive cursor. Invalid encoding is rejected before acceptance. |
| Frame budget | Normal grouped payloads are at most 64KiB and 128 events. Oversized legacy events are preserved intact rather than truncated; actual archive events are bounded separately at 32KiB raw output. Archive reads remain bounded including headers. |
| Wire credit | Protocol sends charge the entire encoded payload against cumulative byte sequence/credit. Browser grouping sums those payload bytes, checks each frame's sequence, and acknowledges the final byte boundary only after ordered parsing/protection. Archive sequence and transport byte sequence remain separate. |
| Consumer limits | Envelope size, bounded pending payload bytes, server credit and the original slow-consumer deadline remain present. No new coalescing timer was added to the transport. |
| Parser/protection | TerminalModel checks contiguous event identities and resize sizes, caps each native parse group at 64KiB, awaits visible and checkpoint parsing, commits parsed output to the immediate journal, and fails confirmation when recovery is unprotected. Historical terminal input is not journaled. |
| Snapshot publication | Native snapshots, pointer replacement and stale-head cleanup remain in one IDB transaction. Tail pruning follows successful transaction confirmation, not submission. Mixed metadata/editor changes require a full state path. |
| Durable deltas | Deltas contain only contiguous terminal-output entries, at most 64 entries/128KiB, and refer directly to a complete base. Both retained heads keep their required bases; restore validates revision/key/base identity and rejects a missing or nested base without overwriting the retained records. |
| Worker queue | The foreground flush awaits one batch; the worker serializes requests. Only synchronous cloning failure uses the existing JSON sanitization retry. Native transaction failure is not translated into success. |

These observations are a bounded code review, not replacement PASS evidence
for the parent's current-source complete tests.

## Authorized recovery-worker repair and targeted verification

Changed only `web/src/recovery/background-snapshot.ts`,
`web/src/recovery/service.ts` and `web/tests/browser/recovery-worker.spec.ts`.
No distribution bundle was rebuilt: the parent's first full real-backend
regression continued against its existing bundle.

`BackgroundSnapshotUnavailableError` now identifies permanent worker/transport
failure separately from native storage errors returned by the worker. Only
that failure can trigger a full foreground snapshot retry. The service clears
the stopped worker, mirror and durable assumptions before using the existing
atomic native transaction. It retains the immediate journal until transaction
confirmation. Actual Abort/Quota/storage failures retain the original warning,
head and retry policy; no remote command or input is replayed.

The retry also checks writer identity and generation. Explicit stop/logout
does not invoke the fallback. A successful old-worker response cannot drain
late teardown edits through a new foreground writer after that generation was
stopped. Account cleanup remains responsible for removing its own records.

### Original failure evidence

The first new case mistakenly waited for `.topbar` in a direct service fixture
whose API responses deliberately do not authenticate a desktop. That setup
failed before exercising recovery and is retained separately; only the new
test's incorrect UI precondition was removed.

With the actual fallback catch temporarily absent, the corrected runtime-worker
case then failed as expected: the immediate journal restored the new text, but
the mode remained `worker`, protection was false, and database-only recovery
returned the old text. The repair was restored immediately before the passing
runs. Original and corrected setup failure evidence remains in
`web/.local/recovery-worker-runtime-before.log` and
`web/.local/recovery-worker-runtime-baseline.log`.

### Final checks

All runs use the matching official Playwright `v1.63.0-noble` container, a
fresh Vite test server on local port 18575, and one browser worker. The cases
exercise native browser IndexedDB and Worker implementations; mocked API
responses isolate this local persistence boundary, so these are not new
real-backend or Windows results. No E08 pressure run was added.

The six cases cover immediate journal protection, constructor unavailability,
initialized worker failure with immediate/database recovery, in-flight explicit
stop, explicit stop after a successful worker response, and the original native
worker transaction abort/retry. The last case now also confirms that an actual
storage abort does not switch to foreground mode.

| Browser | Final result | Duration | Exit code | Local log |
| --- | --- | ---: | ---: | --- |
| Chromium | 6/6 PASS | 11.3s | 0 | `web/.local/recovery-worker-runtime-final-chromium.log` |
| Firefox | 6/6 PASS | 15.8s | 0 | `web/.local/recovery-worker-runtime-final-firefox.log` |
| WebKit | 6/6 PASS | 25.1s | 0 | `web/.local/recovery-worker-runtime-final-webkit.log` |

Command: `npx playwright test --config=playwright.config.ts
tests/browser/recovery-worker.spec.ts --workers=1 --reporter=line`, with
`BLORA_BROWSER` selecting each browser, `BLORA_E2E_PORT=18575` and
`BLORA_E2E_FRESH_SERVER=1`. Chromium uses the container's matching full browser
at `/ms-playwright/chromium-1243/chrome-linux64/chrome`.

Final type checking (`cd web && npm run check`) and whitespace validation
passed. All three targeted browser processes exited normally. The parent
will incorporate these source-level results into the final integrated report.

## Real aborted-pointer regression: native Worker fault injection

The parent's first complete real regression reported the aborted-pointer case
in `web/tests/real/recovery-failure.spec.ts` failing after 6.8s. Source review
confirmed a test realm mismatch: it patched the page's native
`IDBObjectStore.prototype.put`, while production snapshots now commit in a
separate Worker realm. That injection could not abort the production writer's
transaction. This does not demonstrate a pointer-publication defect.

The case now intercepts only the actual production `snapshot.worker-*.js`
asset, prepends a native Worker-realm `put` wrapper, and enables it only after
the original full snapshot exists. The wrapper invokes the original native
`put` and immediately aborts its owning transaction on pointer writes.
Test-only control messages are intercepted before the worker's normal request
handler; the real Worker acknowledges activation and its actual abort count.
Background persistence stays enabled throughout. Reload creates a new Worker
with injection disabled, preserving the original normal recovery step.

All existing failure/protection assertions remain: a visible protection warning,
at least one actual pointer transaction abort, exact equality of both old
pointer and snapshot records, latest draft export from the synchronous journal,
latest text after immediate reload, recovered protection and no page errors.
No production code, account fixture, storage transaction or remote side effect
was changed for this correction. The parent owns integrated execution against
the final rebuilt source. The parent's targeted run subsequently confirmed this
unchanged production writer case PASS in 2.6s on actual HTTPS/two-node resources
with matching full Chromium; the parent retains its integrated JSON at
`.local/nonrelease-20261008-targets/target-baseline.json`.

## Compose immediate reload: synchronous nested commit repair

The real Compose immediate-reload case still failed in the parent's targeted
baseline (10.6s). A new recovery unit case confirmed a production journal defect:
Vue `flush:'sync'` project watchers can call `RecoveryService.commit` during an
outer `applyEntry`. Before the outer entry advances `state.revision`, the nested
entry receives the same revision. It enters the journal first; replay restores
the nested draft metadata but skips the outer project mutation as already
applied. Earlier immediate IDB commits could conceal this defect; the new bounded
background write delay exposes the ordinary journal-only refresh boundary.

The baseline unit failed with `project:null` instead of the complete selected
project, while retaining the nested Compose draft identity. It had no source
database connection at all, and copied the immediate journal before any await,
so database timing could not hide that loss.

`RecoveryService.commit` now captures nested mutations in a synchronous queue.
The outer entry finishes application and immediate journal protection before
queued entries receive their own increasing revisions and are applied/protected.
The entire queue drains before the outer commit returns; protection does not
wait for a microtask or database write. An application error clears remaining
queued work before propagating rather than replaying it in an unrelated commit.
No Docker/Compose component, remote action or latency oracle changed.

The new case preserves complete project and draft body/save metadata, checks
exact journal revisions `[1,2,3]`, and restores from a captured synchronous
journal using an initially empty database. It failed before the repair and
passed after it. Final `npm test` passed all 23 files / 122 tests (1.80s), and
`npm run check` passed. No distribution bundle was rebuilt by this subagent;
the parent will validate Compose again against the integrated source.

## Historical schema injection: isolate the offline database boundary

The parent's final full real run initially failed the migration/unsupported
snapshot case: its recovery-failure context showed the protected export button
and the known test draft, while waiting 5000ms for the schema-999 warning. The
previous complete run passed this case. The test modified IDB pointer snapshots
inside a still-active product document and immediately reloaded, without
isolating its asynchronous background writer. It also treated synchronous
journal protection as proof that a complete draft snapshot was already in IDB.
Those test prerequisites did not establish a reliable offline schema injection
boundary. A particular background overwrite was not independently traced.

Only `web/tests/real/recovery-failure.spec.ts` changed. Each historical-schema
injection now first confirms the known draft in native IDB, then navigates to
the real same-origin read-only `/healthz` JSON document (HTTP 200). That ends the
old product document and its owned Worker before the native read/write
transaction changes schemas 0 and 999. The completed transaction must change
at least one current pointer snapshot. Returning to the product still verifies
the same schema-0 migration/body, schema-999 warning, retained snapshot and exact
draft export, subsequent reload warning and repeated retained export assertions.
No production Worker was selectively disabled, and no app status, assertion,
distribution bundle or remote side effect was modified.

The targeted case ran in the existing local Playwright container with the final
real fixture credentials and actual `https://127.0.0.1:9443` backend. Result:
1/1 PASS, 4190ms case / 5597.81ms run, exit 0; no skips, flakes or unexpected
results. `npm run check` and `git diff --check` passed. The independent target
JSON is `.local/nonrelease-20261008-targets/schema-offline-confirmed.json`.
This targeted result does not erase or relabel the parent's initial complete
run failure.

The container inherited `PLAYWRIGHT_JSON_OUTPUT_FILE` from the still-running
main mock runner. The target temporarily wrote its single-case JSON to that
shared destination despite its explicit reporter/output arguments. This was
reported immediately and the target JSON copied to its independent path before
the main runner completes. The parent will confirm the main runner's final
full-suite stats, which replace that interim result; future reused-container
commands must override `PLAYWRIGHT_JSON_OUTPUT_FILE` explicitly. No second
container, full-suite rerun or additional fixture was started by this subagent.
