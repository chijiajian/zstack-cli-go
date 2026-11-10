function Get-CloudTargetVM {
    param($Connection, [string] $Id, [string] $Name, [double] $TimeoutSeconds = 120)

    $rows = Read-CloudVMInventory $Connection -Id $Id -TimeoutSeconds $TimeoutSeconds
    if (-not $Id) { $rows = @($rows | Where-Object { $_['name'] -eq $Name }) }
    if ($rows.Count -eq 0) { throw (New-CloudError -Operation ResolveVM -ExitCode 4 -Target ($Id ?? $Name)) }
    if ($rows.Count -ne 1) { throw (New-CloudError -Operation ResolveVM -ExitCode 5 -Target ($Id ?? $Name)) }
    return Convert-CloudVM $rows[0] $Connection
}

function Resolve-CloudVMTarget {
    param([object] $VM, [string] $Name, [string] $Id, [object] $Server, [double] $TimeoutSeconds = 120)

    if ($null -eq $VM) {
        $connection = Get-CloudConnection $Server
    } else {
        $key = $VM.PSObject.Properties['ConnectionId']
        if ($null -eq $key -or $key.Value -isnot [string] -or -not $script:Connections.ContainsKey($key.Value)) {
            throw (New-CloudError -Operation ResolveVM -ExitCode 5)
        }
        $connection = $script:Connections[$key.Value]
        $originalId = $null
        if (-not $connection.Objects.TryGetValue($VM.PSObject, [ref] $originalId) -or
            $VM.Id -cne $originalId -or $VM.BrandId -cne $connection.BrandId -or $VM.Generation -cne $connection.Generation) {
            throw (New-CloudError -Operation ResolveVM -ExitCode 5)
        }
        if ($null -ne $Server -and (Get-CloudConnection $Server).Id -cne $connection.Id) {
            throw (New-CloudError -Operation ResolveVM -ExitCode 5 -Target $originalId)
        }
        $Id = $originalId
    }
    $target = Get-CloudTargetVM $connection -Id $Id -Name $Name -TimeoutSeconds $TimeoutSeconds
    return @{ Connection = $connection; VM = $target }
}
