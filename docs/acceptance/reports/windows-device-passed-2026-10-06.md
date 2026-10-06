# Ordinary Windows real-device validation passed

Received `Blora-Windows-Device-e546c8f0-20261006-035356126-5b22bb25.zip`.
SHA-256: `a8a530127bc629530b61aa752359f2ed2c033d73d0cd4a0c540eb8fc1e1dd522`.
All twelve archive entries were inspected. The manifest covers exactly the
eleven other entries and all checksums match. Requested and executed source is
`e546c8f0fc2a766c678d8bef58417b60e1a36392`. Runner SHA-256 matches the fixed
source script: `a37cf45e4df0a7648a856d0ce15d9aae3541bc45a367e4fde2027d77c0ad08ab`.

Command: the fixed-source `scripts/windows-device-validation.ps1` with
`-Ref e546c8f0fc2a766c678d8bef58417b60e1a36392 -ArchiveOnly`.
Platform: **windows/amd64**. UTC run: 2026-10-06 03:53:56.1265558 to
03:54:55.2450290 (11:53:56–11:54:55 UTC+8). Eight stages passed with actual
exit code zero; overall `CHECKS_PASSED_WITH_SCOPE_LIMITS`, no reported failure.
The HTTPS fixed source download passed in 3.843 seconds; dependencies, production
frontend, Chromium installation and Master/Daemon/checker builds all passed.

Real device checks ran UTC 03:54:47.6234787–03:54:55.0346599, 7.411 seconds:

| Required check | Result |
| --- | --- |
| initialize-private-master | PASS |
| real-tls-static-login | PASS |
| enroll-real-daemon | PASS |
| create-stopped-owned-instance | PASS |
| real-user-permission-denial | PASS |
| real-file-save-and-read | PASS |
| real-backup-restore | PASS |
| real-monitor-sampling | PASS |
| real-extension-install-upgrade-remove | PASS |
| real-instance-start-stop | PASS |
| real-terminal-wss-input-resume-close | PASS |
| real-chromium-editor-refresh-save | PASS |
| stopped-state-process-restart | PASS |
| cleanup-owned-resources | PASS |

All fourteen names occur exactly once, without failed or omitted checks. Stage
logs agree with the structured outcomes. Browser login/resource/file creation,
unsaved draft refresh, exact server save readback and final refresh completed;
the LF fix has now passed on the actual Windows device. Restart followed confirmed
stopped resources, then confirmed cleanup; this is not a live-process crash claim.

R4 ordinary Windows real-device validation is closed. No repeat of this runner,
the already completed native group or visible-WebKit group is required without
new relevant changes/failures. Previous failed reports remain historical failures,
not rewritten as passed. The report contains no private state, process logs,
browser traces or screenshots.

This is the defined single-node real-backend/Chromium scope, not every Windows
platform/environment combination. Privileged host mutation, containers, physical
disk/power failures and native notification/IME combinations remain outside this
run. The user's optional-environment decision remains in force. E08 strict delay
is still unmet; release packaging/upgrade/rollback/publication remains deferred.
Current non-release remaining work is **R1 E08 only**.
