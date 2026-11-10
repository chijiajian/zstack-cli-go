Set-StrictMode -Version Latest

$legacyNames = @('Connect-VIServer', 'Disconnect-VIServer', 'Get-VM', 'Start-VM')
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

function Start-VM {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param([Parameter(Mandatory, Position = 0, ValueFromPipeline)] [ValidateNotNullOrEmpty()] [object[]] $VM, [object] $Server)
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
