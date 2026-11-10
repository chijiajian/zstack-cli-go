Set-StrictMode -Version Latest

$script:PackageRoot = $PSScriptRoot
$script:Package = Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot 'package.json') | ConvertFrom-Json -AsHashtable
$script:Exports = Get-Content -Raw -LiteralPath (Join-Path $PSScriptRoot 'exports.json') | ConvertFrom-Json -AsHashtable
if ($script:Package.schemaVersion -ne 1 -or $script:Exports.schemaVersion -ne 1) {
    throw 'Unsupported module package manifest.'
}
$script:VerifiedCli = $null
$script:VerifiedCapabilities = @{}

function Get-CloudPackagedCliPath {
    $architecture = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    if (-not $IsWindows -or $architecture -ne [System.Runtime.InteropServices.Architecture]::X64) {
        throw 'CloudAutomation.UnsupportedPlatform: This package requires Windows x64.'
    }
    return Join-Path $script:PackageRoot 'bin/windows-amd64/zstack-cli.exe'
}

function Get-CloudCliPath {
    param([string] $Capability)

    if ($script:Package.cli.capabilities -cnotcontains $Capability) {
        throw "CloudAutomation.UnsupportedCapability: $Capability"
    }
    if ($null -eq $script:VerifiedCli) {
        $path = Get-CloudPackagedCliPath
        if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
            throw 'CloudAutomation.MissingCli: Reinstall the complete module package.'
        }
        $hash = (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash
        if ($hash -ne $script:Package.cli.sha256) {
            throw 'CloudAutomation.CliIntegrity: The packaged CLI hash does not match.'
        }
        $result = Invoke-CliProcess -Path $path -ArgumentList @('--version') -TimeoutSeconds 5
        $expected = 'zstack-cli version: {0}, commit: {1}' -f $script:Package.cli.version, $script:Package.cli.commit
        if ($result.ExitCode -ne 0 -or $result.Stdout.Trim() -cne $expected) {
            throw 'CloudAutomation.CliVersionMismatch: Reinstall the complete module package.'
        }
        $script:VerifiedCli = $path
    }
    if (-not $script:VerifiedCapabilities.ContainsKey($Capability)) {
        $command, $flags = switch -CaseSensitive ($Capability) {
            'passwordStdin' { ,@('login'); 'password-stdin' }
            'accessKeySecretStdin' { ,@('login'); 'access-key-secret-stdin' }
            'startResultV1' { ,@('instance', 'start'); ,@('uuid-only', 'request-timeout') }
            'vmMemoryBytes' { ,@('get', 'instances'); 'all-pages' }
            default { throw 'CloudAutomation.UnsupportedCapability: Unknown package capability.' }
        }
        $help = Invoke-CliProcess -Path $script:VerifiedCli -ArgumentList ($command + '--help') -TimeoutSeconds 5
        $missing = @($flags | Where-Object { $help.Stdout -cnotmatch ('(?m)^\s+(?:-\w,\s+)?--' + $_ + '(?:\s|$)') })
        if ($help.ExitCode -ne 0 -or $missing.Count -gt 0) {
            throw "CloudAutomation.CliCapabilityMismatch: $Capability is missing from the packaged CLI."
        }
        $script:VerifiedCapabilities[$Capability] = $true
    }
    return $script:VerifiedCli
}

# Load only explicit paths confined to the package. Import itself launches no process.
foreach ($file in @($script:Exports.privateFiles) + @($script:Exports.commands | ForEach-Object { $_.file })) {
    if ($file -notmatch '\A(?:private|public|branding)/[A-Za-z0-9_/-]+\.ps1\z') {
        throw 'Invalid module source path.'
    }
    . (Join-Path $PSScriptRoot $file)
}
$ExecutionContext.SessionState.Module.OnRemove = {
    foreach ($connection in @($script:Connections.Values)) {
        try { Remove-CloudConnection $connection } catch { Write-Warning 'A local session could not be removed.' }
    }
}
$names = @($script:Exports.commands | ForEach-Object {
    if ($_.name -notmatch '\A[A-Za-z]+-[A-Za-z][A-Za-z0-9]+\z') { throw 'Invalid exported command name.' }
    $_.name
})
Export-ModuleMember -Function $names
