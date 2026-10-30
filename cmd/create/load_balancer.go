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

var loadBalancerCmd = &cobra.Command{
	Use:     "load-balancer <name>",
	Aliases: []string{"lb"},
	Short:   "Create a load balancer",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbParam, err := buildCreateLoadBalancerParam(cmd, args[0])
		if err != nil {
			return err
		}

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(lbParam, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		vip, err := zstackclient.ResolveOneResource(cli, "vip", *lbParam.Params.VipUuid)
		if err != nil {
			return err
		}
		lbParam.Params.VipUuid = &vip.UUID

		if lbParam.Params.Ipv6VipUuid != nil {
			ipv6Vip, err := zstackclient.ResolveOneResource(cli, "vip", *lbParam.Params.Ipv6VipUuid)
			if err != nil {
				return err
			}
			lbParam.Params.Ipv6VipUuid = &ipv6Vip.UUID
		}

		result, err := cli.CreateLoadBalancer(lbParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("load-balancer", result, outputFlag)
	},
}

var loadBalancerListenerCmd = &cobra.Command{
	Use:     "load-balancer-listener <name>",
	Aliases: []string{"lb-listener"},
	Short:   "Create a load balancer listener",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbName, _ := cmd.Flags().GetString("load-balancer")
		if lbName == "" {
			return fmt.Errorf("required flag --load-balancer not set")
		}
		listenerParam, err := buildCreateLoadBalancerListenerParam(cmd, args[0])
		if err != nil {
			return err
		}

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(map[string]interface{}{
				"loadBalancer": lbName,
				"request":      listenerParam,
			}, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		lb, err := zstackclient.ResolveOneResource(cli, "load-balancer", lbName)
		if err != nil {
			return err
		}

		result, err := cli.CreateLoadBalancerListener(lb.UUID, listenerParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("load-balancer-listener", result, outputFlag)
	},
}

var loadBalancerServerGroupCmd = &cobra.Command{
	Use:     "load-balancer-server-group <name>",
	Aliases: []string{"lb-server-group"},
	Short:   "Create a load balancer server group",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbName, _ := cmd.Flags().GetString("load-balancer")
		if lbName == "" {
			return fmt.Errorf("required flag --load-balancer not set")
		}
		groupParam := buildCreateLoadBalancerServerGroupParam(cmd, args[0])

		if dryRunFlag {
			format := outputFlag
			if format == "" {
				format = "yaml"
			}
			return utils.PrintDryRun(map[string]interface{}{
				"loadBalancer": lbName,
				"request":      groupParam,
			}, format)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		lb, err := zstackclient.ResolveOneResource(cli, "load-balancer", lbName)
		if err != nil {
			return err
		}

		result, err := cli.CreateLoadBalancerServerGroup(lb.UUID, groupParam)
		if err != nil {
			return err
		}

		return utils.PrintOperationResult("load-balancer-server-group", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(loadBalancerCmd, loadBalancerListenerCmd, loadBalancerServerGroupCmd)

	loadBalancerCmd.Flags().StringP("description", "d", "", "Detailed description of the load balancer")
	loadBalancerCmd.Flags().String("vip", "", "VIP name or UUID")
	loadBalancerCmd.Flags().String("ipv6-vip", "", "IPv6 VIP name or UUID")
	loadBalancerCmd.Flags().String("type", "", "Load balancer type")
	loadBalancerCmd.Flags().String("resource-uuid", "", "Resource UUID")
	loadBalancerCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	loadBalancerCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	loadBalancerCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")

	loadBalancerListenerCmd.Flags().StringP("description", "d", "", "Detailed description of the listener")
	loadBalancerListenerCmd.Flags().String("load-balancer", "", "Load balancer name or UUID")
	loadBalancerListenerCmd.Flags().String("protocol", "", "Protocol type, for example TCP or HTTP")
	loadBalancerListenerCmd.Flags().Int("port", 0, "Load balancer port")
	loadBalancerListenerCmd.Flags().Int("instance-port", 0, "Backend instance port")
	loadBalancerListenerCmd.Flags().String("certificate", "", "Certificate UUID")
	loadBalancerListenerCmd.Flags().String("health-check-protocol", "", "Health check protocol")
	loadBalancerListenerCmd.Flags().String("health-check-method", "", "Health check method")
	loadBalancerListenerCmd.Flags().String("health-check-uri", "", "Health check URI")
	loadBalancerListenerCmd.Flags().String("health-check-http-code", "", "Health check HTTP code")
	loadBalancerListenerCmd.Flags().String("acl-status", "", "ACL status")
	loadBalancerListenerCmd.Flags().StringSlice("acl", []string{}, "ACL UUID list")
	loadBalancerListenerCmd.Flags().String("acl-type", "", "ACL type")
	loadBalancerListenerCmd.Flags().String("security-policy-type", "", "Security policy type")
	loadBalancerListenerCmd.Flags().StringSlice("http-version", []string{}, "HTTP version list")
	loadBalancerListenerCmd.Flags().String("tcp-proxy-protocol", "", "TCP proxy protocol")
	loadBalancerListenerCmd.Flags().StringSlice("http-compress-algo", []string{}, "HTTP compression algorithm list")
	loadBalancerListenerCmd.Flags().String("resource-uuid", "", "Resource UUID")
	loadBalancerListenerCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	loadBalancerListenerCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	loadBalancerListenerCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")

	loadBalancerServerGroupCmd.Flags().StringP("description", "d", "", "Detailed description of the server group")
	loadBalancerServerGroupCmd.Flags().String("load-balancer", "", "Load balancer name or UUID")
	loadBalancerServerGroupCmd.Flags().Int("ip-version", 0, "IP version")
	loadBalancerServerGroupCmd.Flags().String("resource-uuid", "", "Resource UUID")
	loadBalancerServerGroupCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	loadBalancerServerGroupCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	loadBalancerServerGroupCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateLoadBalancerParam(cmd *cobra.Command, name string) (param.CreateLoadBalancerParam, error) {
	vip, _ := cmd.Flags().GetString("vip")
	if vip == "" {
		return param.CreateLoadBalancerParam{}, fmt.Errorf("required flag --vip not set")
	}
	description, _ := cmd.Flags().GetString("description")
	ipv6Vip, _ := cmd.Flags().GetString("ipv6-vip")
	lbType, _ := cmd.Flags().GetString("type")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateLoadBalancerParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateLoadBalancerParamDetail{
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			VipUuid:      stringPtrIfNotEmpty(vip),
			Ipv6VipUuid:  stringPtrIfNotEmpty(ipv6Vip),
			Type:         stringPtrIfNotEmpty(lbType),
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}, nil
}

func buildCreateLoadBalancerListenerParam(cmd *cobra.Command, name string) (param.CreateLoadBalancerListenerParam, error) {
	protocol, _ := cmd.Flags().GetString("protocol")
	if protocol == "" {
		return param.CreateLoadBalancerListenerParam{}, fmt.Errorf("required flag --protocol not set")
	}
	port, _ := cmd.Flags().GetInt("port")
	if port == 0 {
		return param.CreateLoadBalancerListenerParam{}, fmt.Errorf("required flag --port not set")
	}
	description, _ := cmd.Flags().GetString("description")
	instancePort, _ := cmd.Flags().GetInt("instance-port")
	certificate, _ := cmd.Flags().GetString("certificate")
	healthCheckProtocol, _ := cmd.Flags().GetString("health-check-protocol")
	healthCheckMethod, _ := cmd.Flags().GetString("health-check-method")
	healthCheckURI, _ := cmd.Flags().GetString("health-check-uri")
	healthCheckHTTPCode, _ := cmd.Flags().GetString("health-check-http-code")
	aclStatus, _ := cmd.Flags().GetString("acl-status")
	aclUUIDs, _ := cmd.Flags().GetStringSlice("acl")
	aclType, _ := cmd.Flags().GetString("acl-type")
	securityPolicyType, _ := cmd.Flags().GetString("security-policy-type")
	httpVersions, _ := cmd.Flags().GetStringSlice("http-version")
	tcpProxyProtocol, _ := cmd.Flags().GetString("tcp-proxy-protocol")
	httpCompressAlgos, _ := cmd.Flags().GetStringSlice("http-compress-algo")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateLoadBalancerListenerParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateLoadBalancerListenerParamDetail{
			Name:                name,
			Description:         stringPtrIfNotEmpty(description),
			InstancePort:        intPtrIfNotZero(instancePort),
			LoadBalancerPort:    port,
			Protocol:            stringPtrIfNotEmpty(protocol),
			CertificateUuid:     stringPtrIfNotEmpty(certificate),
			HealthCheckProtocol: stringPtrIfNotEmpty(healthCheckProtocol),
			HealthCheckMethod:   stringPtrIfNotEmpty(healthCheckMethod),
			HealthCheckURI:      stringPtrIfNotEmpty(healthCheckURI),
			HealthCheckHttpCode: stringPtrIfNotEmpty(healthCheckHTTPCode),
			AclStatus:           stringPtrIfNotEmpty(aclStatus),
			AclUuids:            aclUUIDs,
			AclType:             stringPtrIfNotEmpty(aclType),
			SecurityPolicyType:  stringPtrIfNotEmpty(securityPolicyType),
			HttpVersions:        httpVersions,
			TcpProxyProtocol:    stringPtrIfNotEmpty(tcpProxyProtocol),
			HttpCompressAlgos:   httpCompressAlgos,
			ResourceUuid:        stringPtrIfNotEmpty(resourceUUID),
			TagUuids:            tagUUIDs,
		},
	}, nil
}

func buildCreateLoadBalancerServerGroupParam(cmd *cobra.Command, name string) param.CreateLoadBalancerServerGroupParam {
	description, _ := cmd.Flags().GetString("description")
	ipVersion, _ := cmd.Flags().GetInt("ip-version")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.CreateLoadBalancerServerGroupParam{
		BaseParam: param.BaseParam{
			SystemTags: systemTags,
			UserTags:   userTags,
		},
		Params: param.CreateLoadBalancerServerGroupParamDetail{
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			IpVersion:    intPtrIfNotZero(ipVersion),
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}
}
