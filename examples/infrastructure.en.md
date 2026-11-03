# Networking and infrastructure

[English](infrastructure.en.md) | [简体中文](infrastructure.md) | [Examples](README.en.md)

Discover resources and command parameters before choosing zone, cluster or network references. These queries may renew a saved account session.

```sh
zstack-cli get zones --context lab -o json
zstack-cli get clusters --context lab -o json
zstack-cli get primary-storages --context lab -o json
zstack-cli get l2-networks --context lab -o json
zstack-cli get l3-networks --context lab -o json
zstack-cli get security-groups --context lab -o json
zstack-cli get load-balancers --context lab -o json
```

Use help to discover write operations. Backend requirements vary, so inspect parameters and current server capabilities before execution.

```sh
zstack-cli host add-kvm --help
zstack-cli primary-storage add-local --help
zstack-cli primary-storage add-nfs --help
zstack-cli backup-storage add-image-store --help
zstack-cli l2-network attach-cluster --help
zstack-cli l3-network add-ip-range --help
zstack-cli security-group add-rule --help
zstack-cli create eip --help
zstack-cli port-forwarding attach --help
zstack-cli create load-balancer-listener --help
zstack-cli load-balancer add-nic --help
zstack-cli global-config set --help
```

Infrastructure, network rules and global configuration can affect existing workloads. They are outside the default write-regression scope. Full help coverage and real backend write validation are separate checks.
