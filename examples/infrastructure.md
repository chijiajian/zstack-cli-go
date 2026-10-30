# 网络与基础设施

先发现实际资源及命令参数；不要从名称猜测 zone/cluster/network 的关联。下面的查询不修改资源，但账户会话可能续期：

```sh
zstack-cli get zones --context lab -o json
zstack-cli get clusters --context lab -o json
zstack-cli get primary-storages --context lab -o json
zstack-cli get l2-networks --context lab -o json
zstack-cli get l3-networks --context lab -o json
zstack-cli get security-groups --context lab -o json
zstack-cli get load-balancers --context lab -o json
```

下面仅展示如何发现写命令，不执行变更。各后端所需参数和限制不同，执行时以帮助及服务器能力为准：

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

基础设施、网络规则和全局配置可能影响既有业务，不属于默认回归脚本的写入范围。这些命令有完整参数参考，但不宣称已做全后端真实写入验证。
