function Get-CloudCliIdentity {
    param([string] $Path)

    $result = Invoke-CliProcess -Path $Path -ArgumentList @('--version') -TimeoutSeconds 5 -StdoutLimit 64KB -StderrLimit 64KB
    if ($result.ExitCode -ne 0 -or $result.Stdout.Trim() -cnotmatch '\Azstack-cli version: ([A-Za-z0-9._+-]+), commit: ([A-Za-z0-9._+-]+)\z') {
        throw 'CloudAutomation.CliVersionMismatch: Unable to verify the CLI version and commit.'
    }
    return @{ version = $Matches[1]; commit = $Matches[2] }
}

function Assert-CloudCliHelp {
    param([string] $Path, [string] $Capability, [string[]] $Command, [string[]] $Flags, [string[]] $Fields)

    $help = Invoke-CliProcess -Path $Path -ArgumentList ($Command + '--help') -TimeoutSeconds 5 -StdoutLimit 64KB -StderrLimit 64KB
    $missing = @($Flags | Where-Object { $help.Stdout -cnotmatch ('(?m)^\s+(?:-\w,\s+)?--' + $_ + '(?:\s|$)') })
    $missingFields = @($Fields | Where-Object { $help.Stdout -cnotmatch ('\b' + $_ + '\b') })
    if ($help.ExitCode -ne 0 -or $missing.Count -or $missingFields.Count) {
        throw "CloudAutomation.CliCapabilityMismatch: $Capability is missing from the CLI ($($Command -join ' '))."
    }
}

function Assert-CloudCliCapability {
    param([string] $Path, [string] $Capability)

    switch -CaseSensitive ($Capability) {
        'passwordStdin' { Assert-CloudCliHelp $Path $Capability @('login') 'password-stdin' }
        'accessKeySecretStdin' { Assert-CloudCliHelp $Path $Capability @('login') 'access-key-secret-stdin' }
        'startResultV1' { Assert-CloudCliHelp $Path $Capability @('instance', 'start') @('uuid-only', 'request-timeout', 'output', 'yes') }
        'vmMemoryBytes' { Assert-CloudCliHelp $Path $Capability @('get', 'instances') 'all-pages' 'memorySizeBytes' }
        'resourceInventoryV1' {
            $inventories = [ordered]@{
                'disks' = @('sizeBytes', 'actualSizeBytes')
                'hosts' = @('totalCpuHz', 'availableCpuHz', 'totalMemoryBytes', 'availableMemoryBytes', 'cpuCount', 'cpuSocketCount')
                'clusters' = @(); 'zones' = @(); 'nics' = @()
                'primary-storages' = @('totalCapacityBytes', 'availableCapacityBytes', 'totalPhysicalCapacityBytes', 'availablePhysicalCapacityBytes', 'systemUsedCapacityBytes')
                'image-storages' = @('totalCapacityBytes', 'availableCapacityBytes')
                'images' = @('sizeBytes', 'actualSizeBytes')
                'instance-offerings' = @('memorySizeBytes', 'cpuCount')
                'disk-offerings' = @('diskSizeBytes')
                'l2-networks' = @('virtualNetworkId')
                'l3-networks' = @(); 'ip-ranges' = @()
            }
            foreach ($inventory in $inventories.GetEnumerator()) {
                Assert-CloudCliHelp $Path $Capability @('get', $inventory.Key) @('all-pages', 'fields', 'request-timeout', 'output') $inventory.Value
            }
            Assert-CloudCliHelp $Path $Capability @('describe', 'instance') @('uuid-only', 'request-timeout', 'output')
        }
        default { throw 'CloudAutomation.UnsupportedCapability: Unknown package capability.' }
    }
}
