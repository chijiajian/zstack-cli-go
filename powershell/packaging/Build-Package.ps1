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
    'Wait-CloudVM' = @('vmMemoryBytes', 'vmLifecycleV1')
}
$privateFiles = @(
    'private/process/Invoke-CliProcess.ps1', 'branding/Resolve-Message.ps1',
    'private/process/Convert-CliResult.ps1', 'private/process/Test-CliContract.ps1', 'private/connections/ConnectionRegistry.ps1',
    'private/objects/Convert-CloudVM.ps1', 'private/objects/Resolve-CloudVMTarget.ps1',
    'private/queries/ResourceInventory.ps1', 'private/queries/PlatformInventory.ps1',
    'private/queries/PlatformReports.ps1', 'private/queries/LongJob.ps1',
    'private/mutations/Invoke-CloudShouldProcess.ps1', 'private/mutations/Wait-CloudVMState.ps1',
    'private/mutations/Invoke-CloudVMOperation.ps1', 'private/mutations/Invoke-CloudDeviceWrite.ps1'
)
$helpSummaries = @{
    'Connect-CloudServer' = 'Connects to the API with an account credential or an access key.'
    'Disconnect-CloudServer' = 'Closes the connection and removes its local session files.'
    'Get-CloudVM' = 'Returns virtual machines by name or ID from the selected connection.'
    'Start-CloudVM' = 'Starts selected virtual machines and waits for them to run.'
    'Wait-CloudVM' = 'Waits for a virtual machine to reach the requested state.'
}
$lifecycleHelp = [ordered]@{
    'New-CloudVM' = 'Creates a VM from a root volume template and waits for the selected state.'
    'Set-CloudVM' = 'Updates specified VM metadata or stopped VM CPU and memory.'
    'Suspend-CloudVM' = 'Pauses a running VM in memory.'
    'Resume-CloudVM' = 'Resumes a paused VM.'
    'Stop-CloudVM' = 'Requests a graceful VM shutdown.'
    'Restart-CloudVM' = 'Requests a VM reboot and verifies its running state.'
    'Remove-CloudVM' = 'Deletes a stopped VM under a recoverable deletion policy.'
    'Restore-CloudVM' = 'Recovers a deleted VM to Stopped.'
}
foreach ($command in $lifecycleHelp.GetEnumerator()) {
    $entries[$command.Key] = $command.Key.Replace('-Cloud', '-') + '.ps1'
    $requirements[$command.Key] = 'vmLifecycleV1'
    $helpSummaries[$command.Key] = $command.Value
}
$inventoryHelp = [ordered]@{
    'Get-CloudHardDisk' = 'Returns disks attached to the supplied live VM objects.'
    'Get-CloudVolume' = 'Returns independent or attached volumes with exact capacities.'
    'Get-CloudVMHost' = 'Returns hypervisor hosts with connection state and resource capacities.'
    'Get-CloudCluster' = 'Returns clusters in the selected cloud or zone.'
    'Get-CloudZone' = 'Returns ZStack zones.'
    'Get-CloudNetworkAdapter' = 'Returns network adapters on the supplied live VM objects.'
    'Get-CloudPrimaryStorage' = 'Returns primary storage with logical and physical capacity.'
    'Get-CloudImageStorage' = 'Returns image storage with total and available capacity.'
    'Get-CloudImage' = 'Returns images including their media type and reported size.'
    'Get-CloudInstanceOffering' = 'Returns VM offerings with CPU and memory.'
    'Get-CloudDiskOffering' = 'Returns disk offerings and provisioned capacity.'
    'Get-CloudL2Network' = 'Returns Layer 2 networks in the selected cloud or zone.'
    'Get-CloudL3Network' = 'Returns Layer 3 networks in the selected cloud or Layer 2 network.'
    'Get-CloudIPRange' = 'Returns IP ranges in the selected cloud or Layer 3 network.'
}
foreach ($command in $inventoryHelp.GetEnumerator()) {
    $entries[$command.Key] = $command.Key.Replace('Get-Cloud', 'Get-') + '.ps1'
    $requirements[$command.Key] = 'resourceInventoryV1'
    $helpSummaries[$command.Key] = $command.Value
}
$deviceHelp = [ordered]@{
    'New-CloudVolume' = 'Creates an independent data volume.'
    'Remove-CloudVolume' = 'Deletes a detached data volume under a recoverable deletion policy.'
    'Restore-CloudVolume' = 'Recovers a deleted data volume.'
    'Add-CloudHardDisk' = 'Attaches an existing data volume to a VM.'
    'Remove-CloudHardDisk' = 'Detaches a data volume and keeps the underlying volume.'
    'Set-CloudHardDisk' = 'Expands a virtual disk to the requested total capacity.'
    'Get-CloudVolumeSnapshot' = 'Returns individual volume snapshots.'
    'New-CloudVolumeSnapshot' = 'Creates a snapshot of a single volume.'
    'Restore-CloudVolumeSnapshot' = 'Reverts a single volume to a snapshot.'
    'Remove-CloudVolumeSnapshot' = 'Deletes a leaf volume snapshot.'
    'New-CloudNetworkAdapter' = 'Attaches an adapter on the selected L3 network.'
    'Remove-CloudNetworkAdapter' = 'Removes the selected VM adapter by exact UUID.'
    'Get-CloudCDDrive' = 'Returns the virtual drives and attached ISO IDs for a VM.'
    'Set-CloudCDDrive' = 'Inserts or ejects ISO media on the selected virtual drive.'
}
foreach ($command in $deviceHelp.GetEnumerator()) {
    $entries[$command.Key] = $command.Key.Replace('-Cloud', '-') + '.ps1'
    $requirements[$command.Key] = 'deviceLifecycleV1'
    $helpSummaries[$command.Key] = $command.Value
}
$platformHelp = [ordered]@{
    'Get-CloudUsage' = 'Returns capacity and sampled host load from the selected cloud.'
    'Invoke-CloudZql' = 'Runs a ZQL query or count and returns each server result set.'
    'Get-CloudLongJob' = 'Queries real cloud long jobs by name or ID.'
    'Wait-CloudLongJob' = 'Waits for a real cloud long job to reach a terminal state.'
    'Resume-CloudLongJob' = 'Resumes an explicitly selected suspended cloud long job.'
    'Get-CloudAccount' = 'Returns visible ZStack accounts.'
    'Get-CloudUser' = 'Returns visible ZStack account users.'
    'Get-CloudRole' = 'Returns visible roles.'
    'Get-CloudPolicy' = 'Returns visible policies and statement counts.'
    'Get-CloudProject' = 'Returns visible IAM2 projects.'
    'Get-CloudSecurityGroup' = 'Returns security groups.'
    'Get-CloudSecurityGroupRule' = 'Returns rules for the selected security groups or connection.'
    'Get-CloudVip' = 'Returns virtual IPs.'
    'Get-CloudLoadBalancer' = 'Returns load balancers, optionally scoped to selected virtual IPs.'
    'Get-CloudLoadBalancerListener' = 'Returns listeners for the selected load balancers or connection.'
    'Get-CloudLoadBalancerServerGroup' = 'Returns server groups for the selected load balancers or connection.'
}
foreach ($command in $platformHelp.GetEnumerator()) {
    $entries[$command.Key] = $command.Key.Replace('-Cloud', '-') + '.ps1'
    $requirements[$command.Key] = switch -Wildcard ($command.Key) {
        '*Usage' { 'usageV1' }
        '*Zql' { 'zqlV1' }
        '*LongJob' { 'longJobV1' }
        default { 'platformInventoryV1' }
    }
    $helpSummaries[$command.Key] = $command.Value
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
