# Second beta preparation and package acceptance — 2026-10-09

## Post-freeze publication verification

Published [v0.1.0-beta.2](https://github.com/BloretCrew/BloraPanel/releases/tag/v0.1.0-beta.2)
at `2026-10-09T11:58:51Z`, release ID `407867988`, `draft=false`,
`prerelease=true`, publication used `--latest=false`. Fixed annotated tag/source:
`96bd198bb90e105d91a623fdbdb679ca8d2da532`. All ten uploaded attachments have
exact sizes and GitHub SHA-256 digests matching the local frozen artifacts.

After source freeze, `release-verify.py` passed 1,640 component manifest records,
three identical 171-file frontend copies, all 1,377 tracked source files,
GPL/notices, binary/CORE compatibility and exclusion of untracked documentation.
The final no-argument packaged smoke passed with this exact version/revision;
the final extracted two-node/SDK smoke, real beta.1 seeded-state upgrade,
stopped snapshot restore and compatible beta.1 snapshot rollback all passed.
Second packaging reproduced the six component archives, CORE metadata and
initial checksum list byte-for-byte. The final nine-entry SHA256SUMS additionally
covers source and provenance; all entries passed checksum verification.

This section was added in a documentation-only follow-up; it does not rewrite
the frozen tag or archives. Final build/provenance details are attached to the
release. Owned test services are stopped. No production deployment occurred.

Version: `v0.1.0-beta.2`, prerelease, GPL-3.0-only. The user authorized publishing
the current version after moving the development checkout to
`/data/projects/blora-panel`. Production installation at
`/data/instances/blora-panel` is **guidance only**; no production files, services
or nodes are installed, restarted or updated by this release task.

The 48-file handoff snapshot matched after migration. The internal conversation
handoff stays local and is excluded from product/source archives. Beta.1 remains
unchanged. The final source revision, artifact digests and publication checks
are recorded in `RELEASE-PROVENANCE.json` accompanying the release; this report
is frozen before final packaging/publication and does not claim those later
steps have already run.

## Source and prepublication checks

| Check | Result / scope |
| --- | --- |
| `make check test` | PASS, complete Go static check and race suite, exit 0; Master 310.933s, Daemon 34.054s, coreupdate 28.008s; opt-in skipped environments are not runtime PASS |
| Frontend locked installation/check/test/build | PASS; 123 tests in 23 files; production build completed, existing chunk warning retained |
| `make sdk` | PASS, SDK and three reference package versions build; no dependency version migration |
| Chromium `core-updates.spec.ts` | PASS, 2/2, 14.3s; mock-client boundary only |
| Real Linux HTTPS Release mechanism drill | PASS, 3/3; actual Master/Daemon switch, unchanged instance PID birth/runId and heartbeat, receipt replay, corrupt download refusal |
| Linux/Windows release builds | PASS; version/revision injected into both components, Windows runtime not executed |
| Extracted no-argument HTTP programs | PASS on prepublication archive; sibling JSON, unrelated cwd, installation-relative state, bundled Web/login, node updater, identity reuse after removing initial password/enrollment files |
| Extracted TLS/SDK/two-node upgrade | PASS on prepublication archive; beta.1 seeded state → beta.2, stopped-state snapshot restore and compatible beta.1 snapshot rollback; login, resource bytes, node identities, read-only grants and task receipts retained |

Final archive versions of the read-only manifest verifier and packaged smoke
checks are rerun after source freeze; their actual outcomes belong in release
provenance rather than being inferred from these prepublication passes.

## Packaging defect found and corrected

The previous documentation tree walker included 211 ignored local visual-review
files (85,088,445 bytes) in each core archive. They were not committed build
inputs. Packaging now includes only tracked documentation, and release
verification rejects any untracked `docs/` member. Generated frontend and SDK
inputs remain explicit. New candidate core archives contain zero untracked docs;
the Linux Master compressed size falls from about 96MiB to about 16MiB.

The new no-argument smoke runner initially omitted Idempotency-Key for node
enrollment and correctly received HTTP 400. The test request was corrected to
use a unique key; server requirements were not weakened. The corrected runner
passed against the same candidate programs. The original failure is retained
here rather than described as an initial all-green run.

## Preserved boundaries

- New Windows launcher/online-update/Job keeper continuity remains NOT RUN on a
  device. Previous Windows reports retain their own fixed commit and scope.
- Updating preserves instances, not every interactive terminal or upload
  connection. Incompatible storage fingerprints or peer protocols block online
  activation; initial adoption from beta.1 is a manual maintenance installation.
- Source-specific E08 results and the 200-line editor limitation are unchanged;
  no new universal latency or platform claim is made.
- Checksums express input/asset correspondence, not independent code signing.
- Original beta.1 archives may contain historical local review evidence; this
  release fixes the packaging inputs without silently rewriting that tag or its
  attachments. This finding is not evidence of a credentials leak.
- Only private, task-owned test roots and services are used. Smoke credentials,
  databases and raw process logs are not uploaded.
