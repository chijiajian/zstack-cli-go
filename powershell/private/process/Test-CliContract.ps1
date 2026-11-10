function Get-CloudCliIdentity {
    param([string] $Path)

    $result = Invoke-CliProcess -Path $Path -ArgumentList @('--version') -TimeoutSeconds 5 -StdoutLimit 64KB -StderrLimit 64KB
    if ($result.ExitCode -ne 0 -or $result.Stdout.Trim() -cnotmatch '\Azstack-cli version: ([A-Za-z0-9._+-]+), commit: ([A-Za-z0-9._+-]+)\z') {
        throw 'CloudAutomation.CliVersionMismatch: Unable to verify the CLI version and commit.'
    }
    return @{ version = $Matches[1]; commit = $Matches[2] }
}

function Assert-CloudCliCapability {
    param([string] $Path, [string] $Capability)

    $field = $null
    $command, $flags = switch -CaseSensitive ($Capability) {
        'passwordStdin' { ,@('login'); 'password-stdin' }
        'accessKeySecretStdin' { ,@('login'); 'access-key-secret-stdin' }
        'startResultV1' { ,@('instance', 'start'); ,@('uuid-only', 'request-timeout', 'output', 'yes') }
        'vmMemoryBytes' { ,@('get', 'instances'); 'all-pages'; $field = 'memorySizeBytes' }
        default { throw 'CloudAutomation.UnsupportedCapability: Unknown package capability.' }
    }
    $help = Invoke-CliProcess -Path $Path -ArgumentList ($command + '--help') -TimeoutSeconds 5 -StdoutLimit 64KB -StderrLimit 64KB
    $missing = @($flags | Where-Object { $help.Stdout -cnotmatch ('(?m)^\s+(?:-\w,\s+)?--' + $_ + '(?:\s|$)') })
    if ($help.ExitCode -ne 0 -or $missing.Count -gt 0 -or ($field -and $help.Stdout -cnotmatch "\b$field\b")) {
        throw "CloudAutomation.CliCapabilityMismatch: $Capability is missing from the CLI."
    }
}
