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

package create

import (
	"fmt"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

var zoneCmd = &cobra.Command{
	Use:   "zone <name>",
	Short: "Create a zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneParam := buildCreateZoneParam(cmd, args[0])
		if dryRunFlag {
			return printCreateDryRun(zoneParam)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		result, err := cli.CreateZone(zoneParam)
		if err != nil {
			return err
		}
		return utils.PrintOperationResult("zone", result, outputFlag)
	},
}

var clusterCmd = &cobra.Command{
	Use:   "cluster <name>",
	Short: "Create a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterParam, err := buildCreateClusterParam(cmd, args[0])
		if err != nil {
			return err
		}
		if dryRunFlag {
			return printCreateDryRun(clusterParam)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		zone, err := zstackclient.ResolveOneResource(cli, "zone", clusterParam.Params.ZoneUuid)
		if err != nil {
			return err
		}
		clusterParam.Params.ZoneUuid = zone.UUID

		result, err := cli.CreateCluster(clusterParam)
		if err != nil {
			return err
		}
		return utils.PrintOperationResult("cluster", result, outputFlag)
	},
}

var l2NetworkCmd = &cobra.Command{
	Use:   "l2-network <name>",
	Short: "Create an L2 VLAN network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l2Param, err := buildCreateL2NetworkParam(cmd, args[0])
		if err != nil {
			return err
		}
		if dryRunFlag {
			return printCreateDryRun(l2Param)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		zone, err := zstackclient.ResolveOneResource(cli, "zone", l2Param.Params.ZoneUuid)
		if err != nil {
			return err
		}
		l2Param.Params.ZoneUuid = zone.UUID

		result, err := cli.CreateL2VlanNetwork(l2Param)
		if err != nil {
			return err
		}
		return utils.PrintOperationResult("l2-network", result, outputFlag)
	},
}

var l3NetworkCmd = &cobra.Command{
	Use:   "l3-network <name>",
	Short: "Create an L3 network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l3Param, err := buildCreateL3NetworkParam(cmd, args[0])
		if err != nil {
			return err
		}
		if dryRunFlag {
			return printCreateDryRun(l3Param)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		l2, err := zstackclient.ResolveOneResource(cli, "l2-network", l3Param.Params.L2NetworkUuid)
		if err != nil {
			return err
		}
		l3Param.Params.L2NetworkUuid = l2.UUID

		result, err := cli.CreateL3Network(l3Param)
		if err != nil {
			return err
		}
		return utils.PrintOperationResult("l3-network", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(zoneCmd, clusterCmd, l2NetworkCmd, l3NetworkCmd)

	zoneCmd.Flags().StringP("description", "d", "", "Detailed description of the zone")
	zoneCmd.Flags().Bool("default", false, "Make this the default zone")
	zoneCmd.Flags().String("resource-uuid", "", "Resource UUID")
	zoneCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	zoneCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	zoneCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")

	clusterCmd.Flags().StringP("description", "d", "", "Detailed description of the cluster")
	clusterCmd.Flags().String("zone", "", "Zone name or UUID")
	clusterCmd.Flags().String("hypervisor", "KVM", "Hypervisor type")
	clusterCmd.Flags().String("type", "", "Cluster type")
	clusterCmd.Flags().String("architecture", "", "Cluster architecture")
	clusterCmd.Flags().String("resource-uuid", "", "Resource UUID")
	clusterCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	clusterCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	clusterCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")

	l2NetworkCmd.Flags().StringP("description", "d", "", "Detailed description of the L2 network")
	l2NetworkCmd.Flags().String("zone", "", "Zone name or UUID")
	l2NetworkCmd.Flags().Int("vlan", 0, "VLAN ID")
	l2NetworkCmd.Flags().String("physical-interface", "", "Physical interface")
	l2NetworkCmd.Flags().String("type", "", "L2 network type")
	l2NetworkCmd.Flags().String("vswitch-type", "", "vSwitch type")
	l2NetworkCmd.Flags().Bool("isolated", false, "Whether the L2 network is isolated")
	l2NetworkCmd.Flags().String("pvlan", "", "Private VLAN")
	l2NetworkCmd.Flags().String("resource-uuid", "", "Resource UUID")
	l2NetworkCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	l2NetworkCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	l2NetworkCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")

	l3NetworkCmd.Flags().StringP("description", "d", "", "Detailed description of the L3 network")
	l3NetworkCmd.Flags().String("l2-network", "", "L2 network name or UUID")
	l3NetworkCmd.Flags().String("type", "", "L3 network type")
	l3NetworkCmd.Flags().String("category", "", "Network category")
	l3NetworkCmd.Flags().Int("ip-version", 0, "IP version")
	l3NetworkCmd.Flags().Bool("system", false, "Whether this is a system network")
	l3NetworkCmd.Flags().String("dns-domain", "", "DNS domain")
	l3NetworkCmd.Flags().Bool("enable-ipam", false, "Enable IPAM")
	l3NetworkCmd.Flags().String("resource-uuid", "", "Resource UUID")
	l3NetworkCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	l3NetworkCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	l3NetworkCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateZoneParam(cmd *cobra.Command, name string) param.CreateZoneParam {
	description, _ := cmd.Flags().GetString("description")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateZoneParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.CreateZoneParamDetail{
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			IsDefault:    changedBoolPtr(cmd, "default"),
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}
}

func buildCreateClusterParam(cmd *cobra.Command, name string) (param.CreateClusterParam, error) {
	zone, _ := cmd.Flags().GetString("zone")
	if zone == "" {
		return param.CreateClusterParam{}, fmt.Errorf("required flag --zone not set")
	}
	hypervisor, _ := cmd.Flags().GetString("hypervisor")
	if hypervisor == "" {
		return param.CreateClusterParam{}, fmt.Errorf("required flag --hypervisor not set")
	}
	description, _ := cmd.Flags().GetString("description")
	clusterType, _ := cmd.Flags().GetString("type")
	architecture, _ := cmd.Flags().GetString("architecture")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateClusterParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.CreateClusterParamDetail{
			ZoneUuid:       zone,
			Name:           name,
			Description:    stringPtrIfNotEmpty(description),
			HypervisorType: hypervisor,
			Type:           stringPtrIfNotEmpty(clusterType),
			Architecture:   stringPtrIfNotEmpty(architecture),
			ResourceUuid:   stringPtrIfNotEmpty(resourceUUID),
			TagUuids:       tagUUIDs,
		},
	}, nil
}

func buildCreateL2NetworkParam(cmd *cobra.Command, name string) (param.CreateL2VlanNetworkParam, error) {
	zone, _ := cmd.Flags().GetString("zone")
	if zone == "" {
		return param.CreateL2VlanNetworkParam{}, fmt.Errorf("required flag --zone not set")
	}
	vlan, _ := cmd.Flags().GetInt("vlan")
	if vlan == 0 {
		return param.CreateL2VlanNetworkParam{}, fmt.Errorf("required flag --vlan not set")
	}
	description, _ := cmd.Flags().GetString("description")
	physicalInterface, _ := cmd.Flags().GetString("physical-interface")
	l2Type, _ := cmd.Flags().GetString("type")
	vswitchType, _ := cmd.Flags().GetString("vswitch-type")
	pvlan, _ := cmd.Flags().GetString("pvlan")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateL2VlanNetworkParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.CreateL2VlanNetworkParamDetail{
			Vlan:              vlan,
			Name:              name,
			Description:       stringPtrIfNotEmpty(description),
			ZoneUuid:          zone,
			PhysicalInterface: stringPtrIfNotEmpty(physicalInterface),
			Type:              stringPtrIfNotEmpty(l2Type),
			VSwitchType:       stringPtrIfNotEmpty(vswitchType),
			Isolated:          changedBoolPtr(cmd, "isolated"),
			Pvlan:             stringPtrIfNotEmpty(pvlan),
			ResourceUuid:      stringPtrIfNotEmpty(resourceUUID),
			TagUuids:          tagUUIDs,
		},
	}, nil
}

func buildCreateL3NetworkParam(cmd *cobra.Command, name string) (param.CreateL3NetworkParam, error) {
	l2, _ := cmd.Flags().GetString("l2-network")
	if l2 == "" {
		return param.CreateL3NetworkParam{}, fmt.Errorf("required flag --l2-network not set")
	}
	description, _ := cmd.Flags().GetString("description")
	l3Type, _ := cmd.Flags().GetString("type")
	category, _ := cmd.Flags().GetString("category")
	ipVersion, _ := cmd.Flags().GetInt("ip-version")
	dnsDomain, _ := cmd.Flags().GetString("dns-domain")
	system, _ := cmd.Flags().GetBool("system")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateL3NetworkParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.CreateL3NetworkParamDetail{
			Name:          name,
			Description:   stringPtrIfNotEmpty(description),
			Type:          stringPtrIfNotEmpty(l3Type),
			L2NetworkUuid: l2,
			Category:      stringPtrIfNotEmpty(category),
			IpVersion:     intPtrIfNotZero(ipVersion),
			System:        system,
			DnsDomain:     stringPtrIfNotEmpty(dnsDomain),
			EnableIPAM:    changedBoolPtr(cmd, "enable-ipam"),
			ResourceUuid:  stringPtrIfNotEmpty(resourceUUID),
			TagUuids:      tagUUIDs,
		},
	}, nil
}

func printCreateDryRun(data interface{}) error {
	format := outputFlag
	if format == "" {
		format = "yaml"
	}
	return utils.PrintDryRun(data, format)
}

func changedBoolPtr(cmd *cobra.Command, flag string) *bool {
	value, _ := cmd.Flags().GetBool(flag)
	if !cmd.Flags().Changed(flag) {
		return nil
	}
	return &value
}
