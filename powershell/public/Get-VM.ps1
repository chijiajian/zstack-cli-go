function Get-CloudVM {
    [CmdletBinding(DefaultParameterSetName = 'Name')]
    param(
        [Parameter(Position = 0, ParameterSetName = 'Name')] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [Parameter(ParameterSetName = 'Name')] [switch] $Exact,
        [Parameter(Mandatory, ParameterSetName = 'Id')] [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]+\z')] [string[]] $Id,
        [object] $Server,
        [switch] $Detail
    )

    if ($Detail -and ($PSCmdlet.ParameterSetName -ne 'Id' -or $Id.Count -ne 1)) {
        throw (New-CloudError -Operation DescribeVM -ExitCode 2)
    }
    $connection = Get-CloudConnection $Server
    if ($Detail) {
        Read-CloudVMDetail $connection $Id[0]
        return
    }
    if ($PSCmdlet.ParameterSetName -eq 'Id' -and $Id.Count -eq 1) {
        Convert-CloudVM (Read-CloudVMInventory $connection -Id $Id[0]) $connection
        return
    }
    $patterns = @($Name | Where-Object { $null -ne $_ } | ForEach-Object {
        if ($_ -eq '') { throw 'Name must not be empty.' }
        [System.Management.Automation.WildcardPattern]::new($_, 'IgnoreCase')
    })
    if ($Exact) {
        if ($patterns.Count -eq 0 -or
            @($Name | Where-Object { [System.Management.Automation.WildcardPattern]::ContainsWildcardCharacters($_) }).Count -gt 0) {
            throw 'Specify one or more literal names with -Name and -Exact.'
        }
        $matches = @(foreach ($literalName in $Name) {
            foreach ($row in (Read-CloudVMInventory $connection -ExactName $literalName)) { $row }
        })
        Convert-CloudVM @($matches | Sort-Object { $_['uuid'] } -Unique) $connection
        return
    }
    $rows = Read-CloudVMInventory $connection
    if ($PSCmdlet.ParameterSetName -eq 'Name' -and $patterns.Count -eq 0) {
        Convert-CloudVM $rows $connection
        return
    }
    $matches = foreach ($inventory in $rows) {
        if ($PSCmdlet.ParameterSetName -eq 'Id' -and $Id -cnotcontains $inventory['uuid']) { continue }
        if ($patterns.Count -gt 0) {
            $matched = $false
            foreach ($pattern in $patterns) {
                if ($pattern.IsMatch([string] $inventory['name'])) { $matched = $true; break }
            }
            if (-not $matched) { continue }
        }
        $inventory
    }
    Convert-CloudVM $matches $connection
}
