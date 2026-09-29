# Function-level tests runnable in PowerShell on Linux; not Windows acceptance.
$ErrorActionPreference='Stop'
$tokens=$null; $errors=$null
$source=Join-Path $PSScriptRoot 'windows-validation.ps1'
$ast=[Management.Automation.Language.Parser]::ParseFile($source,[ref]$tokens,[ref]$errors)
if ($errors.Count) { throw ($errors | Out-String) }
foreach ($f in $ast.FindAll({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst]},$false)) { Invoke-Expression $f.Extent.Text }
$utf8=New-Object Text.UTF8Encoding($false)
$run=Join-Path ([IO.Path]::GetTempPath()) ('blora-runner-test-'+[guid]::NewGuid().ToString('N'))
$report=Join-Path $run 'report'; $repo=$run
[void](New-Item -ItemType Directory $report)
$steps=New-Object System.Collections.Generic.List[object]
$events=New-Object System.Collections.Generic.List[object]
$active=$null; $stage=0
$started=[DateTime]::UtcNow; $commit='test-only'; $Ref='test'; $fatal=''; $savedEnvironment=@{}
$uncovered=@('harness only'); $ProgressPreference='SilentlyContinue'
# Test only: the timeout case runs the Start-Sleep cmdlet in the owned shell,
# not an external grandchild. Windows production uses taskkill /T /F.
function taskkill.exe { param([Parameter(ValueFromRemainingArguments=$true)]$Arguments)
    $index=[Array]::IndexOf($Arguments,'/PID'); Stop-Process -Id ([int]$Arguments[$index+1]) -Force
}
try {
    $value='space path & literal $() ` quote " and 中文'
    if (!(Invoke-Stage 'arguments' 'Write-Output' @($value) $run 10)) { throw 'Argument invocation failed' }
    if (![IO.File]::ReadAllText((Join-Path $report '01-arguments.log')).Contains($value)) { throw 'Arguments changed' }
    if (Invoke-Stage 'failure' 'Write-Error' @('intentional failure') $run 10) { throw 'Failure counted as pass' }
    $event='{"Action":"skip","Package":"example","Test":"TestSkipped","Elapsed":0}'
    if (!(Invoke-Stage 'events' 'Write-Output' @($event) $run 10 -GoEvents)) { throw 'Event stage failed' }
    if ($events.Count -ne 1 -or $events[0].action -ne 'skip') { throw 'Skip not recorded' }
    if (Invoke-Stage 'timeout' 'Start-Sleep' @('5') $run 1) { throw 'Timeout counted as pass' }
    if ($steps[-1].status -ne 'TIMEOUT') { throw 'Timeout missing' }
    Save-Text (Join-Path $run 'private-credentials.json') 'do-not-bundle'
    $final=$ast.EndBlock.Statements | Where-Object { $_ -is [Management.Automation.Language.TryStatementAst] } | Select-Object -Last 1
    # Exercise the real finally block, including summary, checksums and ZIP.
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'FAILED' -or !$result.requiredNativeTestsMissing.Count) { throw 'False success summary' }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $zip=[IO.Compression.ZipFile]::OpenRead((Join-Path $run 'Blora-Windows-Report.zip'))
    try { if ($zip.Entries.FullName -match 'private-credentials') { throw 'Private source entered ZIP' }; if (!($zip.Entries.FullName -contains 'report.json')) { throw 'Report missing from ZIP' } } finally { $zip.Dispose() }
    $requiredNames=@($result.requiredNativeTestsMissing)
    $steps.Clear(); $events.Clear()
    $steps.Add([pscustomobject]@{name='harness-pass';status='PASS';exitCode=0;seconds=0;detail='';log=''})
    foreach ($name in $requiredNames) { $events.Add([pscustomobject]@{stage='windows-native';test=$name;package='harness';action='pass';elapsed=0}) }
    $events.Add([pscustomobject]@{stage='go-all';test='TestHarness';package='harness';action='pass';elapsed=0})
    foreach ($browser in @('chromium','firefox','webkit')) { Save-Text (Join-Path $report "$browser.json") '{"stats":{"expected":1,"unexpected":0,"flaky":0,"skipped":0}}' }
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS' -or $result.requiredNativeTestsMissing.Count) { throw 'Passing summary incorrect' }
    $events.Add([pscustomobject]@{stage='go-all';test='TestNeedsEngine';package='harness';action='skip';elapsed=0})
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'INCOMPLETE') { throw 'Skip did not prevent success' }
    Write-Host 'PASS: parser, literal arguments, failed command, skipped event, timeout, missing coverage, passing/incomplete/failed summaries and ZIP boundaries.'
} finally { Remove-Item -LiteralPath $run -Recurse -Force }
