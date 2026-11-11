function Remove-CloudNetworkAdapter {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $NetworkAdapter,
        [object] $VM,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300,
        [object] $Server
    )
    process {
        foreach ($item in $NetworkAdapter) {
            try {
                $target = Get-CloudDeviceSelection $item NetworkAdapter $Server RemoveNetworkAdapter
                $vmId = Get-CloudDeviceParentId $target VMId $VM VM RemoveNetworkAdapter
                $arguments = @('instance', 'detach-l3', $vmId, '--nic', $target.Id)
                if ($target.ContainsKey('Identity')) { $arguments += @('--l3-network', $target.Identity.L3NetworkId) }
                Invoke-CloudDeviceWrite $PSCmdlet $target $arguments detach-nic RemoveNetworkAdapter "Remove network adapter from VM $vmId" $TimeoutSeconds
            } catch { $PSCmdlet.WriteError($_) }
        }
    }
}
