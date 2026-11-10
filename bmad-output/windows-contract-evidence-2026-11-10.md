# PowerCLI 0.1.2 Windows contract evidence

The [Windows matrix and its downloadable JSON reports](https://github.com/chijiajian/zstack-cli-go/actions/runs/42771892262) ran on 2026-11-10 against source revision `eb27d5977aa5a9e727b776c123e548a56a537277`. It used Windows x64 build 26100, Pester 5.7.1 and a loopback API fixture. Both jobs tested the packaged executable.

The CLI reported version `dev` and commit `eb27d5977aa5a9e727b776c123e548a56a537277`. Both runners tested CLI SHA-256 `c33994c7c5d8613030e949c68b9a9c8e32e1d8c22eb1b2d219fd00f3293d8d0d`.

| PowerShell | Assertions passed | ZStack.PowerCLI 0.1.2 ZIP SHA-256 | Acme.Cloud 0.1.2 ZIP SHA-256 |
| --- | ---: | --- | --- |
| 7.4.20 | 54 | `895332b4d06dbbc1f0737846f2a3411f3481dd83bd39fa3977d04be68ea19b2c` | `0de56c616a8bb5682ca7024a18f0174e9edaa9854eb224692dc46f985b30e6b0` |
| 7.6.6 | 54 | `918a851f68fafa1ac39caba8e41f8198cfafd373a2f96a2db40853c390d51953` | `8b060c4de6edb01d6f3e88f562344a84b0fed5409bbfb5915722242afbedef84` |

The jobs verified CLI identity rejection before package output, brand-specific help, Windows ACLs, `WhatIf`, start timeout and wait, cancellation, terminating and nonterminating errors, compatibility collision handling, offline install, upgrade and rollback. The upgrade test used a synthetic 0.1.3 package alongside 0.1.2.

Each run queried 10,000 mock VMs over 20 measured pairs and recorded 4,000 inventory requests. The added p95 compares the module projection with the CLI’s default JSON output. Peak commit covers the benchmark PowerShell process and its CLI children.

| PowerShell | Added p95 | Combined peak commit | Targets |
| --- | ---: | ---: | --- |
| 7.4.20 | -286.34 ms | 239.18 MiB | Passed 500 ms and 512 MiB |
| 7.6.6 | 302.91 ms | 255.36 MiB | Passed 500 ms and 512 MiB |

Both migration reports recorded 9 of 9 in-scope synthetic samples passing. Across all 12 pairs, five require only command changes, five require rewrites and two require manual review. One rewrite includes unsupported VMware task behavior. There were 12 added lines, 12 removed lines and 13 command replacements.

The [release checklist](../powershell/release/acceptance-checklist.md) records the separate historical cloud runs, cleanup results and production publication decision. Customer scripts and publication identities remain pending.
