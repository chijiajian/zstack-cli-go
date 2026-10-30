package get

import (
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

// Each command supplies only its resource contract and row projection.
func newNetworkQuery[T, R any](name, path, filter string, project func(T) R) *cobra.Command {
	cmd := &cobra.Command{Use: name + " [name]", Short: "List " + name, Args: cobra.MaximumNArgs(1)}
	if name == "security-group-rules" {
		cmd.Use = name
		cmd.Args = cobra.NoArgs
	}
	common.AddQueryFlags(cmd)
	if filter != "" {
		cmd.Flags().String(filter, "", "Filter by "+filter+" name or UUID")
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		params, err := common.BuildQueryParams(cmd, args, "name")
		if err != nil {
			return err
		}
		if filter != "" {
			field := "loadBalancerUuid"
			if filter == "security-group" {
				field = "securityGroupUuid"
			}
			if err := common.ResolveFilter(cmd, params, filter, filter, field); err != nil {
				return err
			}
		}
		result, err := common.Fetch[T](cmd, params, path)
		if err != nil {
			return err
		}
		rows := make([]R, 0, len(result.Items))
		for _, item := range result.Items {
			rows = append(rows, project(item))
		}
		return common.PrintQuery(cmd, rows, result.Total)
	}
	return cmd
}

type SecurityGroupRow struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	State       string `json:"state" yaml:"state" header:"STATE"`
	IpVersion   int    `json:"ipVersion" yaml:"ipVersion" header:"IPVERSION"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

type SecurityGroupRuleRow struct {
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	SecurityGroupUuid string `json:"securityGroupUuid" yaml:"securityGroupUuid" header:"SECURITYGROUPUUID"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	Protocol          string `json:"protocol" yaml:"protocol" header:"PROTOCOL"`
	StartPort         int    `json:"startPort" yaml:"startPort" header:"STARTPORT"`
	EndPort           int    `json:"endPort" yaml:"endPort" header:"ENDPORT"`
	AllowedCidr       string `json:"allowedCidr" yaml:"allowedCidr" header:"ALLOWEDCIDR"`
	Action            string `json:"action" yaml:"action" header:"ACTION"`
	State             string `json:"state" yaml:"state" header:"STATE"`
}

type PortForwardingRuleRow struct {
	Name             string `json:"name" yaml:"name" header:"NAME"`
	UUID             string `json:"uuid" yaml:"uuid" header:"UUID"`
	VipIp            string `json:"vipIp" yaml:"vipIp" header:"VIPIP"`
	GuestIp          string `json:"guestIp" yaml:"guestIp" header:"GUESTIP"`
	ProtocolType     string `json:"protocolType" yaml:"protocolType" header:"PROTOCOLTYPE"`
	VipPortStart     int    `json:"vipPortStart" yaml:"vipPortStart" header:"VIPPORTSTART"`
	VipPortEnd       int    `json:"vipPortEnd" yaml:"vipPortEnd" header:"VIPPORTEND"`
	PrivatePortStart int    `json:"privatePortStart" yaml:"privatePortStart" header:"PRIVATEPORTSTART"`
	PrivatePortEnd   int    `json:"privatePortEnd" yaml:"privatePortEnd" header:"PRIVATEPORTEND"`
	VmNicUuid        string `json:"vmNicUuid" yaml:"vmNicUuid" header:"VMNICUUID"`
	State            string `json:"state" yaml:"state" header:"STATE"`
}

type LoadBalancerRow struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	State       string `json:"state" yaml:"state" header:"STATE"`
	VipUuid     string `json:"vipUuid" yaml:"vipUuid" header:"VIPUUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

type LoadBalancerListenerRow struct {
	Name             string `json:"name" yaml:"name" header:"NAME"`
	UUID             string `json:"uuid" yaml:"uuid" header:"UUID"`
	LoadBalancerUuid string `json:"loadBalancerUuid" yaml:"loadBalancerUuid" header:"LOADBALANCERUUID"`
	Protocol         string `json:"protocol" yaml:"protocol" header:"PROTOCOL"`
	LoadBalancerPort int    `json:"loadBalancerPort" yaml:"loadBalancerPort" header:"LOADBALANCERPORT"`
	InstancePort     int    `json:"instancePort" yaml:"instancePort" header:"INSTANCEPORT"`
	ServerGroupUuid  string `json:"serverGroupUuid" yaml:"serverGroupUuid" header:"SERVERGROUPUUID"`
}

type LoadBalancerServerGroupRow struct {
	Name             string `json:"name" yaml:"name" header:"NAME"`
	UUID             string `json:"uuid" yaml:"uuid" header:"UUID"`
	LoadBalancerUuid string `json:"loadBalancerUuid" yaml:"loadBalancerUuid" header:"LOADBALANCERUUID"`
	IpVersion        int    `json:"ipVersion" yaml:"ipVersion" header:"IPVERSION"`
	Description      string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

func init() {
	GetCmd.AddCommand(
		newNetworkQuery("security-groups", "v1/security-groups", "", func(item view.SecurityGroupInventoryView) SecurityGroupRow {
			return SecurityGroupRow{Name: item.Name, UUID: item.UUID, State: item.State, IpVersion: item.IpVersion, Description: item.Description}
		}),
		newNetworkQuery("security-group-rules", "v1/security-groups/rules", "security-group", func(item view.SecurityGroupRuleInventoryView) SecurityGroupRuleRow {
			return SecurityGroupRuleRow{UUID: item.UUID, SecurityGroupUuid: item.SecurityGroupUuid, Type: item.Type, Protocol: item.Protocol, StartPort: item.StartPort, EndPort: item.EndPort, AllowedCidr: item.AllowedCidr, Action: item.Action, State: item.State}
		}),
		newNetworkQuery("port-forwarding-rules", "v1/port-forwarding", "", func(item view.PortForwardingRuleInventoryView) PortForwardingRuleRow {
			return PortForwardingRuleRow{Name: item.Name, UUID: item.UUID, VipIp: item.VipIp, GuestIp: item.GuestIp, ProtocolType: item.ProtocolType, VipPortStart: item.VipPortStart, VipPortEnd: item.VipPortEnd, PrivatePortStart: item.PrivatePortStart, PrivatePortEnd: item.PrivatePortEnd, VmNicUuid: item.VmNicUuid, State: item.State}
		}),
		newNetworkQuery("load-balancers", "v1/load-balancers", "", func(item view.LoadBalancerInventoryView) LoadBalancerRow {
			return LoadBalancerRow{Name: item.Name, UUID: item.UUID, State: item.State, VipUuid: item.VipUuid, Description: item.Description}
		}),
		newNetworkQuery("load-balancer-listeners", "v1/load-balancers/listeners", "load-balancer", func(item view.LoadBalancerListenerInventoryView) LoadBalancerListenerRow {
			return LoadBalancerListenerRow{Name: item.Name, UUID: item.UUID, LoadBalancerUuid: item.LoadBalancerUuid, Protocol: item.Protocol, LoadBalancerPort: item.LoadBalancerPort, InstancePort: item.InstancePort, ServerGroupUuid: item.ServerGroupUuid}
		}),
		newNetworkQuery("load-balancer-server-groups", "v1/load-balancers/servergroups", "load-balancer", func(item view.LoadBalancerServerGroupInventoryView) LoadBalancerServerGroupRow {
			return LoadBalancerServerGroupRow{Name: item.Name, UUID: item.UUID, LoadBalancerUuid: item.LoadBalancerUuid, IpVersion: item.IpVersion, Description: item.Description}
		}),
	)
}
