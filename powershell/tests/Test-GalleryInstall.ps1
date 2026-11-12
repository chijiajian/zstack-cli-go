#Requires -Version 7.4
[CmdletBinding()]
param(
    [Parameter(Mandatory)] [version] $ModuleVersion,
    [Parameter(Mandatory)] [ValidatePattern('\A[0-9a-f]{40}\z')] [string] $ExpectedCommit,
    [Parameter(Mandatory)] [string] $ReportPath
)

$ErrorActionPreference = 'Stop'
$name = 'ZStack.PowerCLI'
$version = $ModuleVersion.ToString()

if (Get-Module -ListAvailable -Name $name) {
    throw 'Gallery smoke test requires a clean Windows runner without ZStack.PowerCLI installed.'
}

$found = @(Find-PSResource -Name $name -Version $version -Repository PSGallery -ErrorAction Stop)
if ($found.Count -ne 1 -or $found[0].Name -cne $name -or $found[0].Version.ToString() -cne $version) {
    throw 'The requested module version was not found in PSGallery.'
}

Install-PSResource -Name $name -Version $version -Repository PSGallery -Scope CurrentUser -TrustRepository -AcceptLicense -ErrorAction Stop
$installed = @(Get-InstalledPSResource -Name $name -Version $version -ErrorAction Stop)
if ($installed.Count -ne 1 -or $installed[0].Version.ToString() -cne $version) {
    throw 'The requested module version was not installed.'
}

$module = Import-Module -Name $name -RequiredVersion $version -PassThru -ErrorAction Stop
if ($module.Version.ToString() -cne $version -or $module.ExportedFunctions.Count -lt 57) {
    throw 'The installed module version or command exports are incorrect.'
}
foreach ($command in 'Connect-ZStackServer', 'Get-ZStackVM', 'Get-ZStackUsage', 'Get-ZStackLongJob') {
    if (-not $module.ExportedFunctions.ContainsKey($command)) { throw "Missing exported command: $command" }
}

$manifestPath = Join-Path $module.ModuleBase "$name.psd1"
$null = Test-ModuleManifest -Path $manifestPath -ErrorAction Stop
$package = Get-Content -LiteralPath (Join-Path $module.ModuleBase 'package.json') -Raw | ConvertFrom-Json -AsHashtable
$cliPath = Join-Path $module.ModuleBase 'bin/windows-amd64/zstack-cli.exe'
if (-not (Test-Path -LiteralPath $cliPath -PathType Leaf)) { throw 'The Gallery package lacks its Windows CLI.' }
$cliHash = (Get-FileHash -LiteralPath $cliPath -Algorithm SHA256).Hash
if ($cliHash -cne $package.cli.sha256) { throw 'The bundled CLI does not match the package checksum.' }
$identity = & $cliPath version
if ($LASTEXITCODE -ne 0 -or $identity -cne "zstack-cli version: $version, commit: $ExpectedCommit") {
    throw 'The bundled CLI version or commit differs from the release tag.'
}

$noConnection = $null
try { Get-ZStackVM -ErrorAction Stop | Out-Null } catch { $noConnection = $_ }
if ($null -eq $noConnection -or $noConnection.Exception.Message -cnotmatch 'CloudAutomation\.NotConnected') {
    throw 'The installed Get-ZStackVM command did not report its expected disconnected state.'
}

$compatibilityPath = Join-Path $module.ModuleBase "Compatibility/$name.Compatibility.psd1"
$compatibility = Import-Module -Name $compatibilityPath -PassThru -ErrorAction Stop
if (-not $compatibility.ExportedFunctions.ContainsKey('Get-VM')) {
    throw 'The Gallery package lacks the opt-in compatibility commands.'
}

$report = [ordered]@{
    module = $name
    version = $version
    powershell = $PSVersionTable.PSVersion.ToString()
    platform = [System.Runtime.InteropServices.RuntimeInformation]::OSDescription
    source = 'PSGallery'
    exports = $module.ExportedFunctions.Count
    cliCommit = $ExpectedCommit
    cliSha256 = $cliHash
    compatibilityGetVM = $true
    disconnectedCommandCheck = $true
}
$report | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath $ReportPath -Encoding utf8
$report | ConvertTo-Json -Depth 4
