function New-CloudNetworkAdapter {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [Parameter(Mandatory)] [ValidateNotNullOrEmpty()] [object] $L3Network,
        [ValidatePattern('\A([a-fA-F0-9]{2}:){5}[a-fA-F0-9]{2}\z')] [string] $MacAddress,
        [ValidateNotNullOrEmpty()] [string] $StaticIP,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $VM) {
            try {
                $target = Get-CloudDeviceSelection $item VM $Server NewNetworkAdapter
                $network = Get-CloudDeviceSelection $L3Network L3Network $target.Connection.Public NewNetworkAdapter
                $arguments = @('instance', 'attach-l3', $target.Id, '--l3-network', $network.Id)
                if ($PSBoundParameters.ContainsKey('MacAddress')) { $arguments += @('--mac', $MacAddress) }
                if ($PSBoundParameters.ContainsKey('StaticIP')) { $arguments += @('--static-ip', $StaticIP) }
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments attach-nic NewNetworkAdapter "Add adapter on L3 network $($network.Id)" $TimeoutSeconds NetworkAdapter -NewIdentity
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
