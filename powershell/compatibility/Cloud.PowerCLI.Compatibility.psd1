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
}
