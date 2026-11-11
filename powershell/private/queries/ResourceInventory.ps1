# Keep the CLI field projection and the public property names together. The
# public types carry a live connection identity independent of the OEM brand.
$script:InventoryQueries = @{
    Volume = @{ Command = 'disks'; Properties = [ordered]@{
        Name = 'name'; VMId = 'vmInstanceUuid'; PrimaryStorageId = 'primaryStorageUuid'; Type = 'type'
        Format = 'format'; State = 'state'; Status = 'status'; CapacityBytes = 'sizeBytes'
        ActualSizeBytes = 'actualSizeBytes'
    }; GiB = @{ CapacityGB = 'CapacityBytes' } }
    VolumeSnapshot = @{ Command = 'snapshots'; Properties = [ordered]@{
        Name = 'name'; VolumeId = 'volumeUuid'; ParentId = 'parentUuid'; TreeId = 'treeUuid'
        Status = 'status'; State = 'state'; Latest = 'latest'; CreateDate = 'createDate'
    } }
    CDDrive = @{ Command = 'cdroms'; Properties = [ordered]@{
        Name = 'name'; VMId = 'vmInstanceUuid'; IsoImageId = 'isoUuid'; DeviceId = 'deviceId'
    } }
    VMHost = @{ Command = 'hosts'; Properties = [ordered]@{
        Name = 'name'; ClusterId = 'clusterUuid'; ZoneId = 'zoneUuid'; ManagementIP = 'managementIp'
        HypervisorType = 'hypervisorType'; State = 'state'; ConnectionState = 'status'
        Architecture = 'architecture'; NumCpu = 'cpuCount'; CpuSocketCount = 'cpuSocketCount'
        CpuTotalHz = 'totalCpuHz'; CpuAvailableHz = 'availableCpuHz'; MemoryTotalBytes = 'totalMemoryBytes'
        MemoryAvailableBytes = 'availableMemoryBytes'
    }; GiB = @{ MemoryTotalGB = 'MemoryTotalBytes'; MemoryAvailableGB = 'MemoryAvailableBytes' } }
    Cluster = @{ Command = 'clusters'; Properties = [ordered]@{
        Name = 'name'; ZoneId = 'zoneUuid'; State = 'state'; Type = 'type'; HypervisorType = 'hypervisorType'
        Architecture = 'architecture'
    } }
    Zone = @{ Command = 'zones'; Properties = [ordered]@{ Name = 'name'; State = 'state'; Type = 'type'; Description = 'description' } }
    NetworkAdapter = @{ Command = 'nics'; Properties = [ordered]@{
        VMId = 'vmInstanceUuid'; L3NetworkId = 'l3NetworkUuid'; MacAddress = 'mac'; IPAddress = 'ip'
        Netmask = 'netmask'; Gateway = 'gateway'; IPVersion = 'ipVersion'; DeviceId = 'deviceId'; Type = 'type'
        DriverType = 'driverType'
    } }
    PrimaryStorage = @{ Command = 'primary-storages'; Properties = [ordered]@{
        Name = 'name'; ZoneId = 'zoneUuid'; Type = 'type'; State = 'state'; Status = 'status'
        CapacityBytes = 'totalCapacityBytes'; FreeSpaceBytes = 'availableCapacityBytes'
        PhysicalCapacityBytes = 'totalPhysicalCapacityBytes'
        PhysicalFreeSpaceBytes = 'availablePhysicalCapacityBytes'
        SystemUsedCapacityBytes = 'systemUsedCapacityBytes'
    }; GiB = @{ CapacityGB = 'CapacityBytes'; FreeSpaceGB = 'FreeSpaceBytes' } }
    ImageStorage = @{ Command = 'image-storages'; Properties = [ordered]@{
        Name = 'name'; Type = 'type'; State = 'state'; Status = 'status'; CapacityBytes = 'totalCapacityBytes'
        FreeSpaceBytes = 'availableCapacityBytes'
    }; GiB = @{ CapacityGB = 'CapacityBytes'; FreeSpaceGB = 'FreeSpaceBytes' } }
    Image = @{ Command = 'images'; Properties = [ordered]@{
        Name = 'name'; Type = 'type'; MediaType = 'mediaType'; Format = 'format'; Platform = 'platform'
        Architecture = 'architecture'; GuestOsType = 'guestOsType'; State = 'state'; Status = 'status'
        CapacityBytes = 'sizeBytes'; ActualSizeBytes = 'actualSizeBytes'
    }; GiB = @{ CapacityGB = 'CapacityBytes' } }
    InstanceOffering = @{ Command = 'instance-offerings'; Properties = [ordered]@{
        Name = 'name'; NumCpu = 'cpuCount'; MemoryBytes = 'memorySizeBytes'; Type = 'type'; State = 'state'
        AllocatorStrategy = 'allocatorStrategy'
    }; GiB = @{ MemoryGB = 'MemoryBytes' } }
    DiskOffering = @{ Command = 'disk-offerings'; Properties = [ordered]@{
        Name = 'name'; CapacityBytes = 'diskSizeBytes'; Type = 'type'; State = 'state'
        AllocatorStrategy = 'allocatorStrategy'
    }; GiB = @{ CapacityGB = 'CapacityBytes' } }
    L2Network = @{ Command = 'l2-networks'; Properties = [ordered]@{
        Name = 'name'; ZoneId = 'zoneUuid'; Type = 'type'; VirtualNetworkId = 'virtualNetworkId'
        PhysicalInterface = 'physicalInterface'
    } }
    L3Network = @{ Command = 'l3-networks'; Properties = [ordered]@{
        Name = 'name'; L2NetworkId = 'l2NetworkUuid'; ZoneId = 'zoneUuid'; Type = 'type'; State = 'state'
        IPVersion = 'ipVersion'; DnsDomain = 'dnsDomain'; System = 'system'; Category = 'category'
        EnableIPAM = 'enableIPAM'
    } }
    IPRange = @{ Command = 'ip-ranges'; Properties = [ordered]@{
        Name = 'name'; L3NetworkId = 'l3NetworkUuid'; StartIP = 'startIp'; EndIP = 'endIp'; Netmask = 'netmask'
        Gateway = 'gateway'; CIDR = 'networkCidr'; IPVersion = 'ipVersion'; AddressMode = 'addressMode'
        PrefixLength = 'prefixLen'
    } }
}

function Read-CloudResourceRows {
    param($Connection, [string] $Type, [string[]] $Name, [string[]] $Id, [string] $ParentField, [string[]] $ParentIds,
        [Diagnostics.Stopwatch] $Clock, [int] $Budget, [string] $Operation)

    $spec = $script:InventoryQueries[$Type]
    $fields = @('uuid') + @($spec.Properties.Values)
    $patterns = @(foreach ($pattern in $Name) { [System.Management.Automation.WildcardPattern]::new($pattern, 'IgnoreCase') })
    $identityFilter = [System.Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    foreach ($identifier in $Id) { $null = $identityFilter.Add($identifier) }
    $queryField = if ($ParentField) { $ParentField } else { 'uuid' }
    $queryIds = if ($ParentField) { $ParentIds } else { @($identityFilter) }
    $queryIds = @($queryIds | Where-Object { $null -ne $_ })
    if (@($queryIds | Where-Object { $_ -cnotmatch '\A[A-Za-z0-9-]{1,128}\z' }).Count) {
        throw (New-CloudError -Operation "Query$Type" -Failure InvalidInventory -Target $Connection.Id)
    }
    $base = @('get', $spec.Command, '--all-pages', '--sort', '+uuid', '--fields', ($fields -join ','), '-o', 'json')
    $seen = [System.Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    # Limit query-string length without starting a CLI process for each object.
    $batchSize = 32
    for ($offset = 0; $offset -lt [Math]::Max(1, $queryIds.Count); $offset += $batchSize) {
        $batch = @()
        $arguments = $base
        if ($queryIds.Count) {
            $last = [Math]::Min($offset + $batchSize, $queryIds.Count) - 1
            $batch = @($queryIds[$offset..$last])
            if ($batch.Count -eq 1) { $arguments += @('-q', "$queryField=$($batch[0])") }
            else { $arguments += @('-q', "${queryField}?=$($batch -join ',')") }
        }
        $remaining = if ($Clock) { [Math]::Min(25, (Get-CloudRemainingTime $Clock $Budget $Operation $Id[0])) } else { 25 }
        $milliseconds = [long] [Math]::Floor($remaining * 1000)
        $result = Invoke-CloudConnectedCli $Connection ($arguments + @('--request-timeout', "${milliseconds}ms")) $remaining
        $rows = Convert-CliResult $result "Query$Type" -Shape Array -Target $Connection.Id
        foreach ($row in $rows) {
            if ($row -isnot [System.Collections.IDictionary] -or $row['uuid'] -isnot [string] -or
                $row['uuid'] -cnotmatch '\A[A-Za-z0-9-]{1,128}\z' -or
                ($spec.Properties.Contains('Name') -and $row['name'] -isnot [string]) -or
                ($batch.Count -and $batch -cnotcontains $row[$queryField])) {
                throw (New-CloudError -Operation "Query$Type" -Failure InvalidInventory -Target $Connection.Id)
            }
            if ($identityFilter.Count -and -not $identityFilter.Contains($row['uuid'])) { continue }
            if ($patterns.Count) {
                $matched = $false
                foreach ($pattern in $patterns) {
                    if ($pattern.IsMatch($row['name'])) { $matched = $true; break }
                }
                if (-not $matched) { continue }
            }
            if ($seen.Add($row['uuid'])) { $row }
        }
    }
}

function Convert-CloudResource {
    param($Row, $Connection, [string] $Type, [string] $OutputType, $NetworkNames)

    $spec = $script:InventoryQueries[$Type]
    $fields = [ordered]@{
        PSTypeName = "Cloud.Automation.$OutputType"; Id = $Row['uuid']; Name = $null
    }
    foreach ($property in $spec.Properties.GetEnumerator()) {
        $key = $property.Value
        $value = $Row[$key]
        # Local storage can report signed system usage. Preserve the API value.
        if ($key -cmatch '(Bytes|Hz|Count|NetworkId)$' -and
            (-not $Row.Contains($key) -or ($null -ne $value -and
                (($value -isnot [int] -and $value -isnot [long]) -or ($value -lt 0 -and $key -cne 'systemUsedCapacityBytes'))))) {
            throw (New-CloudError -Operation "Query$Type" -Failure InvalidNumericField -Target $Connection.Id)
        }
        $fields[$property.Key] = $value
    }
    if ($spec.ContainsKey('GiB')) {
        foreach ($property in $spec.GiB.GetEnumerator()) {
            $bytes = $fields[$property.Value]
            $fields[$property.Key] = if ($null -ne $bytes) { [double] $bytes / 1GB }
        }
    }
    if ($Type -ceq 'LongJob') { $fields['JobId'] = $fields.Id }
    if ($Type -ceq 'Volume') { $fields['Attached'] = -not [string]::IsNullOrEmpty($fields['VMId']) }
    if ($Type -ceq 'NetworkAdapter') {
        $networkName = $null
        if ($fields['L3NetworkId'] -and $null -ne $NetworkNames) { $null = $NetworkNames.TryGetValue($fields['L3NetworkId'], [ref] $networkName) }
        $fields['NetworkName'] = $networkName
    }
    $fields['ConnectionId'] = $Connection.Id
    $fields['BrandId'] = $Connection.BrandId
    $fields['Generation'] = $Connection.Generation
    $resource = [pscustomobject] $fields
    $identity = @{ Id = $resource.Id; Type = $OutputType }
    if ($fields.Contains('VMId')) { $identity.VMId = $fields['VMId'] }
    if ($fields.Contains('VolumeId')) { $identity.VolumeId = $fields['VolumeId'] }
    if ($fields.Contains('L3NetworkId')) { $identity.L3NetworkId = $fields['L3NetworkId'] }
    $Connection.Objects.Add($resource.PSObject, $identity)
    return $resource
}

function Get-CloudInventory {
    param(
        [string] $Type, [string[]] $Name,
        [string[]] $Id,
        [object] $Server, [string] $ParentField, [string[]] $ParentType, [object[]] $Parent,
        [string] $OutputType = $Type, [switch] $IncludeNetworkName
    )

    $groups = [ordered]@{}
    if ($ParentField) {
        # Validate the complete pipeline before any reads, including explicit -Server.
        foreach ($item in $Parent) {
            $matchingType = @($ParentType | Where-Object { $item.PSObject.TypeNames -ccontains "Cloud.Automation.$_" })
            if ($matchingType.Count -ne 1) { throw (New-CloudError -Operation "Query$Type" -ExitCode 5) }
            $connection = Get-CloudObjectConnection $item $matchingType[0] $Server "Query$Type"
            if (-not $groups.Contains($connection.Id)) {
        $groups[$connection.Id] = @{ Connection = $connection
        Ids = [System.Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal) }
            }
            $null = $groups[$connection.Id].Ids.Add($item.Id)
        }
    } else {
        $connection = Get-CloudConnection $Server
        $groups[$connection.Id] = @{ Connection = $connection; Ids = @() }
    }
    if ($groups.Count -eq 0) { return }
    $capability = $script:InventoryQueries[$Type]['Capability'] ?? 'resourceInventoryV1'
    $null = Get-CloudCliPath $capability
    foreach ($group in $groups.Values) {
        $networkNames = $null
        if ($IncludeNetworkName) {
            $rows = @(Read-CloudResourceRows $group.Connection $Type $Name $Id $ParentField @($group.Ids))
            $networkIds = @($rows | ForEach-Object { $_['l3NetworkUuid'] } | Where-Object { $_ } | Select-Object -Unique)
            $networkNames = [System.Collections.Generic.Dictionary[string,string]]::new([StringComparer]::Ordinal)
            if ($networkIds.Count) {
                foreach ($network in (Read-CloudResourceRows $group.Connection L3Network -Id $networkIds)) {
                    $networkNames.Add($network['uuid'], $network['name'])
                }
            }
            foreach ($row in $rows) { Convert-CloudResource $row $group.Connection $Type $OutputType $networkNames }
        } else {
            Read-CloudResourceRows $group.Connection $Type $Name $Id $ParentField @($group.Ids) | ForEach-Object {
                Convert-CloudResource $_ $group.Connection $Type $OutputType $null
            }
        }
    }
}

function Get-CloudResourceSelection {
    param($Reference, [string[]] $Types, $Server, [string] $Operation)
    if ($Reference -is [string]) {
        if ($Reference -cnotmatch '\A[A-Za-z0-9-]{1,128}\z') { throw (New-CloudError -Operation $Operation -ExitCode 2) }
        return @{ Connection = Get-CloudConnection $Server; Id = $Reference; Name = $Reference }
    }
    $type = @($Types | Where-Object { $Reference.PSObject.TypeNames -ccontains "Cloud.Automation.$_" })
    if ($type.Count -ne 1) { throw (New-CloudError -Operation $Operation -ExitCode 5) }
    $connection = Get-CloudObjectConnection $Reference $type[0] $Server $Operation
    $identity = $null
    $null = $connection.Objects.TryGetValue($Reference.PSObject, [ref] $identity)
    return @{ Connection = $connection; Id = $identity.Id; Name = $Reference.Name; Identity = $identity }
}
