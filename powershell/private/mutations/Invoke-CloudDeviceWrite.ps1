function Get-CloudDeviceSelection {
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

function Read-CloudWrittenDevice {
    param($Connection, [string] $Type, [string] $Id, [string] $OutputType, $Clock, [int] $Budget, [string] $Operation)
    try {
        $rows = @(Read-CloudResourceRows $Connection $Type -Id @($Id) -Clock $Clock -Budget $Budget -Operation $Operation)
        if ($rows.Count -ne 1) { throw 'The written resource was not observed.' }
        Convert-CloudResource $rows[0] $Connection $Type $OutputType $null
    } catch { throw (New-CloudError -Operation $Operation -Failure StateUnverified -Mutation -Target $Id) }
}

function Invoke-CloudDeviceWrite {
    param([System.Management.Automation.PSCmdlet] $Caller, [hashtable] $Target,
        [string[]] $Arguments, [string] $Command, [string] $Operation, [string] $Description,
        [int] $TimeoutSeconds, [string] $ResultType, [string] $OutputType = $ResultType,
        [switch] $NewIdentity)
    $connection = $Target.Connection
    $id = $Target.Id
    $label = '{0} ({1}) at {2}' -f $Target.Name, $id, $connection.Endpoint
    if (-not $Caller.ShouldProcess($label, $Description)) { return }
    $null = Get-CloudCliPath 'deviceLifecycleV1'
    $clock = [Diagnostics.Stopwatch]::StartNew()
    $remaining = Get-CloudRemainingTime $clock $TimeoutSeconds $Operation $id
    $milliseconds = [long] [Math]::Floor($remaining * 1000)
    $options = @('--uuid-only', '--yes', '--request-timeout', "${milliseconds}ms", '-o', 'json')
    $result = Invoke-CloudConnectedCli $connection ($Arguments[0..1] + $options + $Arguments[2..($Arguments.Count - 1)]) $remaining
    $response = Convert-CliResult $result $Operation -Mutation -Target $id
    if ($response['operation'] -cne $Command -or $response['targetUuid'] -cne $id -or
        $response['status'] -cnotin 'succeeded', 'skipped' -or
        $response['uuid'] -isnot [string] -or $response['uuid'] -cnotmatch '\A[A-Za-z0-9-]{1,128}\z' -or
        (-not $NewIdentity -and $response['uuid'] -cne $id) -or
        ($NewIdentity -and $response['status'] -cne 'succeeded')) {
        throw (New-CloudError -Operation $Operation -Failure InvalidWriteResult -Mutation -Target $id)
    }
    if ($ResultType) {
        Read-CloudWrittenDevice $connection $ResultType $response['uuid'] $OutputType $clock $TimeoutSeconds $Operation
    }
}

function Get-CloudDeviceParentId {
    param($Target, [string] $Key, $Reference, [string[]] $Types, [string] $Operation)
    $expected = if ($Target.ContainsKey('Identity')) { $Target.Identity[$Key] }
    if ($null -ne $Reference) {
        $parent = Get-CloudDeviceSelection $Reference $Types $Target.Connection.Public $Operation
        if ($expected -and $expected -cne $parent.Id) { throw (New-CloudError -Operation $Operation -ExitCode 5 -Target $Target.Id) }
        return $parent.Id
    }
    if (-not $expected) { throw (New-CloudError -Operation $Operation -ExitCode 5 -Target $Target.Id) }
    return $expected
}
