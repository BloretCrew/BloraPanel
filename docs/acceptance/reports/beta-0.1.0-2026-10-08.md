# First beta package acceptance — 2026-10-08

## Identity and release boundary

- Version: `v0.1.0-beta.1`; **prerelease**, not a stable release.
- License: **GPL-3.0-only**. Every component contains the root GPL text and its third-party notices.
- Fixed source and annotated tag: `b5d0287b49b53cafaec94ee04bdd8b115cf767b4`.
- Toolchain: Linux amd64, Go 1.25.9 (`X:nodwarf5`), Node.js 24.15.0, npm 11.12.1.
- The project owner has not personally tested the application. The release remains a beta for initial hands-on evaluation.
- Delivery: [GitHub prerelease](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.1), published `2026-10-08T14:59:23Z`, release ID `406948191`. Final API verification confirms `draft=false`, `prerelease=true`, tag `v0.1.0-beta.1`; publication used `--latest=false`.
- All nine remote attachments have `state=uploaded`; names, exact byte sizes and GitHub SHA-256 digests match the local files. Seven are archives, plus `SHA256SUMS` and `RELEASE-PROVENANCE.json`.

The package inputs were frozen at this commit before the package smoke tests. This report and subsequent documentation-only closeout changes do not change the tag or replace its archives. Packaged progress documents therefore contain the then-current “beta in progress” status; this report and the release's `RELEASE-PROVENANCE.json` record the completed checks.

## Checks on the frozen source and actual archives

| Check | Result and scope |
| --- | --- |
| `make check build windows web sdk` | PASS, exit 0: Go checks, Linux/Windows compilation, frontend/TypeScript, SDK and reference builds |
| Frontend unit suite | PASS, 123 tests in 23 files |
| Dependency audit after locked patches | Zero npm audit findings; DOMPurify 3.4.16 and source-map-js 1.2.2 |
| Chromium editor save/shared-document guards | PASS, 3/3; LF/CRLF saving and original 200-line reload with independent cursor/scroll; no skips/flaky |
| `BLORA_VERSION=v0.1.0-beta.1 make package` | PASS from clean committed source |
| `scripts/release-verify.py` | PASS: six component archives, 2,418 manifest records, root license, clean source revision and payload correspondence |
| Packaged frontend correspondence | PASS: Linux Master, Windows Master and standalone Web contain identical 169-file frontend trees |
| Complete source archive | PASS: all 1,330 tracked files match the frozen source |
| `sha256sum -c SHA256SUMS` | PASS for all seven archives, including source |
| Independent second packaging | PASS: six component archives and their original six-entry checksum file are byte-identical; no overwrite of the frozen release |
| Fresh beta startup and beta's own stopped-state restore | PASS using the extracted release programs, not development binaries |
| RC34 → beta upgrade | PASS: old programs initialize and seed the state; beta opens that same stopped state |
| Beta → compatible RC34 snapshot rollback | PASS against the recorded compatible old snapshot; no arbitrary downgrade promise |
| New beta archives on a Windows device | NOT RUN. Cross-compilation and archive verification do not replace device runtime checks |

The startup/restore checks cover a real HTTPS Master, two actual Daemons, TLS/node identities, administrator/member login, read authorization and denied start authorization, stored resource bytes, and idempotent task receipt identity. The SDK is independently rebuilt after extraction and signs the reference extension. The beta's own restore and compatible old-snapshot rollback were separate runs, so rollback does not substitute for beta restore. All test-owned processes were stopped afterward.

Recorded package smoke commands, both exit 0:

```sh
python3 scripts/package-smoke.py dist/releases/v0.1.0-beta.1 --state-restore
python3 scripts/package-smoke.py dist/releases/v0.1.0-beta.1 --state-restore \
  --upgrade-from dist/releases/development-20260929-rc34 \
  --rollback-release dist/releases/development-20260929-rc34
```

Private smoke evidence remains in `/tmp/blora-release-smoke-35vpgx5_` and `/tmp/blora-release-smoke-s8oix0a7`; reproducibility output is in ignored `.local/beta-repro-20261008`. No credentials, private database/state, raw real-process logs or real browser snapshots are published. Public evidence is the sanitized report, matching source/tag, checksums, per-file manifests and provenance attachment.

## Archive hashes

| Archive | SHA-256 |
| --- | --- |
| `blora-master-v0.1.0-beta.1-linux-amd64.tar.gz` | `0e27006f181a0034e91e2bdc5bd5901d057b3e8e9f9c4b659e780c456cd447f7` |
| `blora-daemon-v0.1.0-beta.1-linux-amd64.tar.gz` | `f274cf4808d150641612f3dd75835d49e49864a936138149ee6e7603a69fe03b` |
| `blora-master-v0.1.0-beta.1-windows-amd64.zip` | `9b303b98e037a6b48441aae4b882faf06726e962de0a07078cbdc7de2155fd5c` |
| `blora-daemon-v0.1.0-beta.1-windows-amd64.zip` | `054b8c0b6f87a878d85a07267e5ca1a8c8b8d85c8886d0ba57ecf9d2fa17f065` |
| `blora-web-v0.1.0-beta.1.tar.gz` | `845bdac94308c453abb424e09fc57651d6ce6107bcddbad5a68a927ab3f115bf` |
| `blora-sdk-v0.1.0-beta.1.tar.gz` | `f36cf522aa3add7081ba95d97596c1d478f955778ee3f13b2aa4a60c9a71cdf4` |
| `blora-source-v0.1.0-beta.1.tar.gz` | `958c8dcf38babb42b5a14aa731bf5b9aff800d8ba6829fd7c55696d3ca731d25` |

Full source-tree identity is `0197cb8d8633b67e3b5db8b3228e05b3510e4d3163ea8b100c572ec1d0b1c2be`; production Web identity is `cad1403c1a13210ffc9fba1cbce73fe84218ac095e09a8520e3029c1a6ebb624`. The tree algorithm hashes sorted UTF-8 repository-relative paths, NUL, exact bytes, NUL; Web uses the `web/dist/` path prefix. These identities and toolchain/check results accompany the release in `RELEASE-PROVENANCE.json`.

## Preserved limits

[Source closeout](non-release-closeout-2026-10-08.md) combines a complete real run of 49 PASS / 2 FAIL with two later original-scenario passes. It does not claim a single 51/51 green run. [E08](e08-bounded-closeout-2026-10-08.md) retains Chromium 38.8ms, Firefox 66ms and WebKit 89–132ms (last 95ms), on their recorded source; no universal 50ms/100ms guarantee. The 200-line bulk-input limitation of about 10–11 seconds remains.

[Windows device results](windows-device-passed-2026-10-06.md) retain their earlier fixed-source identity. Extra physical-failure/platform/deployment combinations remain waived or unverified as recorded in the [scope review](scope-review-2026-10-05.md), not passed. SHA-256 checks provide integrity correspondence, not publisher code-signing. Stable-release readiness and production deployment are not part of this beta publication.
