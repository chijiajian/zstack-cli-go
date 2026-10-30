// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package del

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	sdkClient "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

type deleteNetworkRequest struct {
	Resource string `json:"resource" yaml:"resource"`
	Target   string `json:"target" yaml:"target"`
}

var DeleteVipsCmd = &cobra.Command{
	Use:     "vip <name-or-uuid>",
	Aliases: []string{"vips"},
	Short:   "Delete a VIP",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "vip", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteVip(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteEipsCmd = &cobra.Command{
	Use:     "eip <name-or-uuid>",
	Aliases: []string{"eips"},
	Short:   "Delete an EIP",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "eip", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteEip(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteSecurityGroupsCmd = &cobra.Command{
	Use:     "security-group <name-or-uuid>",
	Aliases: []string{"security-groups", "sg"},
	Short:   "Delete a security group",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "security-group", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteSecurityGroup(uuid, param.DeleteModePermissive)
		})
	},
}

var DeletePortForwardingCmd = &cobra.Command{
	Use:     "port-forwarding <name-or-uuid>",
	Aliases: []string{"port-forwarding-rules", "pf"},
	Short:   "Delete a port forwarding rule",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "port-forwarding", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeletePortForwardingRule(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteLoadBalancersCmd = &cobra.Command{
	Use:     "load-balancer <name-or-uuid>",
	Aliases: []string{"load-balancers", "lb"},
	Short:   "Delete a load balancer",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "load-balancer", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteLoadBalancer(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteLoadBalancerListenersCmd = &cobra.Command{
	Use:     "load-balancer-listener <name-or-uuid>",
	Aliases: []string{"load-balancer-listeners", "lb-listener"},
	Short:   "Delete a load balancer listener",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "load-balancer-listener", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteLoadBalancerListener(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteLoadBalancerServerGroupsCmd = &cobra.Command{
	Use:     "load-balancer-server-group <name-or-uuid>",
	Aliases: []string{"load-balancer-server-groups", "lb-server-group"},
	Short:   "Delete a load balancer server group",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "load-balancer-server-group", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteLoadBalancerServerGroup(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteVmScriptsCmd = &cobra.Command{
	Use:     "vm-script <name-or-uuid>",
	Aliases: []string{"vm-scripts"},
	Short:   "Delete a VM script",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "vm-script", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteGuestVmScript(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteZonesCmd = &cobra.Command{
	Use:     "zone <name-or-uuid>",
	Aliases: []string{"zones"},
	Short:   "Delete a zone",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "zone", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteZone(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteClustersCmd = &cobra.Command{
	Use:     "cluster <name-or-uuid>",
	Aliases: []string{"clusters"},
	Short:   "Delete a cluster",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "cluster", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteCluster(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteHostsCmd = &cobra.Command{
	Use:     "host <name-or-uuid>",
	Aliases: []string{"hosts"},
	Short:   "Delete a host",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "host", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteHost(uuid, param.DeleteModePermissive)
		})
	},
}

var DeletePrimaryStoragesCmd = &cobra.Command{
	Use:     "primary-storage <name-or-uuid>",
	Aliases: []string{"primary-storages", "ps"},
	Short:   "Delete a primary storage",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "primary-storage", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeletePrimaryStorage(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteBackupStoragesCmd = &cobra.Command{
	Use:     "backup-storage <name-or-uuid>",
	Aliases: []string{"backup-storages", "bs", "image-storage", "image-storages"},
	Short:   "Delete a backup storage",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "backup-storage", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteBackupStorage(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteL2NetworksCmd = &cobra.Command{
	Use:     "l2-network <name-or-uuid>",
	Aliases: []string{"l2-networks"},
	Short:   "Delete an L2 network",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "l2-network", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteL2Network(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteL3NetworksCmd = &cobra.Command{
	Use:     "l3-network <name-or-uuid>",
	Aliases: []string{"l3-networks"},
	Short:   "Delete an L3 network",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "l3-network", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteL3Network(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteIPRangesCmd = &cobra.Command{
	Use:     "ip-range <name-or-uuid>",
	Aliases: []string{"ip-ranges"},
	Short:   "Delete an IP range",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "ip-range", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteIpRange(uuid, param.DeleteModePermissive)
		})
	},
}

var DeleteAlarmsCmd = &cobra.Command{
	Use:     "alarm <name-or-uuid>",
	Aliases: []string{"alarms"},
	Short:   "Delete an alarm",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteNetworkResource(cmd, "alarm", args[0], func(cli *sdkClient.ZSClient, uuid string) error {
			return cli.DeleteAlarm(uuid, param.DeleteModePermissive)
		})
	},
}

func init() {
	DeleteCmd.AddCommand(
		DeleteAlarmsCmd,
		DeleteVipsCmd,
		DeleteEipsCmd,
		DeleteSecurityGroupsCmd,
		DeletePortForwardingCmd,
		DeleteLoadBalancersCmd,
		DeleteLoadBalancerListenersCmd,
		DeleteLoadBalancerServerGroupsCmd,
		DeleteVmScriptsCmd,
		DeleteZonesCmd,
		DeleteClustersCmd,
		DeleteHostsCmd,
		DeletePrimaryStoragesCmd,
		DeleteBackupStoragesCmd,
		DeleteL2NetworksCmd,
		DeleteL3NetworksCmd,
		DeleteIPRangesCmd,
	)
	DeleteAlarmsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteVipsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteEipsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteSecurityGroupsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeletePortForwardingCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteLoadBalancersCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteLoadBalancerListenersCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteLoadBalancerServerGroupsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteVmScriptsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteZonesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteClustersCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteHostsCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeletePrimaryStoragesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteBackupStoragesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteL2NetworksCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteL3NetworksCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteIPRangesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
}

func deleteNetworkResource(cmd *cobra.Command, resourceType string, target string, deleteFn func(*sdkClient.ZSClient, string) error) error {
	if dryRunFlag {
		format := outputFlag
		if format == "" {
			format = "yaml"
		}
		return utils.PrintDryRun(deleteNetworkRequest{Resource: resourceType, Target: target}, format)
	}

	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}

	resource, err := zstackclient.ResolveOneResource(cli, resourceType, target)
	if err != nil {
		return err
	}

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes && !confirmNetworkDelete(resourceType, resource.Name, resource.UUID) {
		fmt.Println("Aborted by user.")
		return nil
	}

	if err := deleteFn(cli, resource.UUID); err != nil {
		return err
	}
	fmt.Printf("Deleted %s: %s (%s)\n", resourceType, resource.Name, resource.UUID)
	return nil
}

func confirmNetworkDelete(resourceType string, name string, uuid string) bool {
	fmt.Printf("The following %s will be deleted:\n", resourceType)
	fmt.Printf("  - %s (%s)\n", name, uuid)
	fmt.Printf("Are you sure you want to delete this %s? Type 'yes' to confirm: ", resourceType)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "yes" || input == "y"
}
