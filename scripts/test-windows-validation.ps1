# Function-level tests runnable in PowerShell on Linux; not Windows acceptance.
$ErrorActionPreference='Stop'
# Reproduce the user's broken Windows progress renderer: neither the runner nor
# its report packer may invoke these host-dependent commands.
function Write-Progress { throw 'Injected progress renderer IndexOutOfRangeException' }
function Compress-Archive { throw 'Archive progress renderer must not be invoked' }
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
$savedPath=$env:Path; $Retest=$false; $Followup=$false; $BrowsersOnly=$false; $WebKitOnly=$false; $retestPlan=Get-RetestPlan
$uncovered=@('harness only'); $ProgressPreference='SilentlyContinue'
$Remaining=$true
# Test only: the timeout case runs the Start-Sleep cmdlet in the owned shell,
# not an external grandchild. Windows production uses taskkill /T /F.
function taskkill.exe { param([Parameter(ValueFromRemainingArguments=$true)]$Arguments)
    $index=[Array]::IndexOf($Arguments,'/PID'); Stop-Process -Id ([int]$Arguments[$index+1]) -Force
}
try {
    $value='space path & literal $() ` quote " and 中文'
    if (!(Invoke-Stage 'arguments' 'Write-Output' @($value) $run 10)) { throw 'Argument invocation failed' }
    if (!(Invoke-Stage 'child-progress' 'Write-Progress' @('test progress') $run 10)) { throw 'Child progress not suppressed' }
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
    if ($result.runMode -ne 'remaining') { throw 'Remaining mode not recorded' }
    if ($result.outcome -ne 'FAILED' -or !$result.requiredNativeTestsMissing.Count) { throw 'False success summary' }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $zip=[IO.Compression.ZipFile]::OpenRead((Join-Path $run 'Blora-Windows-Report.zip'))
    try { if ($zip.Entries.FullName -match 'private-credentials') { throw 'Private source entered ZIP' }; if (!($zip.Entries.FullName -contains 'report.json')) { throw 'Report missing from ZIP' } } finally { $zip.Dispose() }
    $requiredNames=@($result.requiredNativeTestsMissing)
    $steps.Clear(); $events.Clear()
    $steps.Add([pscustomobject]@{name='harness-pass';status='PASS';exitCode=0;seconds=0;detail='';log=''})
    foreach ($name in $requiredNames) { $events.Add([pscustomobject]@{stage='windows-native';test=$name;package='harness';action='pass';elapsed=0}) }
    $events.Add([pscustomobject]@{stage='go-all';test='TestHarness';package='harness';action='pass';elapsed=0})
    $events.Add([pscustomobject]@{stage='go-race';test='TestHarness';package='harness';action='pass';elapsed=0})
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
    # Retest mode cannot claim success from a nonempty but incomplete filter,
    # another package's identically named case, or browser aggregate counts.
    $steps.Clear(); $events.Clear(); $Retest=$true
    $steps.Add([pscustomobject]@{name='harness-pass';status='PASS';exitCode=0;seconds=0;detail='';log=''})
    foreach ($name in $requiredNames) { $events.Add([pscustomobject]@{stage='windows-native';test=$name;package='harness';action='pass';elapsed=0}) }
    foreach ($case in $retestPlan.goCases) { $events.Add([pscustomobject]@{stage='go-retest';test=$case.test;package=$case.package;action='pass';elapsed=0}) }
    $events.Add([pscustomobject]@{stage='go-race';test='TestHarness';package='harness';action='pass';elapsed=0})
    $browserData=@{stats=@{expected=$retestPlan.browserCaseCount;unexpected=0;flaky=0;skipped=0};suites=@(@{title='nested file';suites=@(@{title='nested group';specs=@($retestPlan.browserTitles | ForEach-Object { @{title=$_;ok=$true;tests=@(@{expectedStatus='passed';status='expected';results=@(@{status='passed'})})} })})})}
    foreach ($browser in @('chromium','firefox','webkit')) { Save-Text (Join-Path $report "$browser.json") ($browserData | ConvertTo-Json -Depth 20) }
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.runMode -ne 'retest' -or $result.outcome -ne 'AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS' -or $result.requiredGoRetestsMissing.Count -or @($result.browserSummary | Where-Object { $_.requiredRetestTitlesMissing.Count }).Count) { throw 'Valid retest summary incorrect' }
    $events[($events.Count-2)].package='wrong-package'
    $browserData.suites[0].suites[0].specs[0].tests[0].results=@(@{status='failed'},@{status='passed'})
    Save-Text (Join-Path $report 'firefox.json') ($browserData | ConvertTo-Json -Depth 20)
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'INCOMPLETE' -or $result.requiredGoRetestsMissing.Count -ne 1 -or @($result.browserSummary | Where-Object browser -eq 'firefox')[0].requiredRetestTitlesMissing.Count -ne 1) { throw 'Incomplete retest or flaky browser accepted as pass' }
    # Followup scope: all ten browser scenarios, 30 WebKit executions,
    # and three executions of each named Go regression remain required.
    $steps.Clear(); $events.Clear(); $Retest=$false; $Followup=$true; $retestPlan=Get-FollowupPlan
    if ($retestPlan.goCases.Count -ne 3 -or $retestPlan.browserFiles.Count -ne 3 -or $retestPlan.browserTitles.Count -ne 10) { throw 'Followup scope changed unexpectedly' }
    $steps.Add([pscustomobject]@{name='harness-pass';status='PASS';exitCode=0;seconds=0;detail='';log=''})
    foreach ($name in $requiredNames) { $events.Add([pscustomobject]@{stage='windows-native';test=$name;package='harness';action='pass';elapsed=0}) }
    foreach ($case in $retestPlan.goCases) { foreach ($execution in 1..3) { $events.Add([pscustomobject]@{stage='go-retest';test=$case.test;package=$case.package;action='pass';elapsed=0}) } }
    $events.Add([pscustomobject]@{stage='go-race';test='TestHarness';package='harness';action='pass';elapsed=0})
    foreach ($browser in @('chromium','firefox','webkit')) {
        $executions=$(if ($browser -eq 'webkit') { 3 } else { 1 })
        $specs=@(foreach ($execution in 1..$executions) { foreach ($title in $retestPlan.browserTitles) { @{title=$title;ok=$true;tests=@(@{expectedStatus='passed';status='expected';results=@(@{status='passed'})})} } })
        $browserData=@{stats=@{expected=($retestPlan.browserCaseCount*$executions);unexpected=0;flaky=0;skipped=0};suites=@(@{specs=$specs})}
        if ($browser -eq 'webkit') {
            $sample=@{signals=@(@{kind='requestfailed';path='/api/v1/tasks'});surface=@{visibility='visible';paint=$false}} | ConvertTo-Json -Depth 10
            $browserData.suites[0].specs[0].tests[0].results[0]['attachments']=@(@{name='blora-runtime-diagnostics';contentType='application/json';body=[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($sample))},@{name='private-credentials';contentType='application/json';path='/do/not/read/private.json'})
        }
        Save-Text (Join-Path $report "$browser.json") ($browserData | ConvertTo-Json -Depth 20)
    }
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.runMode -ne 'followup' -or $result.outcome -ne 'AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS') { throw 'Valid followup summary incorrect' }
    if (!(Test-Path (Join-Path $report 'browser-webkit-diagnostic-1.json')) -or (Test-Path (Join-Path $report 'private-credentials.json'))) { throw 'Diagnostic bundling boundary failed' }
    # An aggregate claiming 30 successes cannot hide a missing third
    # execution of one particular required scenario.
    $completeWebkitSpecs=@($browserData.suites[0].specs)
    $browserData.suites[0].specs=@($completeWebkitSpecs[1..($completeWebkitSpecs.Count-1)])
    Save-Text (Join-Path $report 'webkit.json') ($browserData | ConvertTo-Json -Depth 20)
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'INCOMPLETE' -or @($result.browserSummary | Where-Object browser -eq 'webkit')[0].requiredRetestTitlesMissing.Count -ne 1) { throw 'Missing third browser execution counted as passed' }
    $browserData.suites[0].specs=$completeWebkitSpecs
    # Nominal aggregate success with only one WebKit execution per case, or
    # two Go passes instead of three, must not satisfy this followup plan.
    $events.RemoveAt($events.Count-2)
    $browserData.stats.expected=$retestPlan.browserCaseCount; Save-Text (Join-Path $report 'webkit.json') ($browserData | ConvertTo-Json -Depth 20)
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'INCOMPLETE' -or $result.requiredGoRetestsMissing.Count -ne 1 -or !@($result.steps | Where-Object name -eq 'browser-webkit-case-count').Count) { throw 'Incomplete followup counted as passed' }
    # Browser-only selection must explicitly omit Go/native/SDK; old passes
    # are not needed or imported. Browser completeness remains mandatory.
    $steps.Clear(); $events.Clear(); $BrowsersOnly=$true
    $steps.Add([pscustomobject]@{name='harness-pass';status='PASS';exitCode=0;seconds=0;detail='';log=''})
    foreach ($browser in @('chromium','firefox','webkit')) {
        $executions=$(if ($browser -eq 'webkit') { 3 } else { 1 })
        $specs=@(foreach ($execution in 1..$executions) { foreach ($title in $retestPlan.browserTitles) { @{title=$title;ok=$true;tests=@(@{expectedStatus='passed';status='expected';results=@(@{status='passed'})})} } })
        $browserData=@{stats=@{expected=($retestPlan.browserCaseCount*$executions);unexpected=0;flaky=0;skipped=0};suites=@(@{specs=$specs})}
        Save-Text (Join-Path $report "$browser.json") ($browserData | ConvertTo-Json -Depth 20)
    }
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.runMode -ne 'followup-browser' -or $result.outcome -ne 'AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS' -or $result.goChecksSelected -or $result.nativeChecksSelected -or $result.sdkChecksSelected -or $result.goTestEvents.Count -or $result.requiredNativeTestsMissing.Count -or $result.requiredGoRetestsMissing.Count) { throw 'Browser-only selection imported or required omitted Go evidence' }
    $browserData.stats.expected=$retestPlan.browserCaseCount; Save-Text (Join-Path $report 'webkit.json') ($browserData | ConvertTo-Json -Depth 20)
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'INCOMPLETE' -or !@($result.steps | Where-Object name -eq 'browser-webkit-case-count').Count) { throw 'Browser-only selection accepted an incomplete WebKit result' }
    # Narrow selection still requires every named case three times. Neither
    # old Go results nor reports for omitted engines may become new passes.
    $steps.Clear(); $events.Clear(); $WebKitOnly=$true; $retestPlan=Get-WebKitFollowupPlan
    if ($retestPlan.browserCaseCount -ne 5 -or $retestPlan.browserTitles.Count -ne 5 -or $retestPlan.browserFiles.Count -ne 2 -or $retestPlan.goCases.Count) { throw 'WebKit-only scope changed unexpectedly' }
    $pattern=Get-BrowserTitlePattern $retestPlan
    foreach ($title in (Get-FollowupPlan).browserTitles) {
        if (("query-navigation.spec.ts $title" -match $pattern) -ne ($retestPlan.browserTitles -contains $title)) { throw 'WebKit title pattern selects the wrong scenario' }
    }
    foreach ($title in $retestPlan.browserTitles) {
        if (!("webkit terminal.spec.ts $title" -match $pattern) -or ("query-navigation.spec.ts prefix $title" -match $pattern) -or ("query-navigation.spec.ts $title suffix" -match $pattern) -or ("unrelated.spec.ts $title" -match $pattern)) { throw 'WebKit title/file pattern is not anchored' }
    }
    Remove-Item (Join-Path $report 'chromium.json'),(Join-Path $report 'firefox.json')
    $steps.Add([pscustomobject]@{name='harness-pass';status='PASS';exitCode=0;seconds=0;detail='';log=''})
    $specs=@(foreach ($execution in 1..3) { foreach ($title in $retestPlan.browserTitles) { @{title=$title;ok=$true;tests=@(@{expectedStatus='passed';status='expected';results=@(@{status='passed'})})} } })
    $browserData=@{stats=@{expected=15;unexpected=0;flaky=0;skipped=0};suites=@(@{specs=$specs})}
    Save-Text (Join-Path $report 'webkit.json') ($browserData | ConvertTo-Json -Depth 20)
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.runMode -ne 'followup-webkit' -or $result.outcome -ne 'AUTOMATED_CHECKS_PASSED_WITH_COVERAGE_GAPS' -or @($result.selectedBrowsers).Count -ne 1 -or $result.selectedBrowsers[0] -ne 'webkit' -or $result.browserSummary.Count -ne 1 -or $result.browserSummary[0].expectedCaseCount -ne 15 -or $result.goChecksSelected -or $result.nativeChecksSelected -or $result.sdkChecksSelected -or $result.goTestEvents.Count -or $result.requiredNativeTestsMissing.Count -or $result.requiredGoRetestsMissing.Count) { throw 'WebKit-only selection imported or required omitted evidence' }
    $browserData.suites[0].specs=@($specs[1..($specs.Count-1)])
    Save-Text (Join-Path $report 'webkit.json') ($browserData | ConvertTo-Json -Depth 20)
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'INCOMPLETE' -or $result.browserSummary[0].requiredRetestTitlesMissing.Count -ne 1) { throw 'WebKit-only selection accepted a missing third execution' }
    $steps.Clear(); $steps.Add([pscustomobject]@{name='harness-pass';status='PASS';exitCode=0;seconds=0;detail='';log=''})
    $browserData.suites[0].specs=$specs; $browserData.stats.expected=14
    Save-Text (Join-Path $report 'webkit.json') ($browserData | ConvertTo-Json -Depth 20)
    Remove-Item (Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'INCOMPLETE' -or !@($result.steps | Where-Object name -eq 'browser-webkit-case-count').Count) { throw 'WebKit-only selection accepted an incomplete aggregate' }
    $steps.Clear(); $steps.Add([pscustomobject]@{name='harness-pass';status='PASS';exitCode=0;seconds=0;detail='';log=''})
    Remove-Item (Join-Path $report 'webkit.json'),(Join-Path $run 'Blora-Windows-Report.zip')
    & ([scriptblock]::Create($final.Finally.Extent.Text.Trim().Substring(1,$final.Finally.Extent.Text.Trim().Length-2)))
    $result=Get-Content (Join-Path $report 'report.json') -Raw | ConvertFrom-Json
    if ($result.outcome -ne 'INCOMPLETE' -or !@($result.steps | Where-Object name -eq 'browser-webkit-report').Count -or @($result.steps | Where-Object name -match 'browser-(chromium|firefox)').Count) { throw 'WebKit-only selection misclassified missing reports' }
    # Verify compiler extraction with tiny fixtures; no compiler is downloaded
    # or executed by the harness. Validate the helper's complete syntax too.
    $compilerAst=[Management.Automation.Language.Parser]::ParseFile((Join-Path $PSScriptRoot 'windows-race-compiler.ps1'),[ref]$tokens,[ref]$errors)
    if ($errors.Count) { throw ($errors | Out-String) }
    foreach ($f in $compilerAst.FindAll({param($n) $n -is [Management.Automation.Language.FunctionDefinitionAst]},$false)) { Invoke-Expression $f.Extent.Text }
    function New-CompilerFixture([string]$File,[string]$Entry) {
        $zip=[IO.Compression.ZipFile]::Open($File,[IO.Compression.ZipArchiveMode]::Create)
        try { $stream=New-Object IO.StreamWriter($zip.CreateEntry($Entry).Open()); try { $stream.Write('not executable, test fixture') } finally { $stream.Dispose() } } finally { $zip.Dispose() }
    }
    $fixture=Join-Path $run 'compiler-fixture.zip'; New-CompilerFixture $fixture 'mingw64/bin/gcc.exe'
    $digest=(Get-FileHash $fixture -Algorithm SHA256).Hash
    $compilerRoot=Join-Path $run 'compiler-output'
    $rejected=$false
    try { Expand-CompilerArchive $fixture $compilerRoot ('0'*64) } catch { $rejected=$true }
    if (!$rejected -or (Test-Path $compilerRoot)) { throw 'Bad compiler checksum extracted files' }
    Expand-CompilerArchive $fixture $compilerRoot $digest
    if (![IO.File]::ReadAllText((Join-Path $compilerRoot 'mingw64/bin/gcc.exe')).Contains('test fixture')) { throw 'Verified compiler extraction failed' }
    $badFixture=Join-Path $run 'compiler-traversal.zip'; New-CompilerFixture $badFixture 'mingw64/../../escaped.txt'
    $rejected=$false
    try { Expand-CompilerArchive $badFixture (Join-Path $run 'bad-compiler-output') (Get-FileHash $badFixture -Algorithm SHA256).Hash } catch { $rejected=$true }
    if (!$rejected -or (Test-Path (Join-Path $run 'escaped.txt'))) { throw 'Compiler traversal allowed' }
    $previousOS=$env:OS
    try {
        $env:OS='Windows_NT'
        $shell=Join-Path $PSHOME 'pwsh'; if (!(Test-Path $shell)) { $shell=Join-Path $PSHOME 'powershell.exe' }
        & $shell -NoProfile -File $source -WorkRoot $run -CollectRun $run
        if ($LASTEXITCODE -ne 0) { throw 'Recovery entry failed' }
        $recovered=@(Get-ChildItem $run -Filter 'Blora-Windows-Recovered-*.zip')
        if ($recovered.Count -ne 1) { throw 'Recovered ZIP missing' }
        $zip=[IO.Compression.ZipFile]::OpenRead($recovered[0].FullName)
        try { if (!($zip.Entries.FullName -contains 'RECOVERY.txt') -or !($zip.Entries.FullName -contains 'report.json')) { throw 'Recovery evidence missing' } } finally { $zip.Dispose() }
        $collection=Join-Path $run 'collection'
        foreach ($name in @('20260101-010101-old','20260102-010101-new')) {
            $folder=Join-Path $collection "$name/report"
            [void](New-Item -ItemType Directory -Path $folder -Force)
            Save-Text (Join-Path $folder 'steps-so-far.json') '[]'
        }
        & $shell -NoProfile -File $source -WorkRoot $collection -CollectLatest
        if ($LASTEXITCODE -ne 0 -or @(Get-ChildItem (Join-Path $collection '20260102-010101-new') -Filter '*.zip').Count -ne 1 -or @(Get-ChildItem (Join-Path $collection '20260101-010101-old') -Filter '*.zip').Count -ne 0) { throw 'CollectLatest selected the wrong run' }
    } finally { $env:OS=$previousOS }
    Write-Host 'PASS: failing progress host, child progress suppression, literal arguments, failure/skip/timeout, full/remaining/retest/followup/browser-only/WebKit-only coverage gates, anchored title selection, diagnostic bundling, compiler checksum/traversal, ZIP boundaries and old-run recovery.'
} finally { Remove-Item -LiteralPath $run -Recurse -Force }
