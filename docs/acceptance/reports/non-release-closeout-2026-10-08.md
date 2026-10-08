# Non-release source closeout — 2026-10-08

The user authorized completion of three non-release tasks: current-source regression, code/documentation cleanup and GitHub handoff, and the Chromium rounded-corner repair. Final release stays deferred. This work does not reopen E08 optimization or certify every original environment.

## Scope and source identity

The [October 5 scope decision](scope-review-2026-10-05.md), [ordinary Windows 14-check evidence](windows-device-passed-2026-10-06.md), and [bounded E08 closeout](e08-bounded-closeout-2026-10-08.md) remain effective. Waived physical failures and extra host/browser/device combinations are unexecuted, not PASS. Historical platform and 24-hour reports retain their original source and environment.

Source identity is computed over unique non-ignored source/test/SDK/script paths in cmd, internal, web/src, web/tests, sdk and scripts, plus Go/Make/frontend dependency and browser configuration inputs. For each sorted relative path, hash UTF-8 path, NUL, exact file bytes and NUL with SHA-256. The bundle identity applies the same algorithm to every regular file under web/dist. Documentation/private evidence/generated artifacts are excluded from the source identity. The final numeric identities and counts are below.

## Repairs confirmed during closeout

1. **Chromium rounded material plane.** The separately retained wallpaper material could vary at a resized rounded corner although geometry and samples were unchanged. Only Chromium's body wallpaper pseudo-element stops forcing independent retention. Original six scenes, both physical DPRs, eight stable captures, full-screen RGBA0 and node identity checks remain. Initial 3/3 and two-repeat optical 24/24 checks passed. See the [corner report](chromium-corner-closeout-2026-10-08.md).
2. **Recovery worker runtime fallback.** A transport/lifetime failure after initialization now selects the complete foreground atomic transaction. Actual native storage failures remain failures; writer generation and identity prevent stopped/logged-out writers from republishing. Real transaction-abort injection now operates inside the compiled snapshot Worker, rather than a page-only prototype. See the [recovery review](non-release-reconciliation-review-2026-10-08.md).
3. **Nested recovery revision ordering.** Synchronous Compose observers formerly committed inside an unfinished outer mutation, producing duplicate revisions and losing the project on immediate reload. Owned queued mutations drain synchronously after the outer protection boundary. A focused unit reproduction and unchanged real Compose creation/restore assertions pass; remote effects are not replayed.
4. **Shared Monaco input hang.** Repeated deep traversal of growing history during one journal apply caused a >60-second 200-line input timeout. Cold observers rebind once synchronously per apply, and copied owned history steps are immutable raw values while history arrays/body/cursor/metadata stay reactive. No EditContext workaround, reduced text or timeout increase was adopted. Two independent complete editor repetitions passed; 200-line native insertion still takes about 10–11 seconds. This is a retained limitation, not a new claim of universal 100ms feedback. See the [editor report](editor-closeout-regression-2026-10-08.md).

The first complete real run passed 46/51 and failed five, with zero skip/flaky and fixture exit 0. Its private directory is .local/functional-20261008T122417-d9RIOf. Besides the two product regressions above, the matching headless shell returned Notification.permission=denied despite a grant, the old abort injector targeted the wrong realm, and the upload-close fixture lacked its actual 10,000-file/16MiB sources. The final runner uses matching full Chrome, the real Worker injector and the --performance fixture. No permission return, notification API or storage success is fabricated. Four unchanged real targeted cases subsequently passed 4/4 in 60.584 seconds; their owned 9444 fixture stopped with exit 0.

5. **Terminal toolbar native paint ownership.** A partial viewport switch rerasterized three unrelated toolbar button-edge pixels in Chromium. A stable screenshot prerequisite alone still failed. Retaining only Chromium's unchanged terminal toolbar repaired the paired full-screen comparison; two DPRs × two runs passed all 24 screenshot pairs, followed by seven related native optical checks. No terminal renderer, contents, geometry or output/recovery semantics changed. See the [toolbar report](terminal-toolbar-paint-closeout-2026-10-08.md).

## Current-source validation

| Check | Result and evidence |
| --- | --- |
| Backend race | `go test -race -p 1 ./... -count=1`: exit 0, 18 tested packages; opt-in environment skips remain unverified |
| Build and platform compilation | `make check build windows`: exit 0; frontend production build and final TypeScript check exit 0 |
| Frontend units / SDK | 123/123 in 23 files; standalone SDK and three reference versions build successfully |
| Complete real functional run | 49 PASS / 2 FAIL / 0 skip / 0 flaky, 974.087s; owned fixture cleanup exit 0 |
| Real remaining-case checks | Unsupported snapshot/migration 1/1 PASS (5.598s); original browser-close/extract/resume 1/1 PASS (58.114s), owned fixture cleanup exit 0 |
| Complete Chromium browser run before experiment separation | 175 PASS / 6 FAIL / 0 skip / 0 flaky, 1119.422s, 181 total; includes eight explicitly installed unadopted prototype cases |
| Production optical independent checks | Original partially covered shadow + Dock scenarios 5/5 PASS, 31.227s; unchanged pixel assertions |
| Cross-engine recovery worker guards | Chromium, Firefox and WebKit each six PASS; native storage error/fallback/stop/reload checks retained |
| Complete independent shared editor guard | Matching Chromium PASS with real backend and mock boundary; Firefox PASS (51.806s runner); WebKit PASS (41.397s runner); original 200 lines, reload/cursor/scroll and zero-error assertions retained |
| Final terminal paint ownership checks | Production two DPRs × two repetitions 4/4 PASS (55.459s); related window/shadow/Dock/decode 7/7 PASS (59.987s) |
| Native window-position repair | Initial 3/3 and repeated 24/24 optical checks PASS; both DPR guards also passed in the complete Chromium run |
| Scope registration / source hygiene | Default product suite lists 173 cases in 41 files; explicit experiment config lists 52 cases in 24 files; whitespace, shell syntax and secret-pattern checks pass |

The 51 real functional cases are covered by the complete run's 49 passing cases
and the two independent passing checks. **No single complete 51/51 green run is
claimed.** The retained first and final complete runs both returned failure.
The schema injector now leaves the running product document before committing
the native schema-0/999 snapshots, preventing an active writer from invalidating
its offline fixture setup; migration, retention/export and second-reload
assertions remain. The upload-close case passed independently using the same
90-second bound, 10,000-file extraction, 16MiB original file and complete
server-side hash check. Its full-suite extraction timeout occurred alongside
other heavy browser runs; a specific filesystem root cause is not claimed.

The complete 181-case Chromium run includes eight cases (four files, both
native DPRs) that explicitly install unadopted bar/alpha/grouped-shadow/terminal-row
adapters. All four files now live in the explicit experiment configuration,
with their original comparisons and failure evidence preserved. This moves
the prototype category, not selected failures. The default product list is
173. Its complete-run projection was 169 PASS and four FAIL: offscreen
navigation (`ERR_NETWORK_CHANGED` before application startup), partial
shadow, Dock, and partial terminal painting. Offscreen navigation passed two
independent unchanged checks; partial shadow and Dock passed their complete
independent scenarios with zero pixel differences. The terminal case's three unchanged toolbar-edge pixels still differed after stable screenshots; isolated Chromium toolbar retention repaired the production path. Both DPRs then passed twice (4/4), followed by all seven related native optical checks (7/7). See the [terminal toolbar report](terminal-toolbar-paint-closeout-2026-10-08.md).

These independent passes do not prove that every future complete-suite native
paint or host-network fluctuation is impossible. The two unadopted candidates
that failed are not declared equivalent or shipped. No product guard was
moved, no pixel threshold was raised, and no scene, failure record or required
functional contract was removed.

The WebKit editor's earlier complete scenario passed its functional assertions
but failed the zero-error assertion with a native fetch diagnostic during
reload. A bounded native error-channel control did not reproduce it; no CORS
or cancellation cause is claimed proven. Its independent mock now uses the
existing full management transport fixture, context-owned HTTP routes across
reload, the actual task-summary schema and platform-aware keyboard bindings.
The original zero-error assertion remains, and the corrected boundary passed
on all three engines. A competing Chromium dev-server run also exceeded its
post-reload editor readiness limit; the isolated original scenario passed,
without increasing that deadline.

Private structured evidence locations:

- `.local/nonrelease-20261008-go-race.log`, `.local/nonrelease-20261008-check-build-authorized.log`
- `.local/nonrelease-20261008-final-web-build.log`, `.local/nonrelease-20261008-final-web-unit.log`, `.local/nonrelease-20261008-final-typecheck.log`, `.local/nonrelease-20261008-sdk.log`
- `.local/functional-20261008T130351-2HMpiF/real-functional.json`
- `.local/nonrelease-20261008-targets/schema-offline-confirmed.json`
- `.local/nonrelease-20261008-upload/upload-close.json`
- `.local/nonrelease-20261008-final-mock.json`, `.local/nonrelease-20261008-final-independent.json`
- `.local/optics-closeout-initial-20261008.json`
- `.local/nonrelease-20261008-final-firefox.json`, `.local/nonrelease-20261008-final-webkit-fixed.json`
- `.local/nonrelease-20261008-final-editor-firefox.json`, `.local/editor-webkit-management-boundary-20261008.json`

Final selected source identity (1052 files): `8f6456948c19cf46ed36b902eb8a3e356634d5a37dabdaed1b6877174adda3e5`.

Final rebuilt production Web identity (169 files): `cad1403c1a13210ffc9fba1cbce73fe84218ac095e09a8520e3029c1a6ebb624`.

Parent Git revision before this handoff: `ae78c9ebe6c5ef5e62636d51e7dccbf5555ba4fe`. The associated handoff commit contains these source inputs; its identity is verified against GitHub main after push. The final terminal-only CSS rule followed the complete suites and was verified by 4 terminal executions and 7 related native optical checks. The final production bundle was rebuilt successfully with TypeScript checking and Vite's 3.27s build. This does not claim the complete suites were rerun on that last CSS-only change.

Go opt-in environment SKIPs are not passed environment checks. The browser mock suite validates frontend behavior, not backend operations. The real suite starts its own HTTPS Master and two Daemons, actual Docker/Compose, real filesystem/PTY/backup/extension resources and 121-point monitoring. It excludes only the separately recorded E08 stress assertion under the user's accepted scope. No other required functional case is skipped.

Commands and reproducible runner are in [LOCAL_REGRESSION](../../operations/LOCAL_REGRESSION.md). The run builds locked dependencies, SDK/reference packages and current production inputs before real acceptance. The matching full browser is /ms-playwright/chromium-1243/chrome-linux64/chrome in mcr.microsoft.com/playwright:v1.63.0-noble. Backend race uses go test -race -p 1 ./... -count=1. Frontend uses npm run build, npm test and TypeScript checks. Complete runs contain the final recovery/runtime repairs; subsequent terminal-only CSS and mock boundary changes were verified by the stated focused checks. Independent and historical runs are labeled rather than presented as one final green execution.

## Documentation and handoff

The current [acceptance matrix](../ACCEPTANCE_MATRIX.md), [progress](../../execution/PROGRESS.md), and [remaining work](../../execution/NON_UI_REMAINING_2026-09-26.md) now distinguish implementation, current verified scope, old evidence, waived environments and deferred release. Complete previous versions live in their archive subdirectories; failed results are retained. The default README remains English with a Chinese link, and both versions point to the current scope and regression guide.

Rejected rendering prototypes retain their strict failure comparisons in the explicit experiment config. They are not imported by production and not counted as default product checks. Ordinary corner tests remain in the default suite; no pixel tolerances were relaxed.

Credentials, databases, raw real process logs and real browser snapshots remain under ignored private output directories. The runner stops only its exact owned container and fixture; resource cleanup is confirmed. No global Docker cleanup, production host mutation, tag, final package or public deployment was performed.

## Remaining work

The three authorized non-release tasks are closed after the recorded GitHub handoff. Final fixed-source release packaging, source/Web/SDK/manifest matching, packaged initialization/backup/restore/compatible upgrade and rollback, license/release decisions and publication remain deferred. Historical release-candidate results do not certify that future final package. Further universal 50ms tuning and faster bulk Monaco paste are future improvements, not active tasks added to this closeout.
