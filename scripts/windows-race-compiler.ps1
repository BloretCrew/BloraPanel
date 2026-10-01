#requires -Version 5.1
# Portable test-only compiler. No elevation, system PATH or registry changes.
[CmdletBinding()]
param([Parameter(Mandatory=$true)][string]$Destination)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

function Expand-CompilerArchive([string]$Archive,[string]$Root,[string]$ExpectedSHA256) {
    if ((Get-FileHash -LiteralPath $Archive -Algorithm SHA256).Hash -ine $ExpectedSHA256) { throw 'Compiler archive SHA-256 mismatch; nothing will be extracted or executed.' }
    Add-Type -AssemblyName System.IO.Compression
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $prefix=[IO.Path]::GetFullPath($Root).TrimEnd([IO.Path]::DirectorySeparatorChar)+[IO.Path]::DirectorySeparatorChar
    $zip=[IO.Compression.ZipFile]::OpenRead($Archive)
    try {
        # Validate the complete manifest before writing anything. The official
        # ZIP must contain only the mingw64 tree, no links or traversal paths.
        $entries=New-Object System.Collections.Generic.List[object]
        $names=New-Object 'System.Collections.Generic.HashSet[string]' ([StringComparer]::OrdinalIgnoreCase)
        [long]$total=0
        foreach ($entry in $zip.Entries) {
            $name=$entry.FullName.Replace('\','/')
            if (!$name.StartsWith('mingw64/') -or $name.Contains(':') -or @($name.Split('/') | Where-Object { $_ -eq '..' -or $_ -eq '.' }).Count -or ((($entry.ExternalAttributes -shr 16) -band 0xF000) -eq 0xA000) -or ($entry.ExternalAttributes -band 0x400)) { throw 'Unexpected compiler ZIP path or linked entry.' }
            $target=[IO.Path]::GetFullPath((Join-Path $Root $name))
            if (!$target.StartsWith($prefix,[StringComparison]::OrdinalIgnoreCase) -or !$names.Add($target)) { throw 'Duplicate or out-of-root compiler ZIP entry.' }
            $total+=$entry.Length
            if ($total -gt 3GB -or $entries.Count -ge 50000) { throw 'Compiler ZIP exceeds extraction limits.' }
            $entries.Add([pscustomobject]@{entry=$entry;target=$target;directory=$name.EndsWith('/')})
        }
        $count=0
        foreach ($item in $entries) {
            if ($item.directory) { [void][IO.Directory]::CreateDirectory($item.target) }
            else {
                [void][IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($item.target))
                [IO.Compression.ZipFileExtensions]::ExtractToFile($item.entry,$item.target,$false)
            }
            $count++
            if ($count % 2000 -eq 0) { Write-Host "Compiler extraction: $count / $($entries.Count) files" }
        }
    } finally { $zip.Dispose() }
}

if ($env:OS -ne 'Windows_NT') { throw 'Portable Windows compiler setup must run on Windows.' }
if ($env:PROCESSOR_ARCHITECTURE -ne 'AMD64') { throw 'This pinned race compiler supports Windows amd64. Use an architecture-matching supported Go race toolchain on other platforms.' }
if (Test-Path -LiteralPath $Destination) { throw 'Compiler destination must be a new directory owned by this run.' }
[void][IO.Directory]::CreateDirectory($Destination)
# Official asset digest from the pinned release, not a checksum downloaded
# beside an untrusted archive. https://github.com/brechtsanders/winlibs_mingw/releases/tag/16.2.0posix-14.0.0-ucrt-r2
$url='https://github.com/brechtsanders/winlibs_mingw/releases/download/16.2.0posix-14.0.0-ucrt-r2/winlibs-x86_64-posix-seh-gcc-16.2.0-mingw-w64ucrt-14.0.0-r2.zip'
$sha256='d5dbafc4a170e762ca6143151ec918fb9e2c72736fb14cd704abebc6bdd5276a'
$archive=Join-Path $Destination 'compiler.zip'
Add-Type -AssemblyName System.Net.Http
[Net.ServicePointManager]::SecurityProtocol=[Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
$client=New-Object Net.Http.HttpClient
$client.Timeout=[TimeSpan]::FromMinutes(25)
$response=$null; $inputStream=$null; $outputStream=$null
try {
    Write-Host 'Downloading pinned WinLibs GCC 16.2.0 / MinGW-w64 14.0.0 UCRT (about 261 MB).'
    $response=$client.GetAsync($url,[Net.Http.HttpCompletionOption]::ResponseHeadersRead).GetAwaiter().GetResult()
    [void]$response.EnsureSuccessStatusCode()
    $length=$response.Content.Headers.ContentLength
    if ($length -and $length -gt 320MB) { throw 'Compiler download exceeds the expected bound.' }
    $inputStream=$response.Content.ReadAsStreamAsync().GetAwaiter().GetResult()
    $outputStream=[IO.File]::Open($archive,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)
    $buffer=New-Object byte[] (1MB); [long]$received=0; [long]$next=16MB
    while (($read=$inputStream.Read($buffer,0,$buffer.Length)) -gt 0) {
        $received+=$read
        if ($received -gt 320MB) { throw 'Compiler download exceeds the expected bound.' }
        $outputStream.Write($buffer,0,$read)
        if ($received -ge $next) { Write-Host "Compiler download: $([Math]::Round($received/1MB)) MB"; $next=$received+16MB }
    }
    if ($length -and $received -ne $length) { throw 'Compiler download was truncated.' }
} finally {
    if ($outputStream) { $outputStream.Dispose() }; if ($inputStream) { $inputStream.Dispose() }
    if ($response) { $response.Dispose() }; $client.Dispose()
}
Expand-CompilerArchive $archive $Destination $sha256
$gcc=Join-Path $Destination 'mingw64/bin/gcc.exe'
if (!(Test-Path -LiteralPath $gcc -PathType Leaf)) { throw 'Verified archive lacks gcc.exe.' }
$library=(& $gcc --print-file-name libsynchronization.a | Out-String).Trim()
if ($LASTEXITCODE -ne 0 -or !(Test-Path -LiteralPath $library -PathType Leaf)) { throw 'The compiler lacks the synchronization library required by Go race on Windows.' }
Remove-Item -LiteralPath $archive
Write-Host 'Portable race compiler ready. Only this validation process uses its PATH.'
