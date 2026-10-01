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
    [switch]$Remaining,
    [switch]$Retest,
    [switch]$Followup,
    [switch]$InstallRaceCompiler,
    [switch]$CollectLatest,
    [string]$CollectRun = ''
)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
if ($env:OS -ne 'Windows_NT') { throw 'Run this script on Windows, not WSL or Linux.' }
if ($Ref.StartsWith('-')) { throw 'Ref must be a branch, tag, or commit, not a Git option.' }
if ($Retest -and $Followup) { throw 'Choose either -Retest or -Followup.' }
if ($Retest -or $Followup) { $Remaining = $true }
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
$savedPath = $env:Path
# Do not inherit E2E switches/endpoints or helper modes from another session.
foreach ($item in Get-ChildItem Env:) {
    if ($item.Name -match '^(BLORA_|GOOS$|GOARCH$|GOFLAGS$|CGO_ENABLED$|GORACE$|CC$|CI$|PLAYWRIGHT_)') {
        $savedEnvironment[$item.Name] = $item.Value
        [Environment]::SetEnvironmentVariable($item.Name, $null, 'Process')
    }
}
function Clean([string]$text) {
    $text = $text.Replace($run, '<RUN>')
    $text = $text.Replace($run.Replace('\','\\'), '<RUN>')
    if ($env:USERPROFILE) {
        $text = $text.Replace($env:USERPROFILE, '<USER>')
        $text = $text.Replace($env:USERPROFILE.Replace('\','\\'), '<USER>')
    }
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
    # A terminal opened before winget installation may retain a stale PATH.
    # Refresh registered paths before treating an installed tool as missing.
    $env:Path=$env:Path+';'+[Environment]::GetEnvironmentVariable('Path','Machine')+';'+[Environment]::GetEnvironmentVariable('Path','User')
    if ($command -eq 'go.exe' -and !(Get-Command $command -ErrorAction SilentlyContinue)) {
        foreach ($base in @($env:ProgramFiles,$env:LOCALAPPDATA)) {
            if ($base) {
                $bin=Join-Path $base 'Go/bin'
                if (Test-Path -LiteralPath (Join-Path $bin 'go.exe') -PathType Leaf) { $env:Path+=';'+$bin; break }
            }
        }
    }
    if (Get-Command $command -ErrorAction SilentlyContinue) { return $true }
    if ($InstallTools -and (Get-Command winget.exe -ErrorAction SilentlyContinue)) {
        [void](Invoke-Stage "install-$package" 'winget.exe' @('install','--exact','--id',$package,'--source','winget','--accept-package-agreements','--accept-source-agreements','--disable-interactivity') $run 1800)
        $env:Path=$env:Path+';'+[Environment]::GetEnvironmentVariable('Path','Machine')+';'+[Environment]::GetEnvironmentVariable('Path','User')
    }
    if (Get-Command $command -ErrorAction SilentlyContinue) { return $true }
    Missing "tool-$command" "Install $package, reopen PowerShell, then rerun. -InstallTools can use winget."
    return $false
}
function Get-RetestPlan {
    # Every failing named case in the third returned report, plus the new
    # directory attributes, portable lifecycle and connection-loss regressions.
    $cases = [ordered]@{
        backup = @('TestBackupRoundTripMetadataAndExplicitOverwrite')
        bridge = @('TestSendPreservesClosedTransportIdentity','TestRemoteOperationErrorIsNotTransportLoss')
        filesystem = @('TestWindowsMetadataAttributesAndTimes','TestMetadataVersionCancellationAndLinkBoundary','TestUploadMetadataIsBoundToDurableIdentity','TestRelationFailsWhenCanonicalRootOrSourceObjectIsReplaced','TestArchiveRejectsTraversalLinksDuplicatesAndBombs')
        protocol = @('TestBulkRateWaitHonorsWriteDeadline')
        terminal = @('TestWindowsConPTYRealCommandResizeAndJobClose')
        master = @('TestInstanceSettingsAreVersionedAndRootsRemainSeparate','TestAutostartRunsOncePerDaemonStartupAndRechecksOwner','TestFailedStopBlocksReplacementButOtherNodeRemainsUsable','TestRestartTaskSurvivesMasterAndDatabaseReopen','TestLogSlowConsumerDeadlineAndCursorReattach','TestConsoleInputHasOneDeliveryAndGracefulStopUsesIndependentStdin','TestNativeLogArchiveDoesNotLetSlowBrowserBlockStop','TestMetricsUseLiveNodeAndOwnedProcessScope','TestInstanceMetricDisconnectAndRunReplacement','TestMetricsAggregateOwnedDescendants','TestNodeMaintenancePreservesRuntimeAndSettingsAcrossHeartbeat','TestSchedulesAcceptRealLifecycleConsoleAndFreshBackup','TestTransferRealDirectoryCopyMoveAndMetadata','TestTransferDirectoryMergeChecksFileBaselinesAndKeepsExtraFiles','TestTransferSameNodeSeparateInstancesAndActualOverlapRejection','TestTwoNodesAuthorizationAndDurableLifecycle','TestTransferSourceDaemonRestartRebuildsProof')
    }
    $goCases = @(); foreach ($package in $cases.Keys) { foreach ($test in $cases[$package]) { $goCases += [pscustomobject]@{package="blora.dev/panel/internal/$package";test=$test} } }
    return [pscustomobject]@{
        goCases = $goCases
        browserCaseCount = 32
        browserFiles = @('tests/browser/desktop.spec.ts','tests/browser/accounts.spec.ts','tests/browser/control-bar.spec.ts','tests/browser/editor-history-budget.spec.ts','tests/browser/editor-multicursor.spec.ts','tests/browser/editor-save.spec.ts','tests/browser/editor-split.spec.ts','tests/browser/extensions.spec.ts','tests/browser/terminal.spec.ts')
        browserTitles = @('Monaco text, undo/redo and original view identity survive immediate reload and moving windows','committed Unicode, simultaneous multi-cursor edits and find state survive immediate reload','password changes clear hidden fields, exclude recovery and require a new login after confirmation','system and workspace reduced motion both keep the launcher stationary on entry','undo history keeps the same byte budget before and after refresh','multi-cursor edits remain one reversible protected history group after refresh','remote save task keeps captured body and never overwrites newer local input after reload','editor split panes share text while keeping both cursors and the split visible after refresh','independently packaged reference extension renders and restores its note','xterm checkpoints resume the same session, ACK parsed bytes and never replay input (automatic renderer, worker checkpoints)','xterm checkpoints resume the same session, ACK parsed bytes and never replay input (fallback renderer, worker checkpoints)','xterm checkpoints resume the same session, ACK parsed bytes and never replay input (fallback renderer, main checkpoints)')
    }
}
function Get-FollowupPlan {
    # Fourth returned report: only two Go modules changed; two browser files
    # retain every original assertion and add isolated background transports
    # plus failure-only runtime diagnostics. No old passes are imported.
    return [pscustomobject]@{
        goCases = @(
            [pscustomobject]@{package='blora.dev/panel/internal/master';test='TestCancelAcceptedUploadBeforeNodePreparation'},
            [pscustomobject]@{package='blora.dev/panel/internal/runlog';test='TestStdinSurvivesDaemonAndOutputEOF'},
            [pscustomobject]@{package='blora.dev/panel/internal/runlog';test='TestFinishRechecksDurableRecordAfterHelperExit'}
        )
        browserCaseCount = 6
        browserFiles = @('tests/browser/accounts.spec.ts','tests/browser/terminal.spec.ts')
        browserTitles = @('rejected login clears submitted credentials immediately','account and permission confirmations restore while passwords never enter workspace recovery','password changes clear hidden fields, exclude recovery and require a new login after confirmation','xterm checkpoints resume the same session, ACK parsed bytes and never replay input (automatic renderer, worker checkpoints)','xterm checkpoints resume the same session, ACK parsed bytes and never replay input (fallback renderer, worker checkpoints)','xterm checkpoints resume the same session, ACK parsed bytes and never replay input (fallback renderer, main checkpoints)')
    }
}
function Get-BrowserDiagnostics($Suites) {
    foreach ($suite in @($Suites)) {
        foreach ($spec in @($suite.specs)) {
            foreach ($test in @($spec.tests)) {
                foreach ($result in @($test.results)) {
                    foreach ($attachment in @($result.attachments)) {
                        if ($attachment.name -eq 'blora-runtime-diagnostics' -and $attachment.contentType -eq 'application/json' -and $attachment.body) { $attachment.body }
                    }
                }
            }
        }
        if ($suite.suites) { Get-BrowserDiagnostics $suite.suites }
    }
}
function Copy-BrowserDiagnostics([string]$Browser,$Suites) {
    $index=0
    foreach ($body in @(Get-BrowserDiagnostics $Suites)) {
        if ($index -ge 24) { break }
        if ($body.Length -gt 87384) { continue }
        try {
            $bytes=[Convert]::FromBase64String($body)
            if ($bytes.Length -gt 65536) { continue }
            $text=Clean ([Text.Encoding]::UTF8.GetString($bytes))
            [void]($text | ConvertFrom-Json)
        } catch { continue }
        $index++
        Save-Text (Join-Path $report "browser-$Browser-diagnostic-$index.json") $text
    }
}
function Get-BrowserPassTitles($Suites) {
    foreach ($suite in @($Suites)) {
        foreach ($spec in @($suite.specs)) {
            # Do not accept expected failures, retries or skipped executions as
            # regression evidence. Each selected engine must pass the case.
            if ($spec.ok -and @($spec.tests).Count -gt 0 -and !@($spec.tests | Where-Object { $_.expectedStatus -ne 'passed' -or $_.status -ne 'expected' -or !@($_.results | Where-Object status -eq 'passed').Count -or @($_.results | Where-Object { $_.status -ne 'passed' }).Count }).Count) { $spec.title }
        }
        if ($suite.suites) { Get-BrowserPassTitles $suite.suites }
    }
}
$retestPlan = $(if ($Followup) { Get-FollowupPlan } else { Get-RetestPlan })
$uncovered = @(
    'Existing Windows SCM/Task Scheduler tests enumerate read-only; service start/stop, task changes and firewall rollback require separate isolated lifecycle acceptance.',
    'Browser suites use API doubles with real browser/xterm/storage. The Linux /bin/sh devfixture and Linux-only real-browser scenarios are NOT Windows full-stack evidence.',
    'Windows notification-center rendering/clicks, full mixed-load E08 <=50ms, hardware acceleration and long-running cross-host combinations are not proven by this runner.',
    'Real Docker/Compose requires an explicitly authorized disposable Engine; opt-in Engine tests remain skipped here.',
    'Power loss, device cache loss, Engine disk exhaustion and remote certificate/network operations are not performed.'
)
if ($Followup) {
    $uncovered += 'Followup mode covers the fourth returned report: three named Go regressions (three executions), eleven native checks, full race testing of Master/runlog only, SDK packages, and all six scenarios in the two affected browser files. WebKit runs each scenario three times; Chromium/Firefox once. Other Go packages/browser files, standalone vet/builds, production web/unit checks and native IME are intentionally omitted; prior passes are NOT imported.'
    $uncovered += 'Local repetition reproduced the WebKit background-request diagnostic and occasional checkpoint delays; the Windows account pointer-stability cause remains unproven. Complete transport doubles and foreground/scroll preparation are followup fixture corrections, not proof of a Windows product fix. Original page-error, recovery, password exclusion, ACK and no-replay assertions remain active. Failure-only diagnostics contain error/request paths and paint/control geometry, never field contents or storage dumps.'
} elseif ($Retest) {
    $uncovered += 'Retest mode reruns the returned failing Go cases, the native suite, and all cases in nine affected browser files (excluding the already-passed Chromium-only native IME scenario). Required case names are checked individually. Prior results are NOT imported as passes; full all-package race testing is attempted because it was previously blocked.'
    $uncovered += 'This targeted run omits standalone vet/builds, web production build/unit tests, unaffected browser files and unaffected non-race Go cases. Native IME injection in Firefox/WebKit remains a capability gap, not a passing test.'
} elseif ($Remaining) {
    $uncovered += 'Remaining mode intentionally omits Go vet, standalone Go builds, web production build and web unit tests; prior results are NOT imported as passes. SDK build/package, native checks, full Go regression, available race checks and all three browsers run again.'
}
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
    function Invoke-GoChecks {
        [void](Invoke-Stage 'go-version' 'go.exe' @('version'))
        [void](Invoke-Stage 'go-toolchain' 'go.exe' @('env','GOOS','GOARCH','GOVERSION','CGO_ENABLED'))
        $env:CGO_ENABLED='0'
        [void](Invoke-Stage 'go-download' 'go.exe' @('mod','download') $repo 1800)
        if (!$Remaining) {
            [void](Invoke-Stage 'go-vet' 'go.exe' @('vet','./...') $repo 1800)
            foreach ($component in @('master','daemon','extension-sign')) {
                [void](Invoke-Stage "build-$component" 'go.exe' @('build','-trimpath','-o',(Join-Path $run "bin/blora-$component.exe"),"./cmd/$component"))
            }
        }
        [void](Invoke-Stage 'windows-native' 'go.exe' @('test','-json','-count=1','-timeout=15m','./internal/runtime','./internal/terminal','./internal/runlog','./internal/monitor','./internal/systeminfo','./internal/filesystem','-run','^TestWindows') $repo 3600 -GoEvents)
        if ($Retest -or $Followup) {
            $pattern = '^(' + (($retestPlan.goCases | ForEach-Object { [regex]::Escape($_.test) }) -join '|') + ')$'
            $packages = @($retestPlan.goCases | ForEach-Object { $_.package.Replace('blora.dev/panel/','./') } | Select-Object -Unique)
            $count = $(if ($Followup) { '-count=3' } else { '-count=1' })
            [void](Invoke-Stage 'go-retest' 'go.exe' (@('test','-json',$count,'-p=2','-timeout=20m') + $packages + @('-run',$pattern)) $repo 3600 -GoEvents)
        } else {
            [void](Invoke-Stage 'go-all' 'go.exe' @('test','-json','-count=1','-p=2','-timeout=20m','./...') $repo 7200 -GoEvents)
        }
        $gcc = Get-Command gcc.exe -ErrorAction SilentlyContinue
        if ($InstallRaceCompiler) {
            $compilerRoot = Join-Path $run 'race-compiler'
            $shell = Join-Path $PSHOME 'powershell.exe'; if (!(Test-Path $shell)) { $shell = Join-Path $PSHOME 'pwsh.exe' }
            if (Invoke-Stage 'install-race-compiler' $shell @('-NoLogo','-NoProfile','-NonInteractive','-ExecutionPolicy','Bypass','-File',(Join-Path $repo 'scripts/windows-race-compiler.ps1'),'-Destination',$compilerRoot) $repo 2400) {
                $env:Path=(Join-Path $compilerRoot 'mingw64/bin')+';'+$env:Path
                $gcc = Get-Command gcc.exe -ErrorAction SilentlyContinue
            } else { $gcc = $null }
        }
        if ($gcc) {
            $env:CGO_ENABLED='1'; $env:CC=(Get-Command gcc.exe).Source
            [void](Invoke-Stage 'gcc-version' 'gcc.exe' @('--version'))
            if (Invoke-Stage 'gcc-race-runtime' 'gcc.exe' @('--print-file-name','libsynchronization.a')) {
                $racePackages = $(if ($Followup) { @('./internal/master','./internal/runlog') } else { @('./...') })
                [void](Invoke-Stage 'go-race' 'go.exe' (@('test','-race','-json','-count=1','-p=2','-timeout=20m') + $racePackages) $repo 7200 -GoEvents)
            } else { Missing 'go-race' 'Compiler runtime probe failed.' }
            $env:CGO_ENABLED='0'
        } else { Missing 'go-race' 'No usable gcc.exe. Use -InstallRaceCompiler for a pinned, SHA-256-verified portable MinGW-w64 compiler in this run directory; no system installation is required.' }
    }
    $goChecksRun=$false
    if ($nodeReady) {
        [void](Invoke-Stage 'node-version' 'node.exe' @('--version'))
        [void](Invoke-Stage 'npm-version' 'npm.cmd' @('--version'))
        foreach ($directory in @('sdk','sdk/examples/reference-app','web')) {
            if ($directory -eq 'web' -and $goReady) { Invoke-GoChecks; $goChecksRun=$true }
            $path=Join-Path $repo $directory; $label=$directory.Replace('/','-')
            if (Invoke-Stage "$label-install" 'npm.cmd' @('ci') $path 1800) {
                if (!$Remaining -or $directory -ne 'web') { [void](Invoke-Stage "$label-build" 'npm.cmd' @('run','build') $path 1800) }
                if ($directory -eq 'sdk/examples/reference-app') {
                    if ($goReady) {
                        [void](Invoke-Stage 'sdk-package-default' 'npm.cmd' @('run','package') $path 1800)
                        [void](Invoke-Stage 'sdk-package' 'npm.cmd' @('run','package:fixtures') $path 1800)
                    } else {
                        Missing 'sdk-package' 'Go unavailable: WASI packages cannot be generated. The reference-extension browser scenario will also lack its package.'
                    }
                }
                if ($directory -eq 'web') {
                    if (!$Remaining) { [void](Invoke-Stage 'web-unit' 'npm.cmd' @('test') $path 1800) }
                    $env:CI='1'; $env:BLORA_E2E_FRESH_SERVER='1'
                    foreach ($browser in @('chromium','firefox','webkit')) {
                        if (Invoke-Stage "install-$browser" 'node.exe' @('node_modules/@playwright/test/cli.js','install',$browser) $path 1800) {
                            $env:BLORA_BROWSER=$browser
                            $env:PLAYWRIGHT_JSON_OUTPUT_NAME=Join-Path $report "$browser.json"
                            $browserArgs = @('node_modules/@playwright/test/cli.js','test','--workers=1','--reporter=line,json','--output',(Join-Path $run "browser-artifacts/$browser"))
                            if ($Retest -or $Followup) { $browserArgs += @($retestPlan.browserFiles) }
                            if ($Retest) { $browserArgs += @('--grep-invert','native Chromium IME composition survives immediate reload') }
                            if ($Followup -and $browser -eq 'webkit') { $browserArgs += '--repeat-each=3' }
                            [void](Invoke-Stage "browser-$browser" 'node.exe' $browserArgs $path 7200)
                        } else { Missing "browser-$browser" 'Matching browser download failed.' }
                    }
                }
            } else { Missing "$label-tests" 'Dependency installation failed.' }
        }
    }
    # Master integration tests consume the actual reference package files.
    if ($goReady -and !$goChecksRun) { Invoke-GoChecks }
} catch { $fatal=Clean $_.Exception.Message; Write-Host $fatal -ForegroundColor Red }
finally {
    # Must see actual terminal PASS events, not merely a successful empty filter.
    $required=@('TestWindowsJobKeepsDescendantsAfterParentExits','TestWindowsJobCrashRecoveryPreservesRun','TestWindowsDaemonExitReopenAndStop','TestWindowsKeeperBirthMismatchFailsClosed','TestWindowsKeeperStartupFailureCleansEmptyJob','TestWindowsConPTYRealCommandResizeAndJobClose','TestWindowsDaemonExitDoesNotCloseBusinessPipe','TestWindowsNativeMetricsAndProcessIdentity','TestWindowsNativeServiceEnumeration','TestWindowsTaskSchedulerEnumeration','TestWindowsMetadataAttributesAndTimes')
    $missingTests=@($required | Where-Object { $name=$_; !($events | Where-Object { $_.stage -eq 'windows-native' -and $_.test -eq $name -and $_.action -eq 'pass' }) })
    $skips=@($events | Where-Object action -eq 'skip')
    $missingRetests = @()
    if ($Retest -or $Followup) {
        $requiredExecutions = $(if ($Followup) { 3 } else { 1 })
        $missingRetests = @($retestPlan.goCases | Where-Object { $case=$_; @($events | Where-Object { $_.stage -eq 'go-retest' -and $_.package -eq $case.package -and $_.test -eq $case.test -and $_.action -eq 'pass' }).Count -lt $requiredExecutions })
        if ($missingRetests.Count) { Missing 'go-retest-coverage' "Required cases without PASS: $($missingRetests.Count). See report.json; an empty or incomplete selection is not a pass." }
    } elseif (!($events | Where-Object { $_.stage -eq 'go-all' -and $_.action -eq 'pass' })) { Missing 'go-all-coverage' 'No passing Go test events were collected for the full suite.' }
    if (!($events | Where-Object { $_.stage -eq 'go-race' -and $_.action -eq 'pass' })) { Missing 'go-race-coverage' 'No passing race test events were collected.' }
    $browserSummary=@()
    foreach ($browser in @('chromium','firefox','webkit')) {
        $file=Join-Path $report "$browser.json"
        if (Test-Path $file) {
            try {
                $text=Clean ([IO.File]::ReadAllText($file)); Save-Text $file $text
                $data=$text | ConvertFrom-Json
                $missingTitles = @()
                if ($Retest -or $Followup) {
                    $passedTitles = @(Get-BrowserPassTitles $data.suites)
                    $requiredBrowserExecutions = $(if ($Followup -and $browser -eq 'webkit') { 3 } else { 1 })
                    $missingTitles = @($retestPlan.browserTitles | Where-Object { $title=$_; @($passedTitles | Where-Object { $_ -eq $title }).Count -lt $requiredBrowserExecutions })
                    if ($missingTitles.Count) { Missing "browser-$browser-retest-coverage" "Required scenarios without PASS: $($missingTitles.Count). See browserSummary in report.json." }
                }
                $expectedCount = $(if ($Retest -or $Followup) { $retestPlan.browserCaseCount * $(if ($Followup -and $browser -eq 'webkit') { 3 } else { 1 }) } else { $null })
                $browserSummary+=@{browser=$browser;stats=$data.stats;expectedCaseCount=$expectedCount;requiredRetestTitlesMissing=$missingTitles}
                if ($null -ne $expectedCount -and $data.stats.expected -ne $expectedCount) { Missing "browser-$browser-case-count" "Expected $expectedCount passing executions; omitted or incomplete scenarios are not a pass." }
                if (!$data.stats -or $data.stats.expected -lt 1 -or $data.stats.skipped -gt 0 -or $data.stats.unexpected -gt 0 -or $data.stats.flaky -gt 0) { Missing "browser-$browser-coverage" 'No successful tests, skips, unexpected or flaky outcomes; see JSON.' }
                Copy-BrowserDiagnostics $browser $data.suites
            } catch { Missing "browser-$browser-report" 'Missing or invalid Playwright result data.' }
        } else { Missing "browser-$browser-report" 'No Playwright JSON report produced.' }
    }
    $outcome='AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS'
    if ($fatal -or @($steps | Where-Object { $_.status -in @('FAIL','TIMEOUT') }).Count) { $outcome='FAILED' }
    elseif ($missingTests.Count -or $skips.Count -or @($steps | Where-Object status -eq 'BLOCKED').Count) { $outcome='INCOMPLETE' }
    $result=[ordered]@{schema=2;outcome=$outcome;startedAt=$started.ToString('o');finishedAt=[DateTime]::UtcNow.ToString('o');commit=$commit;requestedRef=$Ref;windows=[Environment]::OSVersion.VersionString;architecture=$env:PROCESSOR_ARCHITECTURE;logicalProcessors=[Environment]::ProcessorCount;powershell=$PSVersionTable.PSVersion.ToString();fatal=$fatal;steps=@($steps.ToArray());requiredNativeTestsMissing=$missingTests;requiredGoRetestsMissing=$missingRetests;goTestEvents=@($events.ToArray());browserSummary=$browserSummary;notCovered=$uncovered}
    $result['runMode'] = $(if ($Followup) { 'followup' } elseif ($Retest) { 'retest' } elseif ($Remaining) { 'remaining' } else { 'full' })
    Save-Text (Join-Path $report 'report.json') ($result | ConvertTo-Json -Depth 30)
    $lines=@('# Blora Windows validation', '', "Result: **$outcome**", "Commit: $commit", "UTC: $($result.startedAt) to $($result.finishedAt)", '', 'This is an automated evidence bundle, not full Windows/platform acceptance.', '', '| Stage | Result | Seconds | Log |','|---|---|---:|---|')
    foreach ($s in $steps) { $lines+="| $($s.name) | $($s.status) | $($s.seconds) | $($s.log) |" }
    $lines+=@('', '## Required native tests without PASS', ($missingTests -join "`n"), '', '## Skipped Go tests')
    foreach ($s in $skips) { $lines+="- $($s.stage): $($s.package) / $($s.test)" }
    $lines+=@('', '## Not covered'); foreach ($gap in $uncovered) { $lines+="- $gap" }
    if ($Retest -or $Followup) {
        $lines+=@('', '## Required Go retests without PASS'); foreach ($case in $missingRetests) { $lines+="- $($case.package) / $($case.test)" }
        $lines+=@('', '## Required browser retests without PASS'); foreach ($summary in $browserSummary) { foreach ($title in $summary.requiredRetestTitlesMissing) { $lines+="- $($summary.browser): $title" } }
    }
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
    $env:Path=$savedPath
    Write-Host "`nResult: $outcome`nSummary: $(Join-Path $report 'README.md')" -ForegroundColor Cyan
    if ($zipReady) { Write-Host "Send this ZIP: $zip" -ForegroundColor Cyan }
}
if ($outcome -eq 'FAILED') { exit 1 }
if ($outcome -eq 'INCOMPLETE') { exit 2 }
exit 0
