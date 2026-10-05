# Ordinary-device real-backend validation. No release, privileged host changes,
# race compiler, API doubles, or repetition of the completed browser followup.
[CmdletBinding()]
param(
    [Parameter(Mandatory=$true)][ValidatePattern('^[a-fA-F0-9]{40}$')][string]$Ref,
    [string]$WorkRoot = (Join-Path $env:TEMP 'BloraDeviceValidation')
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
$runnerHash = (Get-FileHash -LiteralPath $PSCommandPath -Algorithm SHA256).Hash.ToLowerInvariant()

function Quote-Literal([string]$Value) { return "'" + $Value.Replace("'", "''") + "'" }
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
function Invoke-DeviceStage([string]$Name, [string]$Command, [string[]]$Arguments, [string]$Directory, [int]$TimeoutSeconds) {
    $at = [DateTime]::UtcNow
    $number = $steps.Count + 1
    $logName = '{0:D2}-{1}.log' -f $number,$Name
    $log = New-Object IO.StreamWriter((Join-Path $reportDir $logName),$false,$utf8)
    $status = 'FAILED'; $code = -1
    $processStarted = $false
    Write-Host "[$number] RUN $Name"
    $process = New-Object Diagnostics.Process
    try {
        $parts = @($Arguments | ForEach-Object { Quote-Literal $_ })
        $body = '$ProgressPreference=''SilentlyContinue''; $ErrorActionPreference=''Stop''; Set-Location -LiteralPath ' + (Quote-Literal $Directory) + '; & ' + (Quote-Literal $Command) + ' ' + ($parts -join ' ') + '; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }'
        $encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($body))
        $process.StartInfo.FileName = (Get-Process -Id $PID).Path
        $process.StartInfo.Arguments = '-NoProfile -NonInteractive -EncodedCommand ' + $encoded
        $process.StartInfo.UseShellExecute = $false
        $process.StartInfo.RedirectStandardOutput = $true
        $process.StartInfo.RedirectStandardError = $true
        $process.StartInfo.CreateNoWindow = $true
        if (!$process.Start()) { throw 'Owned stage process did not start' }
        $processStarted = $true
        $script:active = $process
        $stdout = $process.StandardOutput.ReadLineAsync()
        $stderr = $process.StandardError.ReadLineAsync()
        $heartbeat = $at
        while (!$process.HasExited -or $null -ne $stdout -or $null -ne $stderr) {
            foreach ($stream in @('stdout','stderr')) {
                $pending = Get-Variable -Name $stream -ValueOnly
                if ($null -ne $pending -and $pending.IsCompleted) {
                    $line = $pending.GetAwaiter().GetResult()
                    if ($null -eq $line) { Set-Variable -Name $stream -Value $null }
                    else {
                        Write-Host $line; $log.WriteLine($line); $log.Flush()
                        $reader = $(if ($stream -eq 'stdout') { $process.StandardOutput } else { $process.StandardError })
                        Set-Variable -Name $stream -Value ($reader.ReadLineAsync())
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
        $process.WaitForExit()
        $code = $process.ExitCode
        if ($code -ne 0) { throw "Stage exited with code $code" }
        $status = 'PASS'
    } finally {
        if ($processStarted -and !$process.HasExited) {
            & taskkill.exe /PID $process.Id /T /F 2>&1 | Out-Null
        }
        $script:active = $null
        $log.Dispose(); $process.Dispose()
        $seconds = ([DateTime]::UtcNow-$at).TotalSeconds
        $steps.Add([pscustomobject]@{name=$Name;status=$status;exitCode=$code;seconds=$seconds;log=$logName})
        Write-Host ('[{0}] {1} {2} ({3:N1}s)' -f $number,$status,$Name,$seconds)
    }
}

try {
    if ($env:OS -ne 'Windows_NT') { throw 'Run this script in native Windows PowerShell, not WSL' }
    foreach ($tool in @('git.exe','go.exe','node.exe','npm.cmd')) {
        if (!(Get-Command $tool -ErrorAction SilentlyContinue)) { throw "Missing $tool; install Git, Go 1.25+ and Node.js 24+ first" }
    }
    Save-Text (Join-Path $reportDir 'run.json') (([ordered]@{runId=$runId;requestedRef=$Ref.ToLowerInvariant();startedAt=$started.ToString('o');runMode='real-device-non-release'}) | ConvertTo-Json)
    $repo = Join-Path $run 'source'
    Invoke-DeviceStage 'clone' 'git.exe' @('clone','--no-checkout','https://github.com/BloretCrew/BloraPanel.git',$repo) $run 300
    Invoke-DeviceStage 'checkout-fixed-source' 'git.exe' @('checkout','--detach',$Ref) $repo 60
    $commit = (& git.exe -C $repo rev-parse HEAD | Out-String).Trim()
    if ($LASTEXITCODE -ne 0 -or $commit -ne $Ref.ToLowerInvariant()) { throw 'Fetched source does not match the fixed requested commit' }
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
    $summary = [ordered]@{runId=$runId;requestedRef=$Ref.ToLowerInvariant();commit=$commit;runnerSHA256=$runnerHash;startedAt=$started.ToString('o');finishedAt=[DateTime]::UtcNow.ToString('o');runMode='real-device-non-release';outcome=$outcome;failure=$failure;steps=@($steps.ToArray());limitations=@('No release packaging, upgrades or compatible release rollback.','No E08 performance certification, privileged host changes, containers, disk exhaustion or physical power-loss testing.','Previous native/browser reports are preserved independently, not imported as current checks.')}
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
