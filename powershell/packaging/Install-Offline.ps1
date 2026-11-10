#Requires -Version 7.4
[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string] $Archive,
    [Parameter(Mandatory)] [ValidatePattern('\A[A-Fa-f0-9]{64}\z')] [string] $SHA256,
    [string] $ModuleDirectory
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
if (-not $ModuleDirectory) {
    if (-not $IsWindows -or [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -ne 'X64') {
        throw 'The default module directory requires Windows x64. Supply -ModuleDirectory to stage the package elsewhere.'
    }
    $ModuleDirectory = Join-Path ([Environment]::GetFolderPath('MyDocuments')) 'PowerShell/Modules'
}
$source = (Resolve-Path -LiteralPath $Archive).Path
$root = [System.IO.Path]::GetFullPath($ModuleDirectory)
$null = [System.IO.Directory]::CreateDirectory($root)
$stage = Join-Path $root ('.cloud-install-' + [guid]::NewGuid().ToString('N'))
$null = [System.IO.Directory]::CreateDirectory($stage)
try {
    $copy = Join-Path $stage 'package.zip'
    Copy-Item -LiteralPath $source -Destination $copy
    if ((Get-FileHash -LiteralPath $copy -Algorithm SHA256).Hash -ne $SHA256) { throw 'Package checksum mismatch.' }
    $payload = Join-Path $stage 'payload'
    [System.IO.Compression.ZipFile]::ExtractToDirectory($copy, $payload)
    $modules = @(Get-ChildItem -LiteralPath $payload -Directory)
    if ($modules.Count -ne 1) { throw 'Package must contain one module.' }
    $versions = @(Get-ChildItem -LiteralPath $modules[0].FullName -Directory)
    if ($versions.Count -ne 1) { throw 'Package must contain one version.' }
    $moduleName = $modules[0].Name
    $version = $versions[0].Name
    if ($moduleName -notmatch '\A[A-Za-z][A-Za-z0-9]*(\.[A-Za-z][A-Za-z0-9]*)+\z' -or $version -notmatch '\A\d+\.\d+(\.\d+){0,2}\z') { throw 'Invalid module package layout.' }
    $versionRoot = $versions[0].FullName
    $manifest = Join-Path $versionRoot "$moduleName.psd1"
    $details = Import-PowerShellDataFile -LiteralPath $manifest
    $package = Get-Content -LiteralPath (Join-Path $versionRoot 'package.json') -Raw | ConvertFrom-Json -AsHashtable
    if ($package.brand.moduleName -cne $moduleName -or $package.version -cne $version -or $details.ModuleVersion -cne $version) {
        throw 'Package manifest does not match its module directory.'
    }
    $exports = Get-Content -LiteralPath (Join-Path $versionRoot 'exports.json') -Raw | ConvertFrom-Json -AsHashtable
    if ($package.schemaVersion -ne 1 -or $exports.schemaVersion -ne 1 -or $package.cli.platform -cne 'windows-amd64') {
        throw 'Unsupported module package manifest.'
    }
    if ($details.RootModule -cne "$moduleName.psm1") { throw 'Invalid root module in package manifest.' }
    $files = @($details.RootModule, 'branding/messages.en.json', 'branding/messages.zh-CN.json')
    foreach ($file in @($exports.privateFiles) + @($exports.commands | ForEach-Object { $_.file })) {
        if ($file -notmatch '\A(?:private|public|branding)/[A-Za-z0-9_/-]+\.ps1\z') { throw 'Invalid module source path.' }
        $files += $file
    }
    $compatibility = Join-Path $versionRoot 'Compatibility'
    if (Test-Path -LiteralPath $compatibility) {
        $files += "Compatibility/$moduleName.Compatibility.psd1", "Compatibility/$moduleName.Compatibility.psm1"
    }
    foreach ($file in $files) {
        if (-not (Test-Path -LiteralPath (Join-Path $versionRoot $file) -PathType Leaf)) {
            throw "Incomplete module package. Missing file: $file"
        }
    }
    $cli = Join-Path $versionRoot 'bin/windows-amd64/zstack-cli.exe'
    if ((Get-FileHash -LiteralPath $cli -Algorithm SHA256).Hash -ne $package.cli.sha256) { throw 'Bundled CLI checksum mismatch.' }
    $target = Join-Path $root "$moduleName/$version"
    if (Test-Path -LiteralPath $target) { throw 'Module version is already installed.' }
    $null = [System.IO.Directory]::CreateDirectory((Split-Path -Parent $target))
    [System.IO.Directory]::Move($versionRoot, $target)
    Join-Path $target "$moduleName.psd1"
} finally { Remove-Item -LiteralPath $stage -Recurse -Force }
