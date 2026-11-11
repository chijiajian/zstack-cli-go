function Get-CloudCliIdentity {
    param([string] $Path)

    $result = Invoke-CliProcess -Path $Path -ArgumentList @('--version') -TimeoutSeconds 20 -StdoutLimit 64KB -StderrLimit 64KB
    if ($result.ExitCode -ne 0 -or $result.Stdout.Trim() -cnotmatch '\Azstack-cli version: ([A-Za-z0-9._+-]+), commit: ([A-Za-z0-9._+-]+)\z') {
        throw 'CloudAutomation.CliVersionMismatch: Unable to verify the CLI version and commit.'
    }
    return @{ version = $Matches[1]; commit = $Matches[2] }
}

function Assert-CloudCliHelp {
    param([string] $Path, [string] $Capability, [string[]] $Command, [string[]] $Flags, [string[]] $Fields)

    $help = Invoke-CliProcess -Path $Path -ArgumentList ($Command + '--help') -TimeoutSeconds 20 -StdoutLimit 64KB -StderrLimit 64KB
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
        'vmLifecycleV1' {
            $writeFlags = @('uuid-only', 'request-timeout', 'output', 'yes', 'dry-run')
            foreach ($operation in 'update', 'pause', 'resume', 'stop', 'restart', 'recover') {
                Assert-CloudCliHelp $Path $Capability @('instance', $operation) $writeFlags
            }
            Assert-CloudCliHelp $Path $Capability @('instance', 'stop') 'stop-ha'
            Assert-CloudCliHelp $Path $Capability @('delete', 'instance') $writeFlags
            Assert-CloudCliHelp $Path $Capability @('create', 'instance') ($writeFlags + @('resource-uuid', 'strategy', 'image', 'l3-network'))
            Assert-CloudCliHelp $Path $Capability @('wait', 'instance') @('uuid-only', 'for', 'timeout', 'interval')
        }
        'deviceLifecycleV1' {
            $writeFlags = @('uuid-only', 'request-timeout', 'output', 'yes', 'dry-run')
            $devices = [ordered]@{
                'create volume' = @('resource-uuid', 'size', 'disk-offering')
                'delete volume' = @()
                'volume attach' = @('instance')
                'volume detach' = @('instance')
                'volume recover' = @()
                'volume resize' = @('size')
                'snapshot create' = @('resource-uuid', 'name')
                'snapshot revert' = @('volume')
                'delete snapshot' = @('volume')
                'instance attach-l3' = @('l3-network')
                'instance detach-l3' = @('nic')
                'instance attach-iso' = @('cdrom', 'iso')
                'instance detach-iso' = @('cdrom')
            }
            foreach ($device in $devices.GetEnumerator()) {
                Assert-CloudCliHelp $Path $Capability ($device.Key -split ' ') ($writeFlags + $device.Value)
            }
            foreach ($inventory in 'snapshots', 'cdroms') {
                Assert-CloudCliHelp $Path $Capability @('get', $inventory) @('all-pages', 'fields', 'request-timeout', 'output')
            }
        }
        'usageV1' { Assert-CloudCliHelp $Path $Capability @('monitor', 'usage') @('section', 'output') }
        'zqlV1' { Assert-CloudCliHelp $Path $Capability @('zql') @('request-timeout', 'output') }
        'longJobV1' {
            Assert-CloudCliHelp $Path $Capability @('get', 'long-jobs') @('all-pages', 'fields', 'request-timeout', 'output')
            Assert-CloudCliHelp $Path $Capability @('wait', 'long-job') @('uuid-only', 'for', 'timeout', 'interval', 'output')
            Assert-CloudCliHelp $Path $Capability @('job', 'resume') @('uuid-only', 'request-timeout', 'output', 'yes')
        }
        'platformInventoryV1' {
            foreach ($inventory in 'accounts', 'users', 'roles', 'policies', 'projects', 'security-groups',
                'security-group-rules', 'vips', 'load-balancers', 'load-balancer-listeners', 'load-balancer-server-groups') {
                Assert-CloudCliHelp $Path $Capability @('get', $inventory) @('all-pages', 'fields', 'request-timeout', 'output')
            }
        }
        default { throw 'CloudAutomation.UnsupportedCapability: Unknown package capability.' }
    }
}
