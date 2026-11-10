# PowerShell preview acceptance

| Evidence | Current state |
| --- | --- |
| Go behavior and race checks | Local fake API and repository `make verify` |
| PowerShell contract | Earlier 0.1.2 development build: 44 assertions passed on macOS arm64 with PowerShell 7.4.13 and again with 7.6.6, using Pester 5.7.1, a loopback fake API and a native macOS development CLI. The upgrade fixture installs versions 0.1.1 and 0.1.2 side by side. Coverage includes both authentication methods and the remaining timeout forwarded to an asynchronous start request. |
| Windows packaging | Windows amd64 CLI cross-built. Both branded module ZIPs and the reproducible offline QA ZIP passed checksum verification. |
| Offline upgrade and rollback | Passed on both native Windows runtimes. The suite installs test versions side by side, checks interruption and incomplete packages, queries with an existing connection after upgrade, and imports the prior version in a fresh session. |
| Windows native import, ACL and installer | Preview 0.1.1 on Windows 11 Enterprise LTSC x64 build 26100.1742, 8 vCPU and 16 GiB on a temporary cloud VM. PowerShell 7.4.20 and 7.6.6 each passed 42 assertions with Pester 5.7.1, the bundled Windows CLI and a loopback API fixture. Coverage includes the ACL, installation and asynchronous start timeout. |
| Query benchmark | Each native Windows 0.1.1 run queried 10,000 VMs over 20 paired samples and made 4,000 inventory requests across 40 measured queries. On 7.4.20, the added p95 was −215.82 ms and combined peak commit was 179.68 MiB. On 7.6.6, the values were 127.72 ms and 178.50 MiB. Both runs met the 500 ms and 512 MiB targets. |
| Real ZStack API | Preview 0.1.2 passed account/password against the authorized test environment and AK/SK against the cloud test environment, using Windows 11 Enterprise LTSC x64 and PowerShell 7.6.6. Both runs connected, queried their dedicated stopped VM, confirmed that `WhatIf` left it stopped, started it, waited for Running and disconnected. An independent API read confirmed both target UUIDs were Running. |
| Customer scripts | Pending sanitized representative scripts |
| Public or OEM publication | Pending real brand/repository/signing identities and explicit publication approval |

Use the [evidence template](evidence-template.md) to record future acceptance runs. **Production and OEM publication decision: no-go.** The preview is available for controlled validation. Representative customer scripts, production brand/repository/signing identities and publication approval remain pending.

The timing delta uses the CLI's default JSON output as the baseline. The module requests a smaller field projection for the same inventory. Windows measured the benchmark process tree's peak commit using a Job Object. PowerShell 7.6.6 passed with the original kit. PowerShell 7.4.20 passed after the fixture startup wait was extended and began distinguishing a process exit from a readiness timeout. The installed module and CLI were identical in both runs. The CLI identified itself as `dev` / `none`. Go recorded base revision `4855aca5bc47c99ae362070456773840240a5` with workspace modifications. Its SHA-256 was `4346f93766ff7e2e59d4b7ca5e0862c741ef6e90e31ecf2f3cdf50b8fe63eb`. The offline QA kits had SHA-256 `0cf8bd465bd2596ba15258e98ac61d4c821b25ad50e6f9cc9df52e9fafe19f2f` (original) and `72970a78e681dd993b6f2c4669cdd722a5ceaa543e73ef3c790c54a3ca146586` (updated fixture).

The Windows live run on 2026-11-10 used the same 0.1.2 module and CLI binary for both authentication methods. The ZStack-branded `ZStack.PowerCLI-0.1.2-windows-amd64.zip` SHA-256 was `5089eca5da93d8ec193f63be2c63d4a82dcef500a112a6cefa3bf0d30b6bb4e1`. The CLI identified itself as `dev` / `none`. The executable embeds Go base revision `4855aca5bc47c99ae362070456773840240a5` with `vcs.modified=true`, so this historical run identifies the tested binary by hash. The tested Windows CLI SHA-256 was `a6ed96d088875d632369f0b6695b70f13466508777fe3a3386042e45fea011fe`. The AK/SK run required correction of an eight-hour UTC clock offset on the temporary Windows VM. The cloud also advertised the connected management node's IP in its async job URL. The CLI now verifies that IP and port against the connected peer and polls through the configured hostname.

The live test was authorized for the two temporary target VMs. The control-side checks used the exact resource UUIDs. The public record uses the following aliases and the execution journal retains the UUIDs.

| Test resource | Result | Cleanup confirmed by follow-up query |
| --- | --- | --- |
| Account/password VM and its root volume | VM moved from Stopped to Running after the authorized start | Both absent |
| AK/SK VM and its root volume | VM moved from Stopped to Running after the authorized start | Both absent |
| Windows runner, its root volume and imported image copy | Used for both acceptance runs | All absent |
| Candidate test account | Create request returned no success confirmation. Lookup was denied. The password test used a separate authorized environment. | Unknown. An IAM-authorized operator still needs to confirm existence before taking action. |

See [BMAD story 6.4](../../bmad-output/stories/6.4.release-evidence.story.md) for the acceptance record.
