# Shared Monaco view regression closeout — 2026-10-08

## Failure and root cause

The first full real-backend run, `.local/functional-20261008T122417-d9RIOf`, failed `tests/real/desktop.spec.ts:156` at its first 200-line Chinese `keyboard.insertText` operation. The 60-second deadline expired before the second editor window was created. This is a production input/recovery cost regression, not an obsolete independent-cursor assertion.

A fresh isolated mock session on port 18574 with the official Playwright 1.63 image and matching Chromium 1243 reproduced the same timeout with the original text and original subsequent assertions. No real credentials or fixture workspace was required. The first failing result is retained in `web/.local/editor-view-control-20261008.json`.

Native CPU profiling and an interrupt during the stalled input identified Vue `traverse`, dependency tracking and dependency cleanup as the dominant hot functions. The captured stack runs through `RecoveryService.commitEntry`, `applyEntry`, `history.push`, and synchronous deep-watch traversal. Every intermediate history/cursor/body/budget mutation repeatedly walked the growing retained edit history while Monaco delivered input. The safe CPU summaries are attached to `web/.local/editor-view-stack-20261008.json`; they contain function names and source locations, not workspace data.

An initial EditContext hypothesis was rejected: removing EditContext only in an isolated diagnostic browser still timed out at 60 seconds (`editor-view-textarea-control-20261008.json`). No production EditContext option, browser renderer, or editor input path was changed. Temporary diagnostic flags and instrumentation were removed from the permanent test.

## Production fix

`RecoveryService` pauses only its internal cold-branch observers during one journal `applyEntry`, then resumes them synchronously while `applying` remains true. Each observer rebinds against the completed transaction instead of traversing every intermediate mutation. External reactive metadata still invalidates the worker mirror; immediate session-storage protection, revision ordering, nested commit serialization and worker fallback remain intact.

Committed editor-history steps now belong to recovery as copied immutable values. Their forward/reverse operations and arrays are frozen, and their non-enumerable Vue marker prevents deep observers from rereading old offset/text pairs. The owning history array, cursor, document body and metadata remain reactive and mutable. Workspace migration and copied complete drafts rebuild this ownership after native/JSON cloning. JSON and native snapshot payloads retain the existing schema and values.

The pause-only intermediate fix passed the original complete scenario in 56.6 seconds, with its insertion alone taking 42.4 seconds. That result had insufficient margin, so immutable owned history was added instead of increasing the timeout.

## Verification

- `npm run test -- --run`: 123 tests passed in 23 files, including immediate recovery, nested commits, terminal delta ownership, mixed external metadata and editor history budget guards.
- `npm run check`: passed.
- The new ownership guard verifies caller data remains mutable, committed operations resist modification, native/JSON clones preserve payloads, migrated history remains immutable, and later append/undo/document metadata changes still work.
- `tests/browser/editor-view-recovery.spec.ts` mirrors the real test's entire original 200-line input, shared draft identity, unequal cursor and scroll states, reload restoration, and insertion into `独立视图行20定位插入0`. Its 60-second timeout remains unchanged, and it also requires zero page errors.
- Final Chromium mock checks include that complete scenario plus save-as/reload undo, split views, multi-cursor editing and history budget across two fresh repetitions: **10/10 passed**, no skip or flaky result, total 126.7 seconds. Evidence: `web/.local/editor-view-owned-fix-20261008.json`.
- The complete original 200-line scenarios took **18.398 / 19.014 seconds**; native insertion itself took **10.548 / 11.281 seconds**, versus 42.438 seconds with the observer-only intermediate fix. This provides over 40 seconds of margin below the unchanged 60-second scenario deadline, but does not claim sub-second or a few-second native bulk insertion.
- After complete-draft copy ownership normalization, recovery units passed **28/28** and TypeScript checking passed again. Port 18574 has no listening service after cleanup.

## Limits

This isolated browser result proves the frontend/recovery fix on matching Chromium with mock resource APIs. Final compiled-bundle real-backend acceptance is owned by the parent integration run; no build, credential output, release, commit or push was performed by this task. Remaining insertion cost is reported honestly: its likely per-character Monaco content-event/journal contribution requires a separate measured investigation before attributing or changing input transaction semantics.

The parent accepted this bounded repair and ended further editor-history architecture work: it removes the reproduced >60-second failure and preserves the original functional assertions with two passing repetitions. These measurements do **not** establish a few-second bulk insertion result or a 100-millisecond ordinary editor interaction guarantee. All profile/EditContext/corner control flags were checked absent from production source and permanent browser tests; their retained failure/CPU evidence exists only under ignored local test-output directories. Final source whitespace checks passed, and the dedicated browser service was stopped.

## WebKit mock boundary follow-up

The parent corrected the test's Home/End input to the existing platform-aware `pressEditorKey` helper. This is a test keyboard-platform correction, not a product cursor/scroll fix. All original shared-view behavior then passed on WebKit, but one final run recorded a task-summary native fetch diagnostic during reload and failed the unchanged zero-page-error assertion (`.local/nonrelease-20261008-final-webkit-fixed.json`).

Trace time 44787.692 contains the native text `Fetch API cannot load .../api/v1/tasks/summary due to access control checks`, 296 milliseconds after reload began. Playwright 1.63's WebKit `_onConsoleMessage` maps any error-level JavaScript console diagnostic into `pageError`; this mapping alone does not prove an uncaught promise. The product fetch call is inside `api.request`'s catch, and `readDocument` also catches its read. One bounded original-scenario control recorded both native DOM `error` and `unhandledrejection` channels across reload. It passed with both DOM channels and `pageErrors` empty in 39.05 seconds (`.local/editor-webkit-native-control-20261008.json`). The transient diagnostic did not reproduce, so a particular native cancellation/CORS cause is **not claimed proven**.

Only the independent test's management transport boundary was corrected: it now uses the existing management fixture to replace the desktop event stream as well as HTTP reads, keeps HTTP routing owned by the context across refresh as the existing navigation guards do, and returns the actual task-summary schema `{active:0,states:{}}`. The 200 lines, all cursor/scroll/shared-draft/reload/insertion assertions and `errors=[]` remain. There is no error-message filter, relaxed deadline, production source change, dependency change or bundle rebuild.

The single permitted post-correction WebKit run passed in 38.145 seconds (41.397 seconds including runner setup), 1/1 expected, zero skipped/flaky/unexpected results. Evidence: `.local/editor-webkit-management-boundary-20261008.json`. The temporary native-channel control code/storage marker was removed; the existing parent-owned container was retained, no new container/network interface was created, and port 18687 was confirmed stopped. This bounded pass does not claim every possible navigation/network failure is impossible; the final full integration run remains the parent's responsibility.
