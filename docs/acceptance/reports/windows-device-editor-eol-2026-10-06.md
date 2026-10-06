# Windows real-device report and editor EOL fix

Received `Blora-Windows-Device-cf83c5ea-20261006-034625234-91b7ff3a.zip`.
Archive SHA-256:
`7f52ec7ce94c0cef6d05e554974fe78cdd15f1d6737be899b12f0e139e250bf4`.
All twelve entries were inspected; the checksum manifest exactly covers the
eleven other entries and every checksum matches. Requested/source commit is
`cf83c5ea39faa6b401d94866a80c85459d89a973`; runner SHA matches that script.
Device platform is `windows/amd64`, run UTC 03:46:25–03:48:01 on 2026-10-06.

Source acquisition and all seven preparation stages passed. The archive download
took 3.990 seconds. Eleven real feature checks passed: initialization, TLS/static
login, daemon enrollment, stopped-instance creation, permission denial, file
save/read, backup/restore, monitor sampling, extension lifecycle, instance
start/stop and terminal WSS execution/reconnect/close. Owned-resource cleanup
also passed. Browser login, resource/file creation and unsaved-draft refresh
completed, but exact saved-content readback failed. Stopped-state process restart
was not executed after that failure. Overall Windows result remains FAILED;
this is twelve passed, one failed, one unexecuted requirement, not fourteen passed.

The previous child-launch/source-acquisition blockage no longer occurs in this
report. Its precise original cause remains undetermined.

## Reproduction and fix

Using current real Master/Daemon and the shipped frontend with Chromium's Windows
user-agent, the browser check reproduced the failure. Safe additional diagnostics
showed `HTTP 200; unexpected-crlf`: the saved body contained CRLF instead of the
declared LF. No credentials or raw document bodies were logged.

Monaco's default EOL depends on browser platform. `documentModel` previously set
CRLF explicitly but assumed LF otherwise, allowing an empty LF document to use
the Windows default. Both initial model creation and history rebuilding now set
either LF or CRLF explicitly from the draft/file format before replaying edits.
Exact byte/body expectations remain unchanged; no newline normalization was
added to the acceptance assertions. No UI appearance or backend change was made.

## Local verification

- `npm run build`: type checking and production build passed, 3,740 modules.
- Focused Windows-UA Chromium editor-save cases: **2/2**, 48.4 seconds. LF and
  CRLF preserve BOM/format and captured save through refresh while retaining newer
  input. These cases use API doubles and are separate from the real workflow.
- Real current Master/Daemon plus production frontend, Windows-UA Chromium:
  **14/14**, exit 0, `/tmp/blora-device-windows-ua-after.json`. The exact LF draft
  survives refresh, saves with matching server readback and survives final refresh;
  stopped-state restart and owned-resource cleanup both passed.
- `node --check web/scripts/device-browser.mjs` and `git diff --check` passed.

Chromium Windows user-agent exercises Monaco's Windows platform branch on a Linux
host; it does not certify native Windows execution. Temporary services, browser
containers and helper resources exited. Next action: run the fixed updated commit
on the ordinary Windows device with `-ArchiveOnly`, returning its unique report.
E08 remains unmet and release work remains deferred.
