function Start-CloudVM {
    [CmdletBinding(DefaultParameterSetName = 'VM', SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline, ParameterSetName = 'VM')] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [Parameter(Mandatory, ParameterSetName = 'Name')] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [Parameter(Mandatory, ParameterSetName = 'Id')] [ValidateNotNullOrEmpty()] [ValidatePattern('\A[A-Za-z0-9-]+\z')] [string[]] $Id,
        [object] $Server
    )

    process {
        $null = Get-CloudCliPath 'startResultV1'
        $parameter = $PSCmdlet.ParameterSetName
        $selectors = $PSBoundParameters[$parameter]
        foreach ($selector in $selectors) {
            $clock = [System.Diagnostics.Stopwatch]::StartNew()
            $target = $null
            try {
                $selection = @{ Server = $Server; TimeoutSeconds = 300 }
                $selection[$parameter] = $selector
                $target = Resolve-CloudVMTarget @selection
            } catch { $PSCmdlet.WriteError($_); continue }
            $vmId = $target.VM.Id
            if (-not (Invoke-CloudShouldProcess $PSCmdlet $target 'Start VM')) { continue }
            if ($target.VM.RawState -ceq 'Running') { $target.VM; continue }
            if ($target.VM.RawState -cne 'Stopped') {
                $PSCmdlet.WriteError((New-CloudError -Operation StartVM -ExitCode 5 -Target $vmId))
                continue
            }
            try {
                $remaining = Get-CloudRemainingTime $clock 300 'StartVM' $vmId
                $milliseconds = [long] [Math]::Floor($remaining * 1000)
                $arguments = @('instance', 'start', $vmId, '--uuid-only', '--yes', '--request-timeout', "${milliseconds}ms", '-o', 'json')
                $result = Invoke-CloudConnectedCli $target.Connection $arguments $remaining
                $response = Convert-CliResult $result 'StartVM' -Mutation -Target $vmId
                if ($response['operation'] -cne 'start' -or $response['items'] -isnot [array] -or $response['items'].Count -ne 1 -or
                    $response['items'][0]['uuid'] -cne $vmId -or $response['items'][0]['status'] -cnotin 'succeeded', 'skipped') {
                    throw (New-CloudError -Operation StartVM -Failure InvalidStartResult -Mutation -Target $vmId)
                }
            } catch { $PSCmdlet.WriteError($_); continue }
            try {
                Wait-CloudVMState $target $clock 300 2
            } catch { $PSCmdlet.WriteError((New-CloudError -Operation StartVM -Failure StateUnverified -Mutation -Target $vmId)) }
        }
    }
}
