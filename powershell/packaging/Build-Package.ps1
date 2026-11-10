#Requires -Version 7.4
[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string[]] $BrandConfig,
    [Parameter(Mandatory)] [string] $CliPath,
    [Parameter(Mandatory)] [version] $ModuleVersion,
    [ValidatePattern('\A[A-Za-z0-9._+-]+\z')] [string] $CliVersion,
    [ValidatePattern('\A[A-Za-z0-9._+-]+\z')] [string] $CliCommit,
    [Parameter(Mandatory)] [string] $OutputDirectory
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$source = Split-Path -Parent $PSScriptRoot
. (Join-Path $source 'branding/Validate-Brand.ps1')
. (Join-Path $source 'private/process/Invoke-CliProcess.ps1')
. (Join-Path $source 'private/process/Test-CliContract.ps1')
$brands = @($BrandConfig | ForEach-Object { Read-CloudBrand $_ })
Assert-UniqueCloudBrands $brands
$binary = (Resolve-Path -LiteralPath $CliPath).Path
if (-not (Test-Path -LiteralPath $binary -PathType Leaf)) { throw 'CLI binary is missing.' }
$reader = [System.IO.BinaryReader]::new([System.IO.File]::OpenRead($binary))
try {
    if ($reader.ReadUInt16() -ne 0x5a4d) { throw 'Expected a Windows x64 CLI executable.' }
    $reader.BaseStream.Position = 0x3c
    $reader.BaseStream.Position = $reader.ReadUInt32()
    if ($reader.ReadUInt32() -ne 0x4550 -or $reader.ReadUInt16() -ne 0x8664) { throw 'Expected a Windows x64 CLI executable.' }
} finally { $reader.Dispose() }
$sha = (Get-FileHash -LiteralPath $binary -Algorithm SHA256).Hash
$contract = Get-Content -Raw -LiteralPath (Join-Path $source 'contracts/cli-capabilities.json') | ConvertFrom-Json -AsHashtable
$capabilities = @($contract.capabilities.Keys | Where-Object { $contract.capabilities[$_].sourceStatus -eq 'verified-locally' })
if (-not $IsWindows) { throw 'Build the distributable package on Windows to verify the bundled CLI.' }
$identity = Get-CloudCliIdentity $binary
if (($PSBoundParameters.ContainsKey('CliVersion') -and $identity.version -cne $CliVersion) -or
    ($PSBoundParameters.ContainsKey('CliCommit') -and $identity.commit -cne $CliCommit)) {
    throw 'CloudAutomation.CliVersionMismatch: The supplied CLI does not match the requested version or commit.'
}
foreach ($capability in $capabilities) { Assert-CloudCliCapability $binary $capability }
$entries = [ordered]@{
    'Connect-CloudServer' = 'Connect-Server.ps1'
    'Disconnect-CloudServer' = 'Disconnect-Server.ps1'
    'Get-CloudVM' = 'Get-VM.ps1'
    'Start-CloudVM' = 'Start-VM.ps1'
    'Wait-CloudVM' = 'Wait-VM.ps1'
}
$requirements = @{
    'Connect-CloudServer' = @('passwordStdin', 'accessKeySecretStdin')
    'Disconnect-CloudServer' = 'passwordStdin'
    'Get-CloudVM' = 'vmMemoryBytes'
    'Start-CloudVM' = @('vmMemoryBytes', 'startResultV1')
    'Wait-CloudVM' = 'vmMemoryBytes'
}
$privateFiles = @(
    'private/process/Invoke-CliProcess.ps1', 'branding/Resolve-Message.ps1',
    'private/process/Convert-CliResult.ps1', 'private/process/Test-CliContract.ps1', 'private/connections/ConnectionRegistry.ps1',
    'private/objects/Convert-CloudVM.ps1', 'private/objects/Resolve-CloudVMTarget.ps1',
    'private/mutations/Invoke-CloudShouldProcess.ps1', 'private/mutations/Wait-CloudVMState.ps1'
)
$helpSummaries = @{
    'Connect-CloudServer' = 'Connects to the API with an account credential or an access key.'
    'Disconnect-CloudServer' = 'Closes the connection and removes its local session files.'
    'Get-CloudVM' = 'Returns virtual machines by name or ID from the selected connection.'
    'Start-CloudVM' = 'Starts selected virtual machines and waits for them to run.'
    'Wait-CloudVM' = 'Waits for a virtual machine to reach the requested state.'
}
$messages = @('branding/messages.en.json', 'branding/messages.zh-CN.json')
$root = [System.IO.Path]::GetFullPath($OutputDirectory)
$null = [System.IO.Directory]::CreateDirectory($root)
foreach ($brand in $brands) {
    $name = $brand.moduleName
    $version = $ModuleVersion.ToString()
    $target = Join-Path $root "$name/$version"
    $archive = Join-Path $root "$name-$version-windows-amd64.zip"
    if ((Test-Path -LiteralPath $target) -or (Test-Path -LiteralPath $archive)) { throw "Package version already exists: $name $version" }
    $staging = [System.IO.Directory]::CreateTempSubdirectory('cloud-package-').FullName
    try {
        $moduleRoot = Join-Path $staging "$name/$version"
        foreach ($file in $privateFiles + $messages) {
            $destination = Join-Path $moduleRoot $file
            $null = [System.IO.Directory]::CreateDirectory((Split-Path -Parent $destination))
            Copy-Item -LiteralPath (Join-Path $source $file) -Destination $destination
        }
        $null = [System.IO.Directory]::CreateDirectory((Join-Path $moduleRoot 'public'))
        $commands = @()
        foreach ($entry in $entries.GetEnumerator()) {
            $original = Join-Path $source ('public/' + $entry.Value)
            $missing = @($requirements[$entry.Key] | Where-Object { $capabilities -cnotcontains $_ })
            if (-not (Test-Path -LiteralPath $original -PathType Leaf) -or $missing.Count -gt 0) { continue }
            $function = $entry.Key.Replace('-Cloud', ('-' + $brand.commandPrefix))
            $file = 'public/' + $entry.Value
            $definition = Get-Content -Raw -LiteralPath $original
            if (-not $definition.StartsWith("function $($entry.Key) {", [StringComparison]::Ordinal)) { throw 'Unexpected public function declaration.' }
            $body = $definition.Substring("function $($entry.Key) {".Length)
            $help = @(
                "function $function {"
                '    # .SYNOPSIS'
                "    # $($helpSummaries[$entry.Key])"
                '    # .DESCRIPTION'
                "    # $($brand.displayName) ($name $version) uses CLI $($identity.version), commit $($identity.commit)."
                "    # Connection configuration namespace: $($brand.configNamespace)."
                '    # .LINK'
                "    # $([uri]::new($brand.helpUri).AbsoluteUri)"
            ) -join "`n"
            $definition = $help + $body
            Set-Content -LiteralPath (Join-Path $moduleRoot $file) -Value $definition -NoNewline -Encoding utf8
            $commands += @{ name = $function; file = $file }
        }
        $tokens = @{
            '__MODULE__' = $name; '__VERSION__' = $version; '__GUID__' = $brand.moduleGuid
            '__PUBLISHER__' = $brand.publisher.Replace("'", "''")
            '__DESCRIPTION__' = $brand.displayName.Replace("'", "''")
            '__PREFIX__' = $brand.commandPrefix
            '__EXPORTS__' = (@($commands | ForEach-Object { "'$($_.name)'" }) -join ', ')
        }
        $tokenPattern = '__[A-Z]+__'
        $replaceToken = [System.Text.RegularExpressions.MatchEvaluator] { param($match) return $tokens[$match.Value] }
        $manifest = Get-Content -Raw -LiteralPath (Join-Path $source 'module/Module.psd1.template')
        [regex]::Replace($manifest, $tokenPattern, $replaceToken) | Set-Content -LiteralPath (Join-Path $moduleRoot "$name.psd1") -NoNewline -Encoding utf8
        Copy-Item -LiteralPath (Join-Path $source 'module/Module.psm1') -Destination (Join-Path $moduleRoot "$name.psm1")
        @{ schemaVersion = 1; privateFiles = $privateFiles; commands = $commands } | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $moduleRoot 'exports.json') -Encoding utf8
        $manifest = [ordered]@{
            schemaVersion = 1; version = $version; brand = $brand
            cli = [ordered]@{ version = $identity.version; commit = $identity.commit; platform = 'windows-amd64'; sha256 = $sha; capabilities = $capabilities }
        } | ConvertTo-Json -Depth 8
        if (-not (Test-Json -Json $manifest -SchemaFile (Join-Path $PSScriptRoot 'package-manifest.schema.json'))) { throw 'Invalid package manifest.' }
        Set-Content -LiteralPath (Join-Path $moduleRoot 'package.json') -Value $manifest -Encoding utf8
        $bin = Join-Path $moduleRoot 'bin/windows-amd64'
        $null = [System.IO.Directory]::CreateDirectory($bin)
        $bundledCli = Join-Path $bin 'zstack-cli.exe'
        Copy-Item -LiteralPath $binary -Destination $bundledCli
        if ((Get-FileHash -LiteralPath $bundledCli -Algorithm SHA256).Hash -ne $sha) {
            throw 'The CLI binary changed during packaging.'
        }
        Copy-Item -LiteralPath (Join-Path (Split-Path -Parent $source) 'LICENSE') -Destination $moduleRoot
        if ($commands.Count -eq $entries.Count) {
            $compatibility = Join-Path $moduleRoot 'Compatibility'
            $null = [System.IO.Directory]::CreateDirectory($compatibility)
            foreach ($extension in 'psd1', 'psm1') {
                $template = Get-Content -Raw -LiteralPath (Join-Path $source "compatibility/Cloud.PowerCLI.Compatibility.$extension")
                [regex]::Replace($template, $tokenPattern, $replaceToken) | Set-Content -LiteralPath (Join-Path $compatibility "$name.Compatibility.$extension") -NoNewline -Encoding utf8
            }
        }
        $tempArchive = Join-Path $staging 'package.zip'
        [System.IO.Compression.ZipFile]::CreateFromDirectory($staging + "/$name", $tempArchive, [System.IO.Compression.CompressionLevel]::Optimal, $true)
        $null = [System.IO.Directory]::CreateDirectory((Split-Path -Parent $target))
        Move-Item -LiteralPath $moduleRoot -Destination $target
        Move-Item -LiteralPath $tempArchive -Destination $archive
        [pscustomobject]@{ ModuleName = $name; Version = $version; Directory = $target; Archive = $archive; SHA256 = (Get-FileHash -LiteralPath $archive).Hash }
    } finally { Remove-Item -LiteralPath $staging -Recurse -Force }
}
