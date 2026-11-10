function Convert-CloudVM {
    param($Inventories, $Connection)

    foreach ($Inventory in $Inventories) {
        if ($Inventory -isnot [System.Collections.IDictionary] -or $Inventory['uuid'] -isnot [string] -or
            $Inventory['uuid'] -notmatch '\A[A-Za-z0-9-]+\z' -or $Inventory['name'] -isnot [string]) {
            throw (New-CloudError -Operation QueryVM -Failure InvalidInventory -Target $Connection.Id)
        }
        $cpu = $Inventory['cpuNum']
        if (-not $Inventory.Contains('memorySizeBytes')) {
            throw (New-CloudError -Operation QueryVM -Failure MissingMemoryCapability -Target $Connection.Id)
        }
        $bytes = $Inventory['memorySizeBytes']
        if (($null -ne $cpu -and ($cpu -isnot [long] -and $cpu -isnot [int] -or $cpu -lt 0 -or $cpu -gt [int]::MaxValue)) -or
            ($null -ne $bytes -and ($bytes -isnot [long] -and $bytes -isnot [int] -or $bytes -lt 0))) {
            throw (New-CloudError -Operation QueryVM -Failure InvalidInventory -Target $Connection.Id)
        }
        $state = $Inventory['state']
        if ($null -ne $state -and $state -isnot [string]) {
            throw (New-CloudError -Operation QueryVM -Failure InvalidInventory -Target $Connection.Id)
        }
        $power = switch -CaseSensitive ($state) {
            'Running' { 'PoweredOn' }
            'Stopped' { 'PoweredOff' }
            default { 'Unknown' }
        }
        $vm = [pscustomobject]@{
            PSTypeName = 'Cloud.Automation.VM'
            Id = $Inventory['uuid']
            Name = $Inventory['name']
            NumCpu = if ($null -ne $cpu) { [int] $cpu }
            MemoryGB = if ($null -ne $bytes) { [double] $bytes / 1GB }
            PowerState = $power
            RawState = $state
            ConnectionId = $Connection.Id
            BrandId = $Connection.BrandId
            Generation = $Connection.Generation
        }
        $Connection.Objects.Add($vm.PSObject, $vm.Id)
        $vm
    }
}

function Read-CloudVMInventory {
    param($Connection, [string] $Id, [double] $TimeoutSeconds = 120)

    $null = Get-CloudCliPath 'vmMemoryBytes'
    $milliseconds = [Math]::Max(1, [long] [Math]::Floor($TimeoutSeconds * 1000))
    $arguments = @('get', 'instances', '--all-pages', '--sort', '+uuid', '--request-timeout', "${milliseconds}ms", '--fields', 'uuid,name,cpuNum,memorySizeBytes,state', '-o', 'json')
    if ($Id) { $arguments += @('-q', "uuid=$Id") }
    $result = Invoke-CloudConnectedCli $Connection $arguments $TimeoutSeconds
    $rows = Convert-CliResult $result 'QueryVM' -Shape Array -Target $Connection.Id
    if ($Id) {
        foreach ($row in $rows) {
            if ($row -isnot [System.Collections.IDictionary] -or $row['uuid'] -cne $Id) {
                throw (New-CloudError -Operation ResolveVM -Failure WrongIdentity -Target $Id)
            }
        }
    }
    return ,$rows
}

function Convert-CloudDetailWarning {
    param([string] $Warning)

    # Older CLI builds may append an API response to this prefix.
    if ($Warning -cmatch '\A((?:host|image|instanceOffering|l3-network) [A-Za-z0-9-]{1,128} unavailable:)(?: ([a-z_]+)\z)?') {
        $category = $Matches[2]
        if ($category -cnotin @('operation_failed', 'invalid_arguments', 'authentication_failed', 'not_found',
                'conflict', 'timeout', 'connection_failed', 'api_failed', 'output_failed', 'canceled')) {
            $category = 'operation_failed'
        }
        return "$($Matches[1]) $category"
    }
    return 'related resource unavailable: operation_failed'
}

function Read-CloudVMDetail {
    param($Connection, [string] $Id)

    $null = Get-CloudCliPath 'resourceInventoryV1'
    $result = Invoke-CloudConnectedCli $Connection @('describe', 'instance', $Id, '--uuid-only', '--request-timeout', '25s', '-o', 'json') 30
    $document = Convert-CliResult $result 'DescribeVM' -Target $Id
    $resource = $document['resource']
    $related = $document['related']
    if ($document['kind'] -cne 'instance' -or $resource -isnot [System.Collections.IDictionary] -or
        $resource['uuid'] -cne $Id -or $related -isnot [System.Collections.IDictionary]) {
        throw (New-CloudError -Operation DescribeVM -Failure InvalidInventory -Target $Id)
    }
    $warnings = @($document['warnings'] | Where-Object { $null -ne $_ })
    if (@($warnings | Where-Object { $_ -isnot [string] }).Count) {
        throw (New-CloudError -Operation DescribeVM -Failure InvalidWarnings -Target $Id)
    }
    $networks = $related['l3Networks']
    $resource['memorySizeBytes'] = $resource['memorySize']
    $vm = Convert-CloudVM $resource $Connection
    $vm | Add-Member -NotePropertyMembers @{
        Description = $resource['description']
        HostId = $resource['hostUuid']
        ImageId = $resource['imageUuid']
        InstanceOfferingId = $resource['instanceOfferingUuid']
        HostName = if ($related['host']) { $related['host']['name'] }
        ImageName = if ($related['image']) { $related['image']['name'] }
        InstanceOfferingName = if ($related['instanceOffering']) { $related['instanceOffering']['name'] }
        L3NetworkIds = @(if ($networks -is [System.Collections.IDictionary]) { $networks.Keys })
        Warnings = [string[]] @($warnings | ForEach-Object { Convert-CloudDetailWarning $_ })
    }
    return $vm
}
