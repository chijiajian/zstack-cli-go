# Module package installation and rollback

Install each release to its own version directory. Existing connections keep using the CLI from the version that created them. Keep that version installed until those connections are closed. To upgrade, open a fresh PowerShell session and import the newer version. To roll back, open a fresh session and import the desired version by manifest path or with `Import-Module -FullyQualifiedName @{ ModuleName = 'ZStack.PowerCLI'; RequiredVersion = '0.1.0' }`. Establish new connections in that session.

The offline installer checks the archive against a SHA-256 value obtained from the trusted release channel. It verifies the package layout, required module files and bundled CLI hash in a temporary directory. The new version becomes discoverable only after validation and a directory move. If installation is interrupted before that move, PowerShell continues to discover the previously installed versions. The installer preserves existing version directories. Release authenticity depends on the delivery channel and signing policy.

With PowerShell 7.4 or later, build hosts can use `-ModuleDirectory` to stage the package into an explicit path. Running the packaged CLI requires Windows x64.

The package provider runs `Build-Package.ps1` on Windows with the Windows x64 CLI it will ship. The builder reads the executable's version and commit and checks command help for the required capabilities. Use `-CliVersion` and `-CliCommit` to assert an expected identity. These parameters must match the executable. Go builds without release metadata report `dev` and `none`.
