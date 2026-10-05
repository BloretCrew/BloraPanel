# Frozen-material layout isolation diagnostic

Date: 2026-10-05. Source baseline: `f183b1ad0b55fcd464fae0b58fe4da3f783df0d2`.

This is a new local diagnostic for the remaining E08 performance issue, not a passing acceptance run. No product source, appearance, material, load requirement, or 50 ms threshold was changed. The previously rejected blur-area clipping and SwiftShader experiments were not repeated.

## Candidate and visual check

The diagnostic opened eight actual Vue application windows at 1440 × 960 in system Chromium, with synthetic read-only API responses. It compared the unchanged windows against `.app-window { contain: layout style }`, a layout/style isolation candidate distinct from the earlier terminal-row containment and paint/layer experiments.

The baseline computed material remained `blur(24px) saturate(1.05)`, background opacity 0.83. Transitions/animations and the clock were held constant for the diagnostic. Two baseline screenshots and one candidate screenshot were pixel-identical; all three have SHA256 `6f3986d49c569b77c480362f0535b8e63a84f229f77a395fee0e01a3e700477d`. This establishes appearance equivalence only for the captured layout, not every application interaction or theme.

## Pointer timing

Each round issued 60 actual mouse moves during a title-bar drag. Latency used the same event timestamp to double-animation-frame boundary as the existing performance test. The order was baseline, candidate, candidate, baseline; the original window geometry was restored before each round.

| Round | Mode | Completed samples | p95 | Maximum |
| --- | --- | ---: | ---: | ---: |
| 1 | Baseline | 59 | 1086.7 ms | 1394.8 ms |
| 2 | Layout/style isolation | 59 | 952.5 ms | 976.9 ms |
| 3 | Layout/style isolation | 60 | 922.7 ms | 948.5 ms |
| 4 | Baseline | 60 | 1005.3 ms | 1122.1 ms |

The first baseline may overlap the end of an independent browser regression, so it is not a clean performance baseline. The later comparison suggests a modest timing difference, but does not establish a robust production improvement. Both candidate rounds remain more than 18 times the 50 ms target. There were no live PTYs, verified transfer or 10,000-entry directory workload in this diagnostic; its samples cannot replace the full E08 scenario.

## Decision and remaining limit

The candidate was **not adopted**. A single-layout microcomparison does not justify adding layout-containment semantics to all application windows; fixed descendants, counters and additional UI combinations would need their own validation. It does not solve the dominant software-compositing cost already demonstrated by the [render trace](local-diagnostics-2026-09-29.md).

The current host still has no `/dev/dri` device. A separate compositor-state query used the same system Chromium 151.0.7922.71 launch settings and returned `gpu_compositing=disabled_software`, `rasterization=disabled_software`; it does not turn this diagnostic into hardware-accelerated acceptance evidence. The original complete-load Chromium/Firefox/WebKit results of 1003.9/1108/545 ms remain failed, and E08's p95 ≤ 50 ms target remains unchanged. Hardware measurements are optional under the user's revised environment scope; they are not another mandatory request to provide a device.

Local diagnostic artifacts are private under `/tmp/blora-r1-layout-isolation` and `web/.local/layout-isolation-probe.mjs`; the result JSON SHA256 is `18c2d58da03681b2bd3cfc05035e17c9c14c610f21ea8eba6af335327e58a6f5`. The initial sandbox-only Vite launch was denied loopback listen access, then the authorized loopback server ran with escalation. Its temporary port was 18573. The probe/browser exited successfully and the owned Vite process was stopped; no test server or browser from this diagnostic remains running.
