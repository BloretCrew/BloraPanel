# Ordinary-device real-backend validation. No release, privileged host changes,
# race compiler, API doubles, or repetition of the completed browser followup.
[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][ValidatePattern('^[a-fA-F0-9]{40}$')][string]$Ref,
    [string]$WorkRoot = (Join-Path $env:TEMP 'BloraDeviceValidation'),
    [switch]$ArchiveOnly
)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$utf8 = New-Object Text.UTF8Encoding($false)
$started = [DateTime]::UtcNow
$runId = $started.ToString('yyyyMMdd-HHmmssfff') + '-' + [guid]::NewGuid().ToString('N').Substring(0,8)
$run = Join-Path $WorkRoot $runId
$reportDir = Join-Path $run 'report'
[void](New-Item -ItemType Directory -Path $reportDir -Force)
$steps = New-Object 'System.Collections.Generic.List[object]'
$commit = ''; $failure = ''; $active = $null; $exitCode = 1
$sourceMethod = ''; $sourceAcquisitionRecovered = $false; $sourceArchiveHash = ''
$runnerHash = (Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant()

function Save-Text([string]$Path, [string]$Text) { [IO.File]::WriteAllText($Path, $Text, $utf8) }
function Assert-DeviceReport($Device) {
    if ($Device.platform -notmatch '^windows/' -or $Device.outcome -ne 'CHECKS_PASSED_WITH_SCOPE_LIMITS') { throw 'Actual Windows device checks did not pass' }
    $required = @('initialize-private-master','real-tls-static-login','enroll-real-daemon','create-stopped-owned-instance','real-user-permission-denial','real-file-save-and-read','real-backup-restore','real-monitor-sampling','real-extension-install-upgrade-remove','real-instance-start-stop','real-terminal-wss-input-resume-close','real-chromium-editor-refresh-save','stopped-state-process-restart','cleanup-owned-resources')
    foreach ($name in $required) {
        $matches = @($Device.checks | Where-Object { $_.name -eq $name })
        if ($matches.Count -ne 1 -or $matches[0].status -ne 'PASS') { throw "Missing successful actual device check: $name" }
    }
    if (@($Device.checks | Where-Object { $_.status -ne 'PASS' }).Count) { throw 'Device report includes a failed check' }
}
function Expand-FixedSourceArchive([string]$ArchivePath, [string]$RepositoryPath, [string]$ExpectedRef) {
    if ((Get-Item -LiteralPath $ArchivePath).Length -gt 128MB) { throw 'Source archive exceeds download size bound' }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $archive = [IO.Compression.ZipFile]::OpenRead($ArchivePath)
    try {
        $prefix = 'BloraPanel-' + $ExpectedRef.ToLowerInvariant() + '/'
        $root = [IO.Path]::GetFullPath($RepositoryPath) + [IO.Path]::DirectorySeparatorChar
        $seen = @{}; [long]$total = 0
        if ($archive.Entries.Count -gt 20000) { throw 'Source archive has too many entries' }
        # Validate the complete archive before creating or replacing source files.
        foreach ($entry in $archive.Entries) {
            $name = $entry.FullName
            if (!$name.StartsWith($prefix,[StringComparison]::OrdinalIgnoreCase) -or $name.Contains('\')) { throw 'Source archive does not match the fixed commit root' }
            $relative = $name.Substring($prefix.Length)
            if (!$relative) { continue }
            if ((($entry.ExternalAttributes -shr 16) -band 0xF000) -eq 0xA000) { throw 'Source archive symlink is unsupported' }
            foreach ($part in $relative.TrimEnd('/') -split '/') {
                if (!$part -or $part -eq '.' -or $part -eq '..' -or $part -match '[:\x00-\x1f]' -or $part -match '[. ]$' -or $part -match '^(?i:CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(?:\.|$)') { throw 'Unsafe source archive path' }
            }
            $target = [IO.Path]::GetFullPath((Join-Path $RepositoryPath $relative))
            if (!$target.StartsWith($root,[StringComparison]::OrdinalIgnoreCase) -or $seen.ContainsKey($target)) { throw 'Source archive path escapes or duplicates a target' }
            $seen[$target] = $true
            $total += $entry.Length
            if ($entry.Length -gt 128MB -or $total -gt 512MB) { throw 'Source archive exceeds expansion bound' }
        }
        if (Test-Path -LiteralPath $RepositoryPath) { throw 'Archive target must be a new task-owned directory' }
        [void](New-Item -ItemType Directory -Path $RepositoryPath)
        foreach ($entry in $archive.Entries) {
            $relative = $entry.FullName.Substring($prefix.Length)
            if (!$relative) { continue }
            $target = Join-Path $RepositoryPath $relative
            if ($entry.FullName.EndsWith('/')) { [void](New-Item -ItemType Directory -Path $target -Force); continue }
            [void](New-Item -ItemType Directory -Path ([IO.Path]::GetDirectoryName($target)) -Force)
            [IO.Compression.ZipFileExtensions]::ExtractToFile($entry,$target,$false)
        }
    } finally { $archive.Dispose() }
}
function Invoke-DeviceStage([string]$Name, [string]$Command, [string[]]$Arguments, [string]$Directory, [int]$TimeoutSeconds) {
    $at = [DateTime]::UtcNow
    $number = $steps.Count + 1
    $logName = '{0:D2}-{1}.log' -f $number,$Name
    $log = New-Object IO.StreamWriter((Join-Path $reportDir $logName),$false,$utf8)
    $status = 'FAILED'; $code = $null; $detail = ''
    $processStarted = $false
    Write-Host "[$number] RUN $Name"
    $process = New-Object Diagnostics.Process
    try {
        $spec = @{command=$Command;arguments=@($Arguments);directory=$Directory} | ConvertTo-Json -Compress
        $specPath = Join-Path $run ('stage-'+$number+'.json')
        $launcherPath = Join-Path $run ('stage-'+$number+'.ps1')
        Save-Text $specPath $spec
        $body = @'
param([string]$SpecPath)
$ProgressPreference='SilentlyContinue'
$env:GIT_TERMINAL_PROMPT='0'
$OutputEncoding=[Console]::OutputEncoding=New-Object Text.UTF8Encoding($false)
[Console]::Out.WriteLine('BLORA_STAGE_CHILD_STARTED')
try {
    $s=Get-Content -LiteralPath $SpecPath -Raw -Encoding UTF8 -ErrorAction Stop | ConvertFrom-Json
    Set-Location -LiteralPath $s.directory -ErrorAction Stop
    $tool=Get-Command $s.command -ErrorAction Stop
    $a=@($s.arguments)
    # Git/npm routinely print ordinary progress to stderr. A diagnostic line
    # must not abort Windows PowerShell 5.1 before the real exit code is read.
    $ErrorActionPreference='Continue'
    $PSNativeCommandUseErrorActionPreference=$false
    & $tool @a
    if ($null -eq $LASTEXITCODE) { throw 'Native command did not return an exit code' }
    exit $LASTEXITCODE
} catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
'@
        Save-Text $launcherPath $body
        $process.StartInfo.FileName = (Get-Process -Id $PID).Path
        # Both arguments are owned file paths, ending in .ps1/.json; Windows
        # filenames cannot contain a quote and neither has a trailing slash.
        if ($launcherPath.Contains('"') -or $specPath.Contains('"')) { throw 'Unsupported quote in stage path' }
        $process.StartInfo.Arguments = '-NoProfile -NonInteractive -ExecutionPolicy Bypass -File "' + $launcherPath + '" "' + $specPath + '"'
        $process.StartInfo.UseShellExecute = $false
        $process.StartInfo.RedirectStandardOutput = $true
        $process.StartInfo.RedirectStandardError = $true
        $process.StartInfo.StandardOutputEncoding = $utf8
        $process.StartInfo.StandardErrorEncoding = $utf8
        $process.StartInfo.CreateNoWindow = $true
        if (!$process.Start()) { throw 'Owned stage process did not start' }
        $processStarted = $true
        $log.WriteLine('LAUNCH: file-based PowerShell; child PID '+$process.Id); $log.Flush()
        $script:active = $process
        $reads = @($process.StandardOutput.ReadLineAsync(), $process.StandardError.ReadLineAsync())
        $streams = @($process.StandardOutput, $process.StandardError)
        $heartbeat = $at
        while ($null -ne $reads[0] -or $null -ne $reads[1]) {
            for ($i=0; $i -lt 2; $i++) {
                if ($null -ne $reads[$i] -and $reads[$i].IsCompleted) {
                    $line = $reads[$i].GetAwaiter().GetResult()
                    if ($null -eq $line) { $reads[$i] = $null }
                    else {
                        Write-Host $line; $log.WriteLine($line); $log.Flush()
                        $reads[$i] = $streams[$i].ReadLineAsync()
                    }
                }
            }
            $now = [DateTime]::UtcNow
            if (($now-$at).TotalSeconds -gt $TimeoutSeconds) {
                $status = 'TIMEOUT'
                # Only this run's child shell tree. Business resource cleanup is
                # owned by the bounded Go helper; a timeout can never pass.
                & taskkill.exe /PID $process.Id /T /F 2>&1 | Out-Null
                throw 'Stage deadline exceeded; cleanup could not be certified'
            }
            if (($now-$heartbeat).TotalSeconds -ge 15) {
                Write-Host ('[{0}] {1}: {2}s elapsed' -f $number,$Name,[int]($now-$at).TotalSeconds)
                $heartbeat = $now
            }
            Start-Sleep -Milliseconds 50
        }
        if (!$process.WaitForExit(10000)) { throw 'Child closed output streams but did not terminate' }
        $code = $process.ExitCode
        if ($code -ne 0) { throw "Stage exited with code $code" }
        $status = 'PASS'
    } catch {
        $detail = $_.Exception.Message
        $log.WriteLine($detail); $log.Flush()
        Write-Host $detail -ForegroundColor Red
        throw
    } finally {
        if ($processStarted -and !$process.HasExited) {
            & taskkill.exe /PID $process.Id /T /F 2>&1 | Out-Null
        }
        $script:active = $null
        $log.Dispose(); $process.Dispose()
        $seconds = ([DateTime]::UtcNow-$at).TotalSeconds
        $steps.Add([pscustomobject]@{name=$Name;status=$status;exitCode=$code;seconds=$seconds;detail=$detail;log=$logName})
        Write-Host ('[{0}] {1} {2} ({3:N1}s)' -f $number,$status,$Name,$seconds)
    }
}

try {
    if ($env:OS -ne 'Windows_NT') { throw 'Run this script in native Windows PowerShell, not WSL' }
    $requiredTools = @('go.exe','node.exe','npm.cmd')
    if (!$ArchiveOnly) { $requiredTools += 'git.exe' }
    foreach ($tool in $requiredTools) {
        if (!(Get-Command $tool -ErrorAction SilentlyContinue)) { throw "Missing $tool; install Git, Go 1.25+ and Node.js 24+ first" }
    }
    Save-Text (Join-Path $reportDir 'run.json') (([ordered]@{runId=$runId;requestedRef=$Ref.ToLowerInvariant();startedAt=$started.ToString('o');runMode='real-device-non-release'}) | ConvertTo-Json)
    $repo = Join-Path $run 'source'
    try {
        if ($ArchiveOnly) { throw 'Fixed source archive selected' }
        Invoke-DeviceStage 'clone' 'git.exe' @('-c','core.autocrlf=false','-c','core.askPass=','-c','credential.interactive=false','clone','--no-checkout','https://github.com/BloretCrew/BloraPanel.git',$repo) $run 180
        Invoke-DeviceStage 'checkout-fixed-source' 'git.exe' @('-c','core.autocrlf=false','checkout','--detach',$Ref) $repo 60
        $commit = (& git.exe -C $repo rev-parse HEAD | Out-String).Trim()
        if ($LASTEXITCODE -ne 0 -or $commit -ne $Ref.ToLowerInvariant()) { throw 'Fetched source does not match the fixed requested commit' }
        $sourceMethod = 'git-fixed-commit'
    } catch {
        if ($ArchiveOnly) { Write-Host 'Downloading GitHub HTTPS source archive for the fixed commit.' }
        else { Write-Host 'Git source acquisition failed; trying GitHub HTTPS archive for the same fixed commit.' -ForegroundColor Yellow }
        if (Test-Path -LiteralPath $repo) { Remove-Item -LiteralPath $repo -Recurse -Force }
        $archivePath = Join-Path $run 'fixed-source.zip'
        $uri = 'https://codeload.github.com/BloretCrew/BloraPanel/zip/' + $Ref.ToLowerInvariant()
        $downloadPath = Join-Path $run 'download-source.ps1'
        Save-Text $downloadPath @'
param([string]$Uri,[string]$OutputPath)
$ProgressPreference='SilentlyContinue'
[Console]::Out.WriteLine('BLORA_SOURCE_DOWNLOAD_STARTED')
try {
    [Net.ServicePointManager]::SecurityProtocol=[Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -UseBasicParsing -Uri $Uri -OutFile $OutputPath -TimeoutSec 120 -ErrorAction Stop
    [Console]::Out.WriteLine('BLORA_SOURCE_DOWNLOAD_COMPLETED')
    exit 0
} catch { [Console]::Error.WriteLine($_.Exception.Message); exit 1 }
'@
        Invoke-DeviceStage 'fixed-source-archive-download' (Get-Process -Id $PID).Path @('-NoProfile','-NonInteractive','-ExecutionPolicy','Bypass','-File',$downloadPath,$uri,$archivePath) $run 150
        Expand-FixedSourceArchive $archivePath $repo $Ref
        $sourceArchiveHash = (Get-FileHash -LiteralPath $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
        $commit = $Ref.ToLowerInvariant()
        $sourceMethod = 'https-fixed-commit-archive'
        $sourceAcquisitionRecovered = !$ArchiveOnly
    }
    $sourceRunnerHash = (Get-FileHash -LiteralPath (Join-Path $repo 'scripts/windows-device-validation.ps1') -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($sourceRunnerHash -ne $runnerHash) { throw 'Downloaded runner does not match the fixed source; download the script from the same commit' }
    $web = Join-Path $repo 'web'
    Invoke-DeviceStage 'dependencies' 'npm.cmd' @('ci','--no-audit','--no-fund') $web 600
    Invoke-DeviceStage 'frontend-test-build' 'npm.cmd' @('run','build') $web 300
    Invoke-DeviceStage 'chromium-only' 'node.exe' @('node_modules/@playwright/test/cli.js','install','chromium') $web 600
    $bin = Join-Path $run 'bin'
    [void](New-Item -ItemType Directory -Path $bin)
    foreach ($target in @('master','daemon','device-validation')) {
        $name = 'blora-' + $target + '.exe'
        Invoke-DeviceStage ('build-'+$target) 'go.exe' @('build','-o',(Join-Path $bin $name),('./cmd/'+$target)) $repo 600
    }
    $helper = Join-Path $bin 'blora-device-validation.exe'
    $helperArguments = @('--master',(Join-Path $bin 'blora-master.exe'),'--daemon',(Join-Path $bin 'blora-daemon.exe'),'--static',(Join-Path $web 'dist'),'--report',(Join-Path $reportDir 'device-report.json'),'--state-parent',(Join-Path $run 'private'),'--browser-script',(Join-Path $web 'scripts/device-browser.mjs'),'--node',(Get-Command node.exe).Source)
    Invoke-DeviceStage 'real-backend-browser-and-recovery' $helper $helperArguments $repo 600
    $device = Get-Content -LiteralPath (Join-Path $reportDir 'device-report.json') -Raw | ConvertFrom-Json
    Assert-DeviceReport $device
    $exitCode = 0
} catch {
    $failure = $_.Exception.Message
    Write-Host "FAIL: $failure" -ForegroundColor Red
} finally {
    if ($null -ne $active) { & taskkill.exe /PID $active.Id /T /F 2>&1 | Out-Null }
    $outcome = $(if ($exitCode -eq 0) { 'CHECKS_PASSED_WITH_SCOPE_LIMITS' } else { 'FAILED' })
    $summary = [ordered]@{runId=$runId;requestedRef=$Ref.ToLowerInvariant();commit=$commit;sourceMethod=$sourceMethod;sourceAcquisitionRecovered=$sourceAcquisitionRecovered;sourceArchiveSHA256=$sourceArchiveHash;runnerSHA256=$runnerHash;startedAt=$started.ToString('o');finishedAt=[DateTime]::UtcNow.ToString('o');runMode='real-device-non-release';outcome=$outcome;failure=$failure;steps=@($steps.ToArray());limitations=@('No release packaging, upgrades or compatible release rollback.','No E08 performance certification, privileged host changes, containers, disk exhaustion or physical power-loss testing.','Previous native/browser reports are preserved independently, not imported as current checks.')}
    Save-Text (Join-Path $reportDir 'report.json') ($summary | ConvertTo-Json -Depth 8)
    $lines = @(Get-ChildItem -LiteralPath $reportDir -File | Where-Object { $_.Name -ne 'SHA256SUMS.txt' } | Sort-Object Name | ForEach-Object { ((Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256).Hash.ToLowerInvariant() + '  ' + $_.Name) })
    Save-Text (Join-Path $reportDir 'SHA256SUMS.txt') (($lines -join "`n") + "`n")
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $zip = Join-Path $run ('Blora-Windows-Device-'+$Ref.Substring(0,8).ToLowerInvariant()+'-'+$runId+'.zip')
    [IO.Compression.ZipFile]::CreateFromDirectory($reportDir,$zip)
    Write-Host "Report: $zip"
    Write-Host 'Return this ZIP. Private credentials, database/state, browser snapshots and raw process logs are excluded.'
}
exit $exitCode
