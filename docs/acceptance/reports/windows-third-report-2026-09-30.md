# Third Windows report

User attachment `Blora-Windows-Report.zip`, SHA-256
`32fcd3bbdf32b0f28ea63763b3ef0e19e29f3fcd33276d72b8cdfed764de74c6`.
Checkout `d7e5ef712c9326af3a602b8a7350af0b1aae823d`, remaining mode,
Windows NT 10.0.26200 amd64 / PowerShell 5.1.26100.9168.
Run completed 2026-09-30 13:43:41–14:27:59 UTC. No fatal runner error.
Raw logs with private machine paths are not published.

| Suite | Pass | Fail | Skip |
|---|---:|---:|---:|
| Chromium | 120 | 0 | 0 |
| Firefox | 116 | 3 | 1 |
| WebKit | 112 | 7 | 1 |
| Go all-package test events (includes subtests) | 330 | 22 | 15 |
| Native selection test events (includes helpers) | 21 | 1 | 0 |

Nine of ten required named native checks passed; ConPTY Unicode/resize/close
still failed. Native event totals include helpers and are not the required-check
count. Race testing remains BLOCKED because GCC is absent. Browser native IME
skips in Firefox/WebKit are explicit capability gaps, not passes.

## Newly confirmed

- Go detection, SDK default and versioned WASI packages now succeed.
- Chromium browser suite passes completely, including the reference extension.
- Compose output checkpoint-before-exit and checkpoint-failure-stop tests pass;
  the prior runtime invalid-handle crash did not recur in this run.
- Both container log archive bound/gap and purge tests pass.
- Firefox's font-size rounding assertion no longer fails.

## Still failing

- Directory metadata still reports Access denied in the direct metadata test,
  backup roundtrip and directory transfer tests. READ_ATTRIBUTES alone was not
  sufficient; the Windows metadata correction is **not validated as fixed**.
- Filesystem upload/merge Unix permission assertions, locked-root rename and
  ZIP expansion fixture remain unresolved.
- Several Master integration fixtures still invoke `/bin/sh` on Windows.
- Protocol bulk deadline test fails during setup's first write, before its
  intended rate-wait assertion.
- ConPTY marker is still missing; no root cause proved.
- Firefox: editor undo/move/reload, Unicode/multicursor restore, reference note
  restoration. The extension package exists now, exposing a different failure.
- WebKit: account recovery timeout, reduced-motion click timeout, editor undo,
  history budget, multicursor focus timeout, extension error text (`Load failed`
  versus expected `fetch`), terminal scenario timeout. Not all are proven product
  bugs; none are dismissed as environment-only or counted as passed.

Next work should fix/reproduce these locally where possible before asking for
another long Windows run. Preserve strict assertions and distinguish equivalent
platform tests from native capability gaps. This report is not full acceptance.
