# Portable wrapper checks, not Windows acceptance.
$ErrorActionPreference='Stop'
function Write-Progress { throw 'Progress renderer must not be invoked' }
function Compress-Archive { throw 'Host archive progress must not be invoked' }
$tokens=$null; $errors=$null
$ast=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot 'windows-device-validation.ps1'),[ref]$tokens,[ref]$errors)
if ($errors.Count) { throw ($errors | Out-String) }
foreach ($f in $ast.FindAll({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst]},$false)) { Invoke-Expression $f.Extent.Text }
$utf8=New-Object Text.UTF8Encoding($false)
$run=Join-Path ([IO.Path]::GetTempPath()) ('blora-device-wrapper-test-'+[guid]::NewGuid().ToString('N'))
$reportDir=Join-Path $run 'report'
[void](New-Item -ItemType Directory $reportDir)
$steps=New-Object 'System.Collections.Generic.List[object]'
$active=$null; $started=[DateTime]::UtcNow; $runId='harness-only'; $Ref='a'*40; $commit=$Ref; $failure='injected failure'; $exitCode=1
$runnerHash=(Get-FileHash -LiteralPath (Join-Path $PSScriptRoot 'windows-device-validation.ps1') -Algorithm SHA256).Hash.ToLowerInvariant()
try {
    $names=@('initialize-private-master','real-tls-static-login','enroll-real-daemon','create-stopped-owned-instance','real-user-permission-denial','real-file-save-and-read','real-backup-restore','real-monitor-sampling','real-extension-install-upgrade-remove','real-instance-start-stop','real-terminal-wss-input-resume-close','real-chromium-editor-refresh-save','stopped-state-process-restart','cleanup-owned-resources')
    $device=[pscustomobject]@{platform='windows/amd64';outcome='CHECKS_PASSED_WITH_SCOPE_LIMITS';checks=@($names | ForEach-Object { [pscustomobject]@{name=$_;status='PASS'} })}
    Assert-DeviceReport $device
    $complete=@($device.checks)
    foreach ($kind in @('missing','duplicate','failed-cleanup','wrong-platform')) {
        $device.checks=@($complete); $device.platform='windows/amd64'
        switch ($kind) {
            'missing' { $device.checks=@($complete[0..12]) }
            'duplicate' { $device.checks=@($complete)+@($complete[0]) }
            'failed-cleanup' { $device.checks=@($complete[0..12])+@([pscustomobject]@{name='cleanup-owned-resources';status='FAIL'}) }
            'wrong-platform' { $device.platform='linux/amd64' }
        }
        $rejected=$false
        try { Assert-DeviceReport $device } catch { $rejected=$true }
        if (!$rejected) { throw "Accepted invalid device evidence: $kind" }
    }
    $command=(Get-Process -Id $PID).Path
    $literal='space & literal $() ` quote '' and 中文'
    $argumentProbe=Join-Path $run 'argument-probe.ps1'
    Save-Text $argumentProbe 'param([string]$Value) Write-Output $Value'
    Invoke-DeviceStage 'literal-arguments' $command @('-NoProfile','-File',$argumentProbe,$literal) $run 15
    if (![IO.File]::ReadAllText((Join-Path $reportDir '01-literal-arguments.log')).Contains($literal)) { throw 'Child argument escaping changed the literal value' }
    Invoke-DeviceStage 'stderr-progress' $command @('-NoProfile','-Command','[Console]::Error.WriteLine("ordinary native progress"); exit 0') $run 15
    if (![IO.File]::ReadAllText((Join-Path $reportDir '02-stderr-progress.log')).Contains('ordinary native progress')) { throw 'Native stderr was lost or treated as failure' }
    $failed=$false
    try { Invoke-DeviceStage 'intentional-failure' $command @('-NoProfile','-Command','exit 7') $run 15 } catch { $failed=$true }
    if (!$failed -or $steps[-1].status -ne 'FAILED' -or $steps[-1].exitCode -ne 7) { throw 'Failed child was accepted' }
    $missing=$false
    try { Invoke-DeviceStage 'missing-command' 'blora-command-does-not-exist' @() $run 15 } catch { $missing=$true }
    if (!$missing -or $steps[-1].exitCode -ne 1 -or !$steps[-1].detail -or !(Get-Content (Join-Path $reportDir '04-missing-command.log') -Raw).Contains('blora-command-does-not-exist')) { throw 'Child shell error was not captured' }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    foreach ($kind in @('valid','traversal','wrong-ref','duplicate','symlink')) {
        $path=Join-Path $run ($kind+'.zip')
        $archive=[IO.Compression.ZipFile]::Open($path,[IO.Compression.ZipArchiveMode]::Create)
        try {
            $prefix='BloraPanel-'+$Ref+'/'
            $entryName=$(switch ($kind) { 'traversal' {$prefix+'../escape.txt'} 'wrong-ref' {'BloraPanel-'+('b'*40)+'/safe.txt'} default {$prefix+'scripts/safe.txt'} })
            $entry=$archive.CreateEntry($entryName)
            if ($kind -eq 'symlink') { $entry.ExternalAttributes=0xA000 -shl 16 }
            $writer=[IO.StreamWriter]::new($entry.Open())
            try { $writer.Write('owned source') } finally { $writer.Dispose() }
            if ($kind -eq 'duplicate') { [void]$archive.CreateEntry($entryName.ToUpperInvariant()) }
        } finally { $archive.Dispose() }
        $target=Join-Path $run ('source-'+$kind)
        $rejected=$false
        try { Expand-FixedSourceArchive $path $target $Ref } catch { $rejected=$true }
        if ($kind -eq 'valid') {
            if ($rejected -or (Get-Content (Join-Path $target 'scripts/safe.txt') -Raw) -ne 'owned source') { throw 'Valid fixed archive was rejected' }
        } elseif (!$rejected -or (Test-Path $target)) { throw "Unsafe archive created source: $kind" }
    }
    Save-Text (Join-Path $run 'private-credentials.json') 'do-not-bundle'
    $final=$ast.EndBlock.Statements | Where-Object { $_ -is [Management.Automation.Language.TryStatementAst] } | Select-Object -Last 1
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $reportDir 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'FAILED' -or $result.commit -ne $Ref -or $result.runnerSHA256 -ne $runnerHash -or $result.steps.Count -ne 4) { throw 'Final report lost failure/source/stages' }
    $archives=@(Get-ChildItem -LiteralPath $run -Filter 'Blora-Windows-Device-*.zip' -File)
    if ($archives.Count -ne 1) { throw 'Expected one current report archive' }
    $archive=[IO.Compression.ZipFile]::OpenRead($archives[0].FullName)
    try {
        if ($archive.Entries.FullName -match 'private-credentials') { throw 'Private state entered archive' }
        $reader=[IO.StreamReader]::new($archive.GetEntry('SHA256SUMS.txt').Open())
        try { $manifest=$reader.ReadToEnd() } finally { $reader.Dispose() }
        $expected=@{}
        foreach ($line in $manifest.TrimEnd() -split '\r?\n') {
            if ($line -notmatch '^([a-f0-9]{64})  (.+)$' -or $expected.ContainsKey($Matches[2])) { throw 'Bad checksum manifest' }
            $expected[$Matches[2]]=$Matches[1]
        }
        if ($expected.Count -ne $archive.Entries.Count-1) { throw 'Checksum coverage incomplete' }
        foreach ($entry in $archive.Entries) {
            if ($entry.FullName -eq 'SHA256SUMS.txt') { continue }
            $stream=$entry.Open(); $hasher=[Security.Cryptography.SHA256]::Create()
            try { $hash=[BitConverter]::ToString($hasher.ComputeHash($stream)).Replace('-','').ToLowerInvariant() } finally { $stream.Dispose(); $hasher.Dispose() }
            if ($expected[$entry.FullName] -ne $hash) { throw 'Archive checksum mismatch' }
        }
    } finally { $archive.Dispose() }
    Write-Host 'PASS: literal arguments, failure retention, report identity, private-state exclusion and exact archive checksums.'
} finally { Remove-Item -LiteralPath $run -Recurse -Force }
