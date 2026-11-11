Set-StrictMode -Version Latest

$legacyNames = @(__LEGACYEXPORTS__)
# An exact-name lookup can autoload a module even with -ListImported.
foreach ($command in (Get-Command -All -ListImported)) {
    if ($command.Name -in $legacyNames) {
        throw "CloudAutomation.CompatibilityConflict: '$($command.Name)' is already loaded. Use a separate PowerShell session."
    }
}
$nativeManifest = Join-Path (Split-Path -Parent $PSScriptRoot) '__MODULE__.psd1'
Import-Module -Name $nativeManifest -Scope Local -ErrorAction Stop

function Connect-VIServer {
    [CmdletBinding()]
    param([Parameter(Mandatory, Position = 0)] [string] $Server, [Parameter(Mandatory)] [pscredential] $Credential)
    & '__MODULE__\Connect-__PREFIX__Server' @PSBoundParameters
}

function Disconnect-VIServer {
    [CmdletBinding()]
    param([Parameter(Mandatory, Position = 0, ValueFromPipeline)] [object[]] $Server)
    process { & '__MODULE__\Disconnect-__PREFIX__Server' @PSBoundParameters }
}

function Get-VM {
    [CmdletBinding(DefaultParameterSetName = 'Name')]
    param(
        [Parameter(Position = 0, ParameterSetName = 'Name')] [ValidateNotNullOrEmpty()] [string[]] $Name,
        [Parameter(Mandatory, ParameterSetName = 'Id')] [string[]] $Id,
        [object] $Server
    )
    & '__MODULE__\Get-__PREFIX__VM' @PSBoundParameters
}

function Get-HardDisk {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [string[]] $Name,
        [string[]] $Id,
        [object] $Server
    )
    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { $targets.AddRange($VM) }
    end {
        if ($targets.Count) {
            $PSBoundParameters['VM'] = $targets.ToArray()
            & '__MODULE__\Get-__PREFIX__HardDisk' @PSBoundParameters
        }
    }
}

function Get-NetworkAdapter {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [string[]] $Id,
        [object] $Server
    )
    begin { $targets = [System.Collections.Generic.List[object]]::new() }
    process { $targets.AddRange($VM) }
    end {
        if ($targets.Count) {
            $PSBoundParameters['VM'] = $targets.ToArray()
            & '__MODULE__\Get-__PREFIX__NetworkAdapter' -IncludeNetworkName @PSBoundParameters
        }
    }
}

function Get-VMHost {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [string[]] $Name,
        [string[]] $Id,
        [object[]] $Location,
        [object] $Server
    )
    if ($PSBoundParameters.ContainsKey('Location')) {
        $PSBoundParameters['Cluster'] = $Location
        $null = $PSBoundParameters.Remove('Location')
    }
    & '__MODULE__\Get-__PREFIX__VMHost' @PSBoundParameters
}

function Get-Cluster {
    [CmdletBinding()]
    param(
        [Parameter(Position = 0)] [string[]] $Name,
        [string[]] $Id,
        [object] $Server
    )
    & '__MODULE__\Get-__PREFIX__Cluster' @PSBoundParameters
}

function Start-VM {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM,
        [object] $Server,
        [ValidateRange(1, 3600)] [int] $TimeoutSeconds = 300
    )
    process {
        $forward = @{}
        foreach ($key in $PSBoundParameters.Keys) { if ($key -ne 'VM') { $forward[$key] = $PSBoundParameters[$key] } }
        foreach ($item in $VM) {
            if ($item -is [string]) { & '__MODULE__\Start-__PREFIX__VM' -Name $item @forward }
            else { & '__MODULE__\Start-__PREFIX__VM' -VM $item @forward }
        }
    }
}

Export-ModuleMember -Function $legacyNames
