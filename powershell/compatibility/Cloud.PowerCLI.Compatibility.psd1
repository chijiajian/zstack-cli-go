@{
    RootModule = '__MODULE__.Compatibility.psm1'
    ModuleVersion = '__VERSION__'
    Author = '__PUBLISHER__'
    Description = 'Opt-in PowerCLI migration subset for __DESCRIPTION__'
    PowerShellVersion = '7.4'
    CompatiblePSEditions = @('Core')
    FunctionsToExport = @(__LEGACYEXPORTS__)
    CmdletsToExport = @()
    VariablesToExport = @()
    AliasesToExport = @()
    PrivateData = @{
        PSData = @{
            LicenseUri = '__LICENSE_URI__'
            Tags = @('Windows', 'x64', 'PSEdition_Core')
__PROJECT_URI_LINE__
        }
    }
}
