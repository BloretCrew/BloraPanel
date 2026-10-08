# E08 bounded closeout — 2026-10-08

This is the source-specific bounded performance record. The subsequent
[Chromium corner repair](chromium-corner-closeout-2026-10-08.md) and
[editor regression repair](editor-closeout-regression-2026-10-08.md) resolve
later closeout findings without rewriting these measurements or earlier failures.
See the current acceptance matrix for the integrated non-release disposition.

The user explicitly changed the scope on 2026-10-08: accept feedback around 100ms in the rare mixed-window stress scenario, make only a few final attempts, and finish this optimization round. The original 50ms target remains recorded separately. This is a bounded performance closeout, not a claim that every browser meets 50ms or that the project has been released.

## Retained implementation

- Data-only background recovery snapshots and bounded terminal deltas; immediate recovery journals, protected output acknowledgements and reload recovery remain intact.
- Grouped terminal transport and parallel native terminal parsing/background checkpoint preparation, with both completions required before acknowledgement.
- Synchronous native window movement, conservative software-renderer positioning and browser-owned application/text/glyph rendering.
- Cached wallpaper material, Dock optics and window shadows; covered native regions avoid unnecessary painting. Fractional scale and unsupported cases retain the original native fallback.
- The compact select clipping was a real WebKit native-control appearance issue. Explicit control appearance fixes it while retaining the requested original compact rounded fields. Evidence: [corrected controls](../../../web/.local/select-review-after/webkit-light-controls.png).

Only wallpaper material and optical decoration are cached. Application contents and terminal glyphs are not replaced with snapshots. No host GPU settings, priorities, services or credentials were changed.

## Final bounded attempt

One final candidate was independently retested: the existing native rectangular body clip with a flattened wallpaper layer. Its earlier 79ms result did not repeat; run 1118 measured **139ms**. The existing product baseline subsequently measured 132ms, so this candidate offers no demonstrated benefit and was not adopted. Earlier partial-body, original-transform-control and automatic-row-skip experiments also remain outside production; their failures are preserved in the [architecture history](e08-render-architecture-2026-10-06.md).

The optimization round stops here; no further architecture experiments are scheduled. The remaining runs verified the unchanged product, including one independent WebKit repetition to show variability rather than selecting its lowest result.

## Same-source verification

Source fingerprint (545 files): `4c8db790cc97dcd6f1ba69b0d8007e40b0044f781cde9c9c9e87674c5ca1de9f`.

Served production bundle fingerprint (169 files): `5a04e608c1958244f78594fc7539be40a95b19b8b5ff98cac486091c0021d1d3`.

Environment: local Linux, Ryzen 7 5800H, 12 available logical CPUs, official Playwright `v1.63.0-noble` container. Browser defaults remain unchanged; renderer labels are recorded metadata, not proof of physical GPU availability. Chromium reports SwiftShader. Linux WebKit's generic Apple GPU label follows the conservative native fallback.

All final product runs use eight real windows, two real PTYs configured for 1KiB every 10ms, 10,000 directory entries and a verified 16MiB cross-node transfer. Measurement uses the first 120 continuous trusted pointer moves and event timestamp through two animation frames. No diagnostic rendering flags, reduced workload, extra soak or altered acceptance threshold are used.

| Browser / run | Pointer p95 | Maximum | Frames over 50ms | Maximum unacknowledged output | Verified transfer | Original 50ms target |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Chromium 153 / 1124 | 38.8ms | 43.7ms | 0 / 152 | 86,175 bytes | 12.931s | PASS |
| Firefox 155 / 1122 | 66ms | 92ms | 5 / 130 | 76,905 bytes | 13.901s | FAIL |
| WebKit 26.6 / 1120 | 132ms | 170ms | 12 / 34 | 86,265 bytes | 13.123s | FAIL |
| WebKit 26.6 / 1126 | 95ms | 113ms | 3 / 26 | 87,076 bytes | 27.825s | FAIL |

The earlier same-source WebKit run 1082 measured 89ms, so the three recorded product runs span 89–132ms, with the final repetition at 95ms. The 132ms run remains visible: the product is not demonstrated to stay below 100ms on every run. The user's approximate acceptance range is a scope decision; it is not an additional automatically passing assertion.

All four final product runs report `renderExperiment=baseline`, `terminalRendererExperiment=baseline`, no optional clipping/retention/positioning controls, ready material cache, worker recovery persistence and instances still running. Actual native terminal rendering remains `default/default`; Firefox and WebKit use eight native software window offsets, while Chromium keeps its existing transform path. No hidden terminal row adapter is installed.

Evidence: [Chromium](../../../web/.local/e08-20261008-1124-final-chromium-original.json), [Firefox](../../../web/.local/e08-20261008-1122-final-firefox-original.json), [WebKit](../../../web/.local/e08-20261008-1120-final-webkit-original.json), [WebKit repetition](../../../web/.local/e08-20261008-1126-final-webkit-repeat.json), [rejected final candidate](../../../web/.local/e08-20261008-1118-final-flat-body-webkit.json). Real reports disable browser snapshots, trace and screenshots; no credentials or raw process logs are published.

## Checks and limitations

- Production build and 121 unit tests passed for the retained source (1078). Final type checking (1116) and whitespace validation passed.
- WebKit's 21 native pixel, geometry, pointer, output/protected-ACK and recovery checks passed for this source (1080). Protocol/master race checks passed (0845).
- Final product runs complete destination verification before applying the unchanged 50ms latency assertion. Firefox and both final WebKit runs failed only that latency assertion (line 603), not the data or protection guards. There were no skipped or flaky cases.
- A separate added Chromium DPR1 static screenshot oracle cannot settle a tiny corner region, including when restoring the old positioning implementation. The failure remains recorded; it is not weakened, deleted, claimed fixed or counted as passing. This is a known diagnostic limitation outside the original functional checks.
- These are local Linux browser results, not new Windows, physical Safari or released-package certification. Existing Windows evidence is retained; no additional Windows execution is requested for this bounded optimization round.

## Final disposition

The bounded performance optimization round is finished under the user's changed scope. The last candidate was rejected; the verified implementation is retained. Original 50ms acceptance is satisfied by Chromium and remains unmet in the final Firefox/WebKit runs. WebKit's final 95ms result and 132ms variation are both retained; no stable ≤100ms guarantee is claimed.

Each of the five fresh local fixtures (1118, 1120, 1122, 1124, 1126) was stopped after its browser exited, with cleanup exit 0. A final process check found no remaining Playwright, Node, Docker test client or development fixture process in this workspace. No commit, push, public deployment or release was performed. R2/R3/R4 history remains unchanged and release work R5 stays deferred at the user's request. Further pursuit of a universal 50ms bound is a future performance improvement, not an active task in this round. The separate static Chromium screenshot instability remains documented as an unresolved diagnostic, not a passing result.
