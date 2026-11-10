# PowerShell preview acceptance

| Evidence | Current state |
| --- | --- |
| Go behavior and race checks | Local fake API and repository `make verify` |
| PowerShell contract | Current 0.1.2 source: 44 assertions passed on macOS arm64 with PowerShell 7.4.13 and again with 7.6.6, using Pester 5.7.1, a loopback fake API and a native macOS development CLI. The upgrade fixture installs versions 0.1.1 and 0.1.2 side by side. Coverage includes both authentication methods and the remaining timeout forwarded to an asynchronous start request. |
| Windows packaging | Windows amd64 CLI cross-built. Both branded module ZIPs and the reproducible offline QA ZIP passed checksum verification. |
| Offline upgrade and rollback | Passed on both native Windows runtimes. The suite installs test versions side by side, checks interruption and incomplete packages, queries with an existing connection after upgrade, and imports the prior version in a fresh session. |
| Windows native import, ACL and installer | Preview 0.1.1 on Windows 11 Enterprise LTSC x64 build 26100.1742, 8 vCPU and 16 GiB on a temporary cloud VM. PowerShell 7.4.20 and 7.6.6 each passed 42 assertions with Pester 5.7.1, the bundled Windows CLI and a loopback API fixture. Coverage includes the ACL, installation and asynchronous start timeout. |
| Query benchmark | Each native Windows 0.1.1 run queried 10,000 VMs over 20 paired samples and made 4,000 inventory requests across 40 measured queries. On 7.4.20, the added p95 was −215.82 ms and combined peak commit was 179.68 MiB. On 7.6.6, the values were 127.72 ms and 178.50 MiB. Both runs met the 500 ms and 512 MiB targets. |
| Real ZStack API | Preview 0.1.2 passed account/password against the authorized test environment and AK/SK against the cloud test environment, using Windows 11 Enterprise LTSC x64 and PowerShell 7.6.6. Both runs connected, queried their dedicated stopped VM, confirmed that `WhatIf` left it stopped, started it, waited for Running and disconnected. An independent API read confirmed both target UUIDs were Running. |
| Customer scripts | Pending sanitized representative scripts |
| Public or OEM publication | Pending real brand/repository/signing identities and explicit publication approval |

Use the [evidence template](evidence-template.md) to record future acceptance runs. Production publication remains gated on the selected customer acceptance cases and publication identities.

The timing delta uses the CLI's default JSON output as the baseline. The module requests a smaller field projection for the same inventory. Windows measured the benchmark process tree's peak commit using a Job Object. PowerShell 7.6.6 passed with the original kit. PowerShell 7.4.20 passed after the fixture startup wait was extended and began distinguishing a process exit from a readiness timeout. The installed module and CLI were identical in both runs.

The Windows live run on 2026-11-10 used the same 0.1.2 module and CLI binary for both authentication methods. The tested Windows CLI SHA-256 was `a6ed96d088875d632369f0b6695b70f13466508777fe3a3386042e45fea011fe`. The AK/SK run required correction of an eight-hour UTC clock offset on the temporary Windows VM. The cloud also advertised the connected management node's IP in its async job URL. The CLI now verifies that IP and port against the connected peer and polls through the configured hostname.

The two target VMs, Windows runner, three root volumes and imported image copy were verified removed. An earlier test-account creation request returned no success confirmation, and account lookup was denied, so its outcome remains unknown. The account/password run used the separate authorized test environment. See [BMAD story 6.4](../../bmad-output/stories/6.4.release-evidence.story.md) for the acceptance record.
