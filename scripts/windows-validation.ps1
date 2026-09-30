#requires -Version 5.1
<#!
Run in a normal Windows PowerShell terminal. No remote access is required.
Only the new work directory is used; existing checkouts are never modified.
Optional tool installation uses winget and may show normal installer prompts.
#>
[CmdletBinding()]
param(
    [string]$WorkRoot = (Join-Path $env:USERPROFILE 'BloraValidation'),
    [string]$Ref = 'main',
    [switch]$InstallTools,
    [switch]$CollectLatest,
    [string]$CollectRun = ''
)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
if ($env:OS -ne 'Windows_NT') { throw 'Run this script on Windows, not WSL or Linux.' }
if ($Ref.StartsWith('-')) { throw 'Ref must be a branch, tag, or commit, not a Git option.' }
$utf8 = New-Object System.Text.UTF8Encoding($false)
function New-ReportZip([string]$Folder,[string]$Destination) {
    # Avoid Microsoft.PowerShell.Archive / Write-Progress host rendering entirely.
    Add-Type -AssemblyName System.IO.Compression
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $archive=[IO.Compression.ZipFile]::Open($Destination,[IO.Compression.ZipArchiveMode]::Create)
    try {
        foreach ($file in Get-ChildItem -LiteralPath $Folder -File) {
            if ($file.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'Report contains a linked file; refusing to follow it.' }
            [void][IO.Compression.ZipFileExtensions]::CreateEntryFromFile($archive,$file.FullName,$file.Name,[IO.Compression.CompressionLevel]::Optimal)
        }
    } finally { $archive.Dispose() }
}
if ($CollectLatest -or $CollectRun) {
    if ($CollectLatest -and $CollectRun) { throw 'Choose either -CollectLatest or -CollectRun.' }
    if ($CollectLatest) {
        $candidate=Get-ChildItem -LiteralPath $WorkRoot -Directory | Where-Object { Test-Path -LiteralPath (Join-Path $_.FullName 'report/steps-so-far.json') } | Sort-Object Name -Descending | Select-Object -First 1
        if (!$candidate) { throw 'No previous run found. Use -CollectRun with the previous work directory.' }
        $CollectRun=$candidate.FullName
    }
    $oldReport=Join-Path $CollectRun 'report'
    if (!(Test-Path -LiteralPath (Join-Path $oldReport 'steps-so-far.json'))) { throw 'No stage checkpoint in this run directory.' }
    $destination=Join-Path $CollectRun ('Blora-Windows-Recovered-'+[guid]::NewGuid().ToString('N').Substring(0,8)+'.zip')
    [IO.File]::WriteAllText((Join-Path $oldReport 'RECOVERY.txt'),"Recovered existing logs only. No tests were rerun and no missing test is counted as passed. Original report/exit records may describe an interrupted run. Collected UTC: $([DateTime]::UtcNow.ToString('o'))",$utf8)
    New-ReportZip $oldReport $destination
    Write-Host "Recovered previous evidence; tests NOT resumed.`nSend this ZIP: $destination" -ForegroundColor Cyan
    exit 0
}
$run = Join-Path $WorkRoot ((Get-Date -Format 'yyyyMMdd-HHmmss') + '-' + [guid]::NewGuid().ToString('N').Substring(0,8))
$report = Join-Path $run 'report'
$repo = Join-Path $run 'source'
[void](New-Item -ItemType Directory -Path $report -Force)
$steps = New-Object System.Collections.Generic.List[object]
$events = New-Object System.Collections.Generic.List[object]
$started = [DateTime]::UtcNow
$commit = 'not fetched'
$active = $null
$stage = 0
$fatal = ''
$savedEnvironment = @{}
# Do not inherit E2E switches/endpoints or helper modes from another session.
foreach ($item in Get-ChildItem Env:) {
    if ($item.Name -match '^(BLORA_|GOOS$|GOARCH$|GOFLAGS$|CGO_ENABLED$|GORACE$|CC$|CI$|PLAYWRIGHT_)') {
        $savedEnvironment[$item.Name] = $item.Value
        [Environment]::SetEnvironmentVariable($item.Name, $null, 'Process')
    }
}
function Clean([string]$text) {
    $text = $text.Replace($run, '<RUN>')
    if ($env:USERPROFILE) { $text = $text.Replace($env:USERPROFILE, '<USER>') }
    return $text
}
function Save-Text([string]$path, [string]$text) { [IO.File]::WriteAllText($path, $text, $utf8) }
function Missing([string]$name, [string]$reason) {
    $steps.Add([pscustomobject]@{name=$name;status='BLOCKED';exitCode=$null;seconds=0;detail=$reason;log=$null})
    Write-Host "[BLOCKED] $name : $reason" -ForegroundColor Yellow
}
function Invoke-Stage {
    param([string]$Name,[string]$Exe,[string[]]$Arguments=@(),[string]$Directory=$repo,[int]$TimeoutSeconds=1800,[switch]$GoEvents)
    $script:stage++
    $id = '{0:d2}-{1}' -f $script:stage, $Name
    Write-Host "`n[$script:stage] $Name" -ForegroundColor Cyan
    $logName = $id + '.log'
    $writer = New-Object IO.StreamWriter((Join-Path $report $logName), $false, $utf8)
    $clock = [Diagnostics.Stopwatch]::StartNew()
    $exitCode = -1; $status = 'FAIL'; $detail = ''; $process = $null; $nextHeartbeat=15
    try {
        # Encode structured arguments; no user path/ref is interpolated as shell code.
        $spec = @{exe=$Exe;arguments=@($Arguments);directory=$Directory} | ConvertTo-Json -Compress
        $payload = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($spec))
        $child = @"
`$ErrorActionPreference='Stop'
`$ProgressPreference='SilentlyContinue'
`$OutputEncoding=[Console]::OutputEncoding=New-Object Text.UTF8Encoding(`$false)
`$s=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('$payload')) | ConvertFrom-Json
try { Set-Location -LiteralPath `$s.directory; `$a=@(`$s.arguments); & `$s.exe @a; if (`$null -eq `$LASTEXITCODE) {exit 0}; exit `$LASTEXITCODE } catch { [Console]::Error.WriteLine(`$_.Exception.Message); exit 1 }
"@
        $psi = New-Object Diagnostics.ProcessStartInfo
        $psi.FileName = Join-Path $PSHOME 'powershell.exe'
        if (!(Test-Path $psi.FileName)) { $psi.FileName = Join-Path $PSHOME 'pwsh.exe' }
        if (!(Test-Path $psi.FileName)) { $psi.FileName = Join-Path $PSHOME 'pwsh' } # Function-level harness on Linux; entry still requires Windows.
        $psi.Arguments = '-NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -EncodedCommand ' + [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($child))
        $psi.UseShellExecute=$false; $psi.CreateNoWindow=$true
        $psi.RedirectStandardOutput=$true; $psi.RedirectStandardError=$true
        $psi.StandardOutputEncoding=$utf8; $psi.StandardErrorEncoding=$utf8
        $process = New-Object Diagnostics.Process
        $process.StartInfo=$psi
        [void]$process.Start(); $script:active=$process
        $reads = @($process.StandardOutput.ReadLineAsync(), $process.StandardError.ReadLineAsync())
        $streams = @($process.StandardOutput, $process.StandardError)
        while ($null -ne $reads[0] -or $null -ne $reads[1]) {
            if ($clock.Elapsed.TotalSeconds -gt $TimeoutSeconds) {
                $status='TIMEOUT'; $detail="Stage exceeded $TimeoutSeconds seconds; owned process tree termination requested. Inspect cleanup before rerunning."
                & taskkill.exe /PID $process.Id /T /F 2>&1 | ForEach-Object { $writer.WriteLine((Clean "$_")) }
                break
            }
            if ($clock.Elapsed.TotalSeconds -ge $nextHeartbeat) {
                Write-Host "[RUNNING] $Name - $([int]$clock.Elapsed.TotalSeconds)s"
                $nextHeartbeat=$clock.Elapsed.TotalSeconds+15
            }
            for ($i=0;$i -lt 2;$i++) {
                if ($null -ne $reads[$i] -and $reads[$i].IsCompleted) {
                    $line = $reads[$i].GetAwaiter().GetResult()
                    if ($null -eq $line) { $reads[$i]=$null; continue }
                    $safe = Clean $line
                    $writer.WriteLine($safe); $writer.Flush()
                    if ($GoEvents -and $i -eq 0) {
                        try {
                            $e = $line | ConvertFrom-Json
                            if ($e.Action -in @('pass','fail','skip') -and $e.Test) {
                                $events.Add([pscustomobject]@{stage=$Name;package=$e.Package;test=$e.Test;action=$e.Action;elapsed=$e.Elapsed})
                            }
                            if ($e.Action -in @('run','pass','fail','skip')) { Write-Host "  $($e.Action) $($e.Package) $($e.Test)" }
                            elseif ($e.Action -eq 'output' -and $e.Output) { Write-Host (Clean $e.Output.TrimEnd()) }
                        } catch { Write-Host $safe }
                    } else { Write-Host $safe }
                    $reads[$i]=$streams[$i].ReadLineAsync()
                }
            }
            Start-Sleep -Milliseconds 20
        }
        if ($status -ne 'TIMEOUT') {
            if (!$process.WaitForExit(10000)) { throw 'Output streams closed but child did not terminate.' }
            $exitCode=$process.ExitCode
            if ($exitCode -eq 0) { $status='PASS' }
        }
    } catch { $detail=Clean $_.Exception.Message; $writer.WriteLine($detail); Write-Host $detail -ForegroundColor Red }
    finally {
        if ($null -ne $process -and $null -ne $script:active) {
            if (!$process.HasExited) { & taskkill.exe /PID $process.Id /T /F 2>&1 | ForEach-Object { $writer.WriteLine((Clean "$_")) } }
            $process.Dispose()
        }
        $script:active=$null; $writer.Dispose(); $clock.Stop()
        $steps.Add([pscustomobject]@{name=$Name;status=$status;exitCode=$exitCode;seconds=[Math]::Round($clock.Elapsed.TotalSeconds,2);detail=$detail;log=$logName})
        Save-Text (Join-Path $report 'steps-so-far.json') (ConvertTo-Json -InputObject @($steps.ToArray()) -Depth 8)
        Write-Host "[$status] $Name ($([int]$clock.Elapsed.TotalSeconds)s)"
    }
    return ($status -eq 'PASS')
}
function Tool([string]$command,[string]$package) {
    if (Get-Command $command -ErrorAction SilentlyContinue) { return $true }
    if ($InstallTools -and (Get-Command winget.exe -ErrorAction SilentlyContinue)) {
        [void](Invoke-Stage "install-$package" 'winget.exe' @('install','--exact','--id',$package,'--source','winget','--accept-package-agreements','--accept-source-agreements','--disable-interactivity') $run 1800)
        $env:Path=$env:Path+';'+[Environment]::GetEnvironmentVariable('Path','Machine')+';'+[Environment]::GetEnvironmentVariable('Path','User')
    }
    if (Get-Command $command -ErrorAction SilentlyContinue) { return $true }
    Missing "tool-$command" "Install $package, reopen PowerShell, then rerun. -InstallTools can use winget."
    return $false
}
$uncovered = @(
    'Existing Windows SCM/Task Scheduler tests enumerate read-only; service start/stop, task changes and firewall rollback require separate isolated lifecycle acceptance.',
    'Browser suites use API doubles with real browser/xterm/storage. The Linux /bin/sh devfixture and Linux-only real-browser scenarios are NOT Windows full-stack evidence.',
    'Windows notification-center rendering/clicks, full mixed-load E08 <=50ms, hardware acceleration and long-running cross-host combinations are not proven by this runner.',
    'Real Docker/Compose requires an explicitly authorized disposable Engine; opt-in Engine tests remain skipped here.',
    'Power loss, device cache loss, Engine disk exhaustion and remote certificate/network operations are not performed.'
)
try {
    Write-Host "Work directory: $run`nA report ZIP will be written even when a stage fails." -ForegroundColor Cyan
    $gitReady=Tool 'git.exe' 'Git.Git'
    $goReady=Tool 'go.exe' 'GoLang.Go'
    $nodeReady=Tool 'npm.cmd' 'OpenJS.NodeJS.LTS'
    if (!$gitReady) { throw 'Git unavailable: cannot fetch source.' }
    if (!(Invoke-Stage 'clone' 'git.exe' @('clone','--no-checkout','https://github.com/BloretCrew/BloraPanel.git',$repo) $run)) { throw 'Clone failed; inspect the log.' }
    if (!(Invoke-Stage 'fetch-ref' 'git.exe' @('-C',$repo,'fetch','origin',$Ref) $run)) { throw 'Requested ref could not be fetched.' }
    if (!(Invoke-Stage 'checkout' 'git.exe' @('-C',$repo,'checkout','--detach','FETCH_HEAD') $run)) { throw 'Checkout failed.' }
    $commit=(& git.exe -C $repo rev-parse HEAD).Trim()
    [void](Invoke-Stage 'git-version' 'git.exe' @('--version'))
    if ($goReady) {
        [void](Invoke-Stage 'go-version' 'go.exe' @('version'))
        [void](Invoke-Stage 'go-toolchain' 'go.exe' @('env','GOOS','GOARCH','GOVERSION','CGO_ENABLED'))
        $env:CGO_ENABLED='0'
        [void](Invoke-Stage 'go-download' 'go.exe' @('mod','download') $repo 1800)
        [void](Invoke-Stage 'go-vet' 'go.exe' @('vet','./...') $repo 1800)
        foreach ($component in @('master','daemon','extension-sign')) {
            [void](Invoke-Stage "build-$component" 'go.exe' @('build','-trimpath','-o',(Join-Path $run "bin/blora-$component.exe"),"./cmd/$component"))
        }
        [void](Invoke-Stage 'windows-native' 'go.exe' @('test','-json','-count=1','-timeout=15m','./internal/runtime','./internal/terminal','./internal/runlog','./internal/monitor','./internal/systeminfo','-run','^TestWindows') $repo 3600 -GoEvents)
        [void](Invoke-Stage 'go-all' 'go.exe' @('test','-json','-count=1','-p=2','-timeout=20m','./...') $repo 7200 -GoEvents)
        if (Get-Command gcc.exe -ErrorAction SilentlyContinue) {
            $env:CGO_ENABLED='1'; $env:CC=(Get-Command gcc.exe).Source
            [void](Invoke-Stage 'gcc-version' 'gcc.exe' @('--version'))
            [void](Invoke-Stage 'go-race' 'go.exe' @('test','-race','-json','-count=1','-p=2','-timeout=20m','./...') $repo 7200 -GoEvents)
            $env:CGO_ENABLED='0'
        } else { Missing 'go-race' 'No gcc.exe on PATH. Windows race needs a compatible mingw-w64 C compiler; normal and native tests still run.' }
    }
    if ($nodeReady) {
        [void](Invoke-Stage 'node-version' 'node.exe' @('--version'))
        [void](Invoke-Stage 'npm-version' 'npm.cmd' @('--version'))
        foreach ($directory in @('sdk','sdk/examples/reference-app','web')) {
            $path=Join-Path $repo $directory; $label=$directory.Replace('/','-')
            if (Invoke-Stage "$label-install" 'npm.cmd' @('ci') $path 1800) {
                [void](Invoke-Stage "$label-build" 'npm.cmd' @('run','build') $path 1800)
                if ($directory -eq 'sdk/examples/reference-app') { [void](Invoke-Stage 'sdk-package' 'npm.cmd' @('run','package:fixtures') $path 1800) }
                if ($directory -eq 'web') {
                    [void](Invoke-Stage 'web-unit' 'npm.cmd' @('test') $path 1800)
                    $env:CI='1'; $env:BLORA_E2E_FRESH_SERVER='1'
                    foreach ($browser in @('chromium','firefox','webkit')) {
                        if (Invoke-Stage "install-$browser" 'node.exe' @('node_modules/@playwright/test/cli.js','install',$browser) $path 1800) {
                            $env:BLORA_BROWSER=$browser
                            $env:PLAYWRIGHT_JSON_OUTPUT_NAME=Join-Path $report "$browser.json"
                            [void](Invoke-Stage "browser-$browser" 'node.exe' @('node_modules/@playwright/test/cli.js','test','--workers=1','--reporter=line,json','--output',(Join-Path $run "browser-artifacts/$browser")) $path 7200)
                        } else { Missing "browser-$browser" 'Matching browser download failed.' }
                    }
                }
            } else { Missing "$label-tests" 'Dependency installation failed.' }
        }
    }
} catch { $fatal=Clean $_.Exception.Message; Write-Host $fatal -ForegroundColor Red }
finally {
    # Must see actual terminal PASS events, not merely a successful empty filter.
    $required=@('TestWindowsJobKeepsDescendantsAfterParentExits','TestWindowsJobCrashRecoveryPreservesRun','TestWindowsDaemonExitReopenAndStop','TestWindowsKeeperBirthMismatchFailsClosed','TestWindowsKeeperStartupFailureCleansEmptyJob','TestWindowsConPTYRealCommandResizeAndJobClose','TestWindowsDaemonExitDoesNotCloseBusinessPipe','TestWindowsNativeMetricsAndProcessIdentity','TestWindowsNativeServiceEnumeration','TestWindowsTaskSchedulerEnumeration')
    $missingTests=@($required | Where-Object { $name=$_; !($events | Where-Object { $_.stage -eq 'windows-native' -and $_.test -eq $name -and $_.action -eq 'pass' }) })
    $skips=@($events | Where-Object action -eq 'skip')
    if (!($events | Where-Object { $_.stage -eq 'go-all' -and $_.action -eq 'pass' })) { Missing 'go-all-coverage' 'No passing Go test events were collected for the full suite.' }
    $browserSummary=@()
    foreach ($browser in @('chromium','firefox','webkit')) {
        $file=Join-Path $report "$browser.json"
        if (Test-Path $file) {
            try {
                $text=Clean ([IO.File]::ReadAllText($file)); Save-Text $file $text
                $data=$text | ConvertFrom-Json
                $browserSummary+=@{browser=$browser;stats=$data.stats}
                if (!$data.stats -or $data.stats.expected -lt 1 -or $data.stats.skipped -gt 0 -or $data.stats.unexpected -gt 0 -or $data.stats.flaky -gt 0) { Missing "browser-$browser-coverage" 'No successful tests, skips, unexpected or flaky outcomes; see JSON.' }
            } catch { Missing "browser-$browser-report" 'Missing or invalid Playwright result data.' }
        } else { Missing "browser-$browser-report" 'No Playwright JSON report produced.' }
    }
    $outcome='AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS'
    if ($fatal -or @($steps | Where-Object { $_.status -in @('FAIL','TIMEOUT') }).Count) { $outcome='FAILED' }
    elseif ($missingTests.Count -or $skips.Count -or @($steps | Where-Object status -eq 'BLOCKED').Count) { $outcome='INCOMPLETE' }
    $result=[ordered]@{schema=1;outcome=$outcome;startedAt=$started.ToString('o');finishedAt=[DateTime]::UtcNow.ToString('o');commit=$commit;requestedRef=$Ref;windows=[Environment]::OSVersion.VersionString;architecture=$env:PROCESSOR_ARCHITECTURE;logicalProcessors=[Environment]::ProcessorCount;powershell=$PSVersionTable.PSVersion.ToString();fatal=$fatal;steps=@($steps.ToArray());requiredNativeTestsMissing=$missingTests;goTestEvents=@($events.ToArray());browserSummary=$browserSummary;notCovered=$uncovered}
    Save-Text (Join-Path $report 'report.json') ($result | ConvertTo-Json -Depth 30)
    $lines=@('# Blora Windows validation', '', "Result: **$outcome**", "Commit: $commit", "UTC: $($result.startedAt) to $($result.finishedAt)", '', 'This is an automated evidence bundle, not full Windows/platform acceptance.', '', '| Stage | Result | Seconds | Log |','|---|---|---:|---|')
    foreach ($s in $steps) { $lines+="| $($s.name) | $($s.status) | $($s.seconds) | $($s.log) |" }
    $lines+=@('', '## Required native tests without PASS', ($missingTests -join "`n"), '', '## Skipped Go tests')
    foreach ($s in $skips) { $lines+="- $($s.stage): $($s.package) / $($s.test)" }
    $lines+=@('', '## Not covered'); foreach ($gap in $uncovered) { $lines+="- $gap" }
    if ($fatal) { $lines+=@('', '## Fatal error', $fatal) }
    $lines+=@('', 'Only this report folder is zipped. Source, browser traces, credentials, environment dumps and test state are excluded.', 'Before sharing, review the logs for local paths or identifiers emitted by failed tests. Timeout/abort cleanup requires inspection; do not terminate unrelated services.')
    Save-Text (Join-Path $report 'README.md') ($lines -join "`r`n")
    $hashes=Get-ChildItem $report -File | Where-Object Name -ne 'SHA256SUMS.txt' | ForEach-Object { "$((Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLower())  $($_.Name)" }
    Save-Text (Join-Path $report 'SHA256SUMS.txt') ($hashes -join "`r`n")
    $zip=Join-Path $run 'Blora-Windows-Report.zip'
    $zipReady=$false
    try { New-ReportZip $report $zip; $zipReady=$true }
    catch { Write-Host "ZIP creation failed: $($_.Exception.Message). Report files remain at $report" -ForegroundColor Red; $outcome='FAILED' }
    foreach ($item in Get-ChildItem Env:) { if ($item.Name -match '^(BLORA_|GOOS$|GOARCH$|GOFLAGS$|CGO_ENABLED$|GORACE$|CC$|CI$|PLAYWRIGHT_)') { [Environment]::SetEnvironmentVariable($item.Name,$null,'Process') } }
    foreach ($key in $savedEnvironment.Keys) { [Environment]::SetEnvironmentVariable($key,$savedEnvironment[$key],'Process') }
    Write-Host "`nResult: $outcome`nSummary: $(Join-Path $report 'README.md')" -ForegroundColor Cyan
    if ($zipReady) { Write-Host "Send this ZIP: $zip" -ForegroundColor Cyan }
}
if ($outcome -eq 'FAILED') { exit 1 }
if ($outcome -eq 'INCOMPLETE') { exit 2 }
exit 0
