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

package resources

import (
	"fmt"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

type infrastructureRequest struct {
	Resource string      `json:"resource" yaml:"resource"`
	Target   string      `json:"target,omitempty" yaml:"target,omitempty"`
	Related  string      `json:"related,omitempty" yaml:"related,omitempty"`
	Params   interface{} `json:"params,omitempty" yaml:"params,omitempty"`
}

var HostCmd = &cobra.Command{
	Use:   "host",
	Short: "Manage hosts",
	Long:  "Add and manage ZStack hosts.",
}

var PrimaryStorageCmd = &cobra.Command{
	Use:     "primary-storage",
	Aliases: []string{"ps"},
	Short:   "Manage primary storage",
	Long:    "Add primary storage and manage cluster attachments.",
}

var BackupStorageCmd = &cobra.Command{
	Use:     "backup-storage",
	Aliases: []string{"bs"},
	Short:   "Manage backup storage",
	Long:    "Add backup storage and manage zone attachments.",
}

var L2NetworkCmd = &cobra.Command{
	Use:   "l2-network",
	Short: "Manage L2 networks",
	Long:  "Attach and detach ZStack L2 networks to clusters.",
}

var L3NetworkCmd = &cobra.Command{
	Use:   "l3-network",
	Short: "Manage L3 networks",
	Long:  "Manage IP ranges, DNS, and host routes on ZStack L3 networks.",
}

var hostAddKVMCmd = &cobra.Command{
	Use:   "add-kvm <name>",
	Short: "Add a KVM host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		hostParam, err := buildAddKVMHostParam(cmd, args[0])
		if err != nil {
			return err
		}
		if isInfrastructureDryRun(cmd) {
			return printInfrastructureDryRun(cmd, hostParam)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		cluster, err := zstackclient.ResolveOneResource(cli, "cluster", hostParam.Params.ClusterUuid)
		if err != nil {
			return err
		}
		hostParam.Params.ClusterUuid = cluster.UUID

		result, err := cli.AddKVMHost(hostParam)
		if err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("host/%s added", result.Name), result)
	},
}

var primaryStorageAddNFSCmd = &cobra.Command{
	Use:   "add-nfs <name>",
	Short: "Add an NFS primary storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		psParam, err := buildAddNFSPrimaryStorageParam(cmd, args[0])
		if err != nil {
			return err
		}
		if isInfrastructureDryRun(cmd) {
			return printInfrastructureDryRun(cmd, psParam)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		zone, err := zstackclient.ResolveOneResource(cli, "zone", psParam.Params.ZoneUuid)
		if err != nil {
			return err
		}
		psParam.Params.ZoneUuid = zone.UUID

		result := view.PrimaryStorageInventoryView{}
		if err := cli.Post("v1/primary-storage/nfs", psParam, &result); err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("primary-storage/%s added", result.Name), result)
	},
}

var primaryStorageAttachClusterCmd = &cobra.Command{
	Use:   "attach-cluster <name-or-uuid>",
	Short: "Attach a primary storage to a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPrimaryStorageCluster(cmd, args[0], true)
	},
}

var primaryStorageDetachClusterCmd = &cobra.Command{
	Use:   "detach-cluster <name-or-uuid>",
	Short: "Detach a primary storage from a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPrimaryStorageCluster(cmd, args[0], false)
	},
}

var backupStorageAddImageStoreCmd = &cobra.Command{
	Use:   "add-image-store <name>",
	Short: "Add an image store backup storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bsParam, err := buildAddImageStoreBackupStorageParam(cmd, args[0])
		if err != nil {
			return err
		}
		if isInfrastructureDryRun(cmd) {
			return printInfrastructureDryRun(cmd, bsParam)
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		result, err := cli.AddImageStoreBackupStorage(bsParam)
		if err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("backup-storage/%s added", result.Name), result)
	},
}

var backupStorageAttachZoneCmd = &cobra.Command{
	Use:   "attach-zone <name-or-uuid>",
	Short: "Attach a backup storage to a zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBackupStorageZone(cmd, args[0], true)
	},
}

var backupStorageDetachZoneCmd = &cobra.Command{
	Use:   "detach-zone <name-or-uuid>",
	Short: "Detach a backup storage from a zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBackupStorageZone(cmd, args[0], false)
	},
}

var l2NetworkAttachClusterCmd = &cobra.Command{
	Use:   "attach-cluster <name-or-uuid>",
	Short: "Attach an L2 network to a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runL2NetworkCluster(cmd, args[0], true)
	},
}

var l2NetworkDetachClusterCmd = &cobra.Command{
	Use:   "detach-cluster <name-or-uuid>",
	Short: "Detach an L2 network from a cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runL2NetworkCluster(cmd, args[0], false)
	},
}

var l3NetworkAddIPRangeCmd = &cobra.Command{
	Use:   "add-ip-range <name-or-uuid>",
	Short: "Add an IP range to an L3 network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cidr, _ := cmd.Flags().GetString("cidr")
		if cidr != "" {
			return runL3NetworkAddIPRangeByCIDR(cmd, args[0], cidr)
		}
		return runL3NetworkAddIPRange(cmd, args[0])
	},
}

var l3NetworkAddDNSCmd = &cobra.Command{
	Use:   "add-dns <name-or-uuid>",
	Short: "Add DNS to an L3 network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dns, _ := cmd.Flags().GetString("dns")
		if dns == "" {
			return fmt.Errorf("required flag --dns not set")
		}
		if isInfrastructureDryRun(cmd) {
			return printInfrastructureDryRun(cmd, infrastructureRequest{Resource: "l3-network", Target: args[0], Related: dns})
		}
		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		l3, err := zstackclient.ResolveOneResource(cli, "l3-network", args[0])
		if err != nil {
			return err
		}
		result, err := cli.AddDnsToL3Network(l3.UUID, param.AddDnsToL3NetworkParam{
			Params: param.AddDnsToL3NetworkParamDetail{Dns: dns},
		})
		if err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("l3-network/%s added dns/%s", l3.Name, dns), result)
	},
}

var l3NetworkRemoveDNSCmd = &cobra.Command{
	Use:   "remove-dns <name-or-uuid>",
	Short: "Remove DNS from an L3 network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dns, _ := cmd.Flags().GetString("dns")
		if dns == "" {
			return fmt.Errorf("required flag --dns not set")
		}
		if isInfrastructureDryRun(cmd) {
			return printInfrastructureDryRun(cmd, infrastructureRequest{Resource: "l3-network", Target: args[0], Related: dns})
		}
		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		l3, err := zstackclient.ResolveOneResource(cli, "l3-network", args[0])
		if err != nil {
			return err
		}
		if err := cli.RemoveDnsFromL3Network(l3.UUID, dns, param.DeleteModePermissive); err != nil {
			return err
		}
		fmt.Printf("l3-network/%s removed dns/%s\n", l3.Name, dns)
		return nil
	},
}

var l3NetworkAddHostRouteCmd = &cobra.Command{
	Use:   "add-host-route <name-or-uuid>",
	Short: "Add a host route to an L3 network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix, _ := cmd.Flags().GetString("prefix")
		nexthop, _ := cmd.Flags().GetString("nexthop")
		if prefix == "" {
			return fmt.Errorf("required flag --prefix not set")
		}
		if nexthop == "" {
			return fmt.Errorf("required flag --nexthop not set")
		}
		routeParam := param.AddHostRouteToL3NetworkParam{
			Params: param.AddHostRouteToL3NetworkParamDetail{Prefix: prefix, Nexthop: nexthop},
		}
		if isInfrastructureDryRun(cmd) {
			return printInfrastructureDryRun(cmd, infrastructureRequest{Resource: "l3-network", Target: args[0], Params: routeParam})
		}
		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		l3, err := zstackclient.ResolveOneResource(cli, "l3-network", args[0])
		if err != nil {
			return err
		}
		result, err := cli.AddHostRouteToL3Network(l3.UUID, routeParam)
		if err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("l3-network/%s added host-route/%s", l3.Name, prefix), result)
	},
}

var l3NetworkRemoveHostRouteCmd = &cobra.Command{
	Use:   "remove-host-route <name-or-uuid>",
	Short: "Remove a host route from an L3 network",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix, _ := cmd.Flags().GetString("prefix")
		if prefix == "" {
			return fmt.Errorf("required flag --prefix not set")
		}
		removeParam := param.RemoveHostRouteFromL3NetworkParam{
			Params: param.RemoveHostRouteFromL3NetworkParamDetail{Prefix: prefix},
		}
		if isInfrastructureDryRun(cmd) {
			return printInfrastructureDryRun(cmd, infrastructureRequest{Resource: "l3-network", Target: args[0], Params: removeParam})
		}
		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}
		l3, err := zstackclient.ResolveOneResource(cli, "l3-network", args[0])
		if err != nil {
			return err
		}
		if err := cli.DeleteWithSpec("v1/l3-networks", l3.UUID, "hostroute", fmt.Sprintf("deleteMode=%s", param.DeleteModePermissive), removeParam); err != nil {
			return err
		}
		fmt.Printf("l3-network/%s removed host-route/%s\n", l3.Name, prefix)
		return nil
	},
}

func init() {
	addInfrastructurePersistentFlags(HostCmd)
	addInfrastructurePersistentFlags(PrimaryStorageCmd)
	addInfrastructurePersistentFlags(BackupStorageCmd)
	addInfrastructurePersistentFlags(L2NetworkCmd)
	addInfrastructurePersistentFlags(L3NetworkCmd)

	hostAddKVMCmd.Flags().String("cluster", "", "Cluster name or UUID")
	hostAddKVMCmd.Flags().String("management-ip", "", "Management IP")
	hostAddKVMCmd.Flags().String("username", "root", "SSH username")
	hostAddKVMCmd.Flags().String("password", "", "SSH password")
	hostAddKVMCmd.Flags().Int("ssh-port", 0, "SSH port")
	hostAddKVMCmd.Flags().StringP("description", "d", "", "Detailed description of the host")
	hostAddKVMCmd.Flags().String("resource-uuid", "", "Resource UUID")
	hostAddKVMCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	hostAddKVMCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	hostAddKVMCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")

	addPrimaryStorageFlags(primaryStorageAddNFSCmd)
	addPrimaryStorageFlags(primaryStorageAddLocalCmd)
	primaryStorageAttachClusterCmd.Flags().String("cluster", "", "Cluster name or UUID")
	primaryStorageDetachClusterCmd.Flags().String("cluster", "", "Cluster name or UUID")

	backupStorageAddImageStoreCmd.Flags().String("url", "", "Image store URL")
	backupStorageAddImageStoreCmd.Flags().String("hostname", "", "Image store host name or IP")
	backupStorageAddImageStoreCmd.Flags().String("username", "root", "SSH username")
	backupStorageAddImageStoreCmd.Flags().String("password", "", "SSH password")
	backupStorageAddImageStoreCmd.Flags().Int("ssh-port", 0, "SSH port")
	backupStorageAddImageStoreCmd.Flags().Bool("import-images", false, "Import existing images")
	backupStorageAddImageStoreCmd.Flags().StringP("description", "d", "", "Detailed description of the backup storage")
	backupStorageAddImageStoreCmd.Flags().String("type", "", "Backup storage type")
	backupStorageAddImageStoreCmd.Flags().String("resource-uuid", "", "Resource UUID")
	backupStorageAddImageStoreCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	backupStorageAddImageStoreCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	backupStorageAddImageStoreCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
	backupStorageAttachZoneCmd.Flags().String("zone", "", "Zone name or UUID")
	backupStorageDetachZoneCmd.Flags().String("zone", "", "Zone name or UUID")

	l2NetworkAttachClusterCmd.Flags().String("cluster", "", "Cluster name or UUID")
	l2NetworkAttachClusterCmd.Flags().String("l2-provider-type", "", "L2 provider type")
	l2NetworkDetachClusterCmd.Flags().String("cluster", "", "Cluster name or UUID")

	l3NetworkAddIPRangeCmd.Flags().String("name", "", "IP range name")
	l3NetworkAddIPRangeCmd.Flags().StringP("description", "d", "", "Detailed description of the IP range")
	l3NetworkAddIPRangeCmd.Flags().String("start-ip", "", "Start IP")
	l3NetworkAddIPRangeCmd.Flags().String("end-ip", "", "End IP")
	l3NetworkAddIPRangeCmd.Flags().String("netmask", "", "Netmask")
	l3NetworkAddIPRangeCmd.Flags().String("gateway", "", "Gateway")
	l3NetworkAddIPRangeCmd.Flags().String("cidr", "", "Network CIDR")
	l3NetworkAddIPRangeCmd.Flags().String("type", "", "IP range type")
	l3NetworkAddIPRangeCmd.Flags().String("resource-uuid", "", "Resource UUID")
	l3NetworkAddIPRangeCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	l3NetworkAddIPRangeCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	l3NetworkAddIPRangeCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")

	l3NetworkAddDNSCmd.Flags().String("dns", "", "DNS server")
	l3NetworkRemoveDNSCmd.Flags().String("dns", "", "DNS server")
	l3NetworkAddHostRouteCmd.Flags().String("prefix", "", "Route prefix")
	l3NetworkAddHostRouteCmd.Flags().String("nexthop", "", "Route nexthop")
	l3NetworkRemoveHostRouteCmd.Flags().String("prefix", "", "Route prefix")

	HostCmd.AddCommand(hostAddKVMCmd)
	PrimaryStorageCmd.AddCommand(primaryStorageAddLocalCmd, primaryStorageAddNFSCmd, primaryStorageAttachClusterCmd, primaryStorageDetachClusterCmd)
	BackupStorageCmd.AddCommand(backupStorageAddImageStoreCmd, backupStorageAttachZoneCmd, backupStorageDetachZoneCmd)
	L2NetworkCmd.AddCommand(l2NetworkAttachClusterCmd, l2NetworkDetachClusterCmd)
	L3NetworkCmd.AddCommand(l3NetworkAddIPRangeCmd, l3NetworkAddDNSCmd, l3NetworkRemoveDNSCmd, l3NetworkAddHostRouteCmd, l3NetworkRemoveHostRouteCmd)
}

func addInfrastructurePersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool("dry-run", false, "Preview the API request without sending it")
	cmd.PersistentFlags().StringP("output", "o", "text", "Output format: text, table, json, yaml, or name")
	cmd.PersistentFlags().StringSlice("fields", nil, "Custom fields to display in table output")
}

func buildAddKVMHostParam(cmd *cobra.Command, name string) (param.AddKVMHostParam, error) {
	cluster, _ := cmd.Flags().GetString("cluster")
	managementIP, _ := cmd.Flags().GetString("management-ip")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	if cluster == "" {
		return param.AddKVMHostParam{}, fmt.Errorf("required flag --cluster not set")
	}
	if managementIP == "" {
		return param.AddKVMHostParam{}, fmt.Errorf("required flag --management-ip not set")
	}
	if username == "" {
		return param.AddKVMHostParam{}, fmt.Errorf("required flag --username not set")
	}
	if password == "" {
		return param.AddKVMHostParam{}, fmt.Errorf("required flag --password not set")
	}
	description, _ := cmd.Flags().GetString("description")
	sshPort, _ := cmd.Flags().GetInt("ssh-port")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.AddKVMHostParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.AddKVMHostParamDetail{
			Username:     username,
			Password:     password,
			SshPort:      intPtrIfNotZero(sshPort),
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			ManagementIp: managementIP,
			ClusterUuid:  cluster,
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}, nil
}

func buildAddNFSPrimaryStorageParam(cmd *cobra.Command, name string) (param.AddNfsPrimaryStorageParam, error) {
	url, _ := cmd.Flags().GetString("url")
	zone, _ := cmd.Flags().GetString("zone")
	if url == "" {
		return param.AddNfsPrimaryStorageParam{}, fmt.Errorf("required flag --url not set")
	}
	if zone == "" {
		return param.AddNfsPrimaryStorageParam{}, fmt.Errorf("required flag --zone not set")
	}
	description, _ := cmd.Flags().GetString("description")
	storageType, _ := cmd.Flags().GetString("type")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.AddNfsPrimaryStorageParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.AddNfsPrimaryStorageParamDetail{
			Url:          url,
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			Type:         stringPtrIfNotEmpty(storageType),
			ZoneUuid:     zone,
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}, nil
}

func buildAddImageStoreBackupStorageParam(cmd *cobra.Command, name string) (param.AddImageStoreBackupStorageParam, error) {
	url, _ := cmd.Flags().GetString("url")
	hostname, _ := cmd.Flags().GetString("hostname")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	if url == "" {
		return param.AddImageStoreBackupStorageParam{}, fmt.Errorf("required flag --url not set")
	}
	if hostname == "" {
		return param.AddImageStoreBackupStorageParam{}, fmt.Errorf("required flag --hostname not set")
	}
	if username == "" {
		return param.AddImageStoreBackupStorageParam{}, fmt.Errorf("required flag --username not set")
	}
	if password == "" {
		return param.AddImageStoreBackupStorageParam{}, fmt.Errorf("required flag --password not set")
	}
	description, _ := cmd.Flags().GetString("description")
	sshPort, _ := cmd.Flags().GetInt("ssh-port")
	storageType, _ := cmd.Flags().GetString("type")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.AddImageStoreBackupStorageParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.AddImageStoreBackupStorageParamDetail{
			Hostname:     hostname,
			Username:     username,
			Password:     password,
			SshPort:      intPtrIfNotZero(sshPort),
			Url:          url,
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			Type:         stringPtrIfNotEmpty(storageType),
			ImportImages: infrastructureChangedBoolPtr(cmd, "import-images"),
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}, nil
}

func runPrimaryStorageCluster(cmd *cobra.Command, primaryStorageName string, attach bool) error {
	clusterName, _ := cmd.Flags().GetString("cluster")
	if clusterName == "" {
		return fmt.Errorf("required flag --cluster not set")
	}
	if isInfrastructureDryRun(cmd) {
		return printInfrastructureDryRun(cmd, infrastructureRequest{Resource: "primary-storage", Target: primaryStorageName, Related: clusterName})
	}
	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}
	ps, err := zstackclient.ResolveOneResource(cli, "primary-storage", primaryStorageName)
	if err != nil {
		return err
	}
	cluster, err := zstackclient.ResolveOneResource(cli, "cluster", clusterName)
	if err != nil {
		return err
	}
	if attach {
		result, err := cli.AttachPrimaryStorageToCluster(cluster.UUID, ps.UUID, param.AttachPrimaryStorageToClusterParam{})
		if err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("primary-storage/%s attached cluster/%s", ps.Name, cluster.Name), result)
	}
	if err := cli.DetachPrimaryStorageFromCluster(cluster.UUID, ps.UUID, param.DeleteModePermissive); err != nil {
		return err
	}
	fmt.Printf("primary-storage/%s detached cluster/%s\n", ps.Name, cluster.Name)
	return nil
}

func runBackupStorageZone(cmd *cobra.Command, backupStorageName string, attach bool) error {
	zoneName, _ := cmd.Flags().GetString("zone")
	if zoneName == "" {
		return fmt.Errorf("required flag --zone not set")
	}
	if isInfrastructureDryRun(cmd) {
		return printInfrastructureDryRun(cmd, infrastructureRequest{Resource: "backup-storage", Target: backupStorageName, Related: zoneName})
	}
	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}
	bs, err := zstackclient.ResolveOneResource(cli, "backup-storage", backupStorageName)
	if err != nil {
		return err
	}
	zone, err := zstackclient.ResolveOneResource(cli, "zone", zoneName)
	if err != nil {
		return err
	}
	if attach {
		result, err := cli.AttachBackupStorageToZone(zone.UUID, bs.UUID, param.AttachBackupStorageToZoneParam{})
		if err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("backup-storage/%s attached zone/%s", bs.Name, zone.Name), result)
	}
	if err := cli.DetachBackupStorageFromZone(zone.UUID, bs.UUID, param.DeleteModePermissive); err != nil {
		return err
	}
	fmt.Printf("backup-storage/%s detached zone/%s\n", bs.Name, zone.Name)
	return nil
}

func runL2NetworkCluster(cmd *cobra.Command, l2Name string, attach bool) error {
	clusterName, _ := cmd.Flags().GetString("cluster")
	if clusterName == "" {
		return fmt.Errorf("required flag --cluster not set")
	}
	l2ProviderType, _ := cmd.Flags().GetString("l2-provider-type")
	request := infrastructureRequest{Resource: "l2-network", Target: l2Name, Related: clusterName}
	if l2ProviderType != "" {
		request.Params = map[string]string{"l2ProviderType": l2ProviderType}
	}
	if isInfrastructureDryRun(cmd) {
		return printInfrastructureDryRun(cmd, request)
	}
	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}
	l2, err := zstackclient.ResolveOneResource(cli, "l2-network", l2Name)
	if err != nil {
		return err
	}
	cluster, err := zstackclient.ResolveOneResource(cli, "cluster", clusterName)
	if err != nil {
		return err
	}
	if attach {
		result, err := cli.AttachL2NetworkToCluster(l2.UUID, cluster.UUID, param.AttachL2NetworkToClusterParam{
			Params: param.AttachL2NetworkToClusterParamDetail{L2ProviderType: stringPtrIfNotEmpty(l2ProviderType)},
		})
		if err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("l2-network/%s attached cluster/%s", l2.Name, cluster.Name), result)
	}
	if err := cli.DetachL2NetworkFromCluster(l2.UUID, cluster.UUID, param.DeleteModePermissive); err != nil {
		return err
	}
	fmt.Printf("l2-network/%s detached cluster/%s\n", l2.Name, cluster.Name)
	return nil
}

func runL3NetworkAddIPRange(cmd *cobra.Command, l3Name string) error {
	ipRangeParam, err := buildAddIPRangeParam(cmd)
	if err != nil {
		return err
	}
	if isInfrastructureDryRun(cmd) {
		return printInfrastructureDryRun(cmd, infrastructureRequest{Resource: "l3-network", Target: l3Name, Params: ipRangeParam})
	}
	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}
	l3, err := zstackclient.ResolveOneResource(cli, "l3-network", l3Name)
	if err != nil {
		return err
	}
	result, err := cli.AddIpRange(l3.UUID, ipRangeParam)
	if err != nil {
		return err
	}
	return printInfrastructureResult(cmd, fmt.Sprintf("l3-network/%s added ip-range/%s", l3.Name, result.Name), result)
}

func runL3NetworkAddIPRangeByCIDR(cmd *cobra.Command, l3Name string, cidr string) error {
	ipRangeParam, err := buildAddIPRangeByCIDRParam(cmd, cidr)
	if err != nil {
		return err
	}
	if isInfrastructureDryRun(cmd) {
		return printInfrastructureDryRun(cmd, infrastructureRequest{Resource: "l3-network", Target: l3Name, Params: ipRangeParam})
	}
	cli := zstackclient.GetClient()
	if cli == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}
	l3, err := zstackclient.ResolveOneResource(cli, "l3-network", l3Name)
	if err != nil {
		return err
	}
	result, err := cli.AddIpRangeByNetworkCidr(l3.UUID, ipRangeParam)
	if err != nil {
		return err
	}
	return printInfrastructureResult(cmd, fmt.Sprintf("l3-network/%s added ip-range/%s", l3.Name, result.Name), result)
}

func buildAddIPRangeParam(cmd *cobra.Command) (param.AddIpRangeParam, error) {
	name, _ := cmd.Flags().GetString("name")
	startIP, _ := cmd.Flags().GetString("start-ip")
	endIP, _ := cmd.Flags().GetString("end-ip")
	netmask, _ := cmd.Flags().GetString("netmask")
	if name == "" {
		return param.AddIpRangeParam{}, fmt.Errorf("required flag --name not set")
	}
	if startIP == "" {
		return param.AddIpRangeParam{}, fmt.Errorf("required flag --start-ip not set")
	}
	if endIP == "" {
		return param.AddIpRangeParam{}, fmt.Errorf("required flag --end-ip not set")
	}
	if netmask == "" {
		return param.AddIpRangeParam{}, fmt.Errorf("required flag --netmask not set")
	}
	description, _ := cmd.Flags().GetString("description")
	gateway, _ := cmd.Flags().GetString("gateway")
	ipRangeType, _ := cmd.Flags().GetString("type")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.AddIpRangeParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.AddIpRangeParamDetail{
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			StartIp:      startIP,
			EndIp:        endIP,
			Netmask:      netmask,
			Gateway:      stringPtrIfNotEmpty(gateway),
			IpRangeType:  stringPtrIfNotEmpty(ipRangeType),
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}, nil
}

func buildAddIPRangeByCIDRParam(cmd *cobra.Command, cidr string) (param.AddIpRangeByNetworkCidrParam, error) {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return param.AddIpRangeByNetworkCidrParam{}, fmt.Errorf("required flag --name not set")
	}
	description, _ := cmd.Flags().GetString("description")
	gateway, _ := cmd.Flags().GetString("gateway")
	ipRangeType, _ := cmd.Flags().GetString("type")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")

	return param.AddIpRangeByNetworkCidrParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.AddIpRangeByNetworkCidrParamDetail{
			Name:         name,
			Description:  stringPtrIfNotEmpty(description),
			NetworkCidr:  cidr,
			Gateway:      stringPtrIfNotEmpty(gateway),
			IpRangeType:  stringPtrIfNotEmpty(ipRangeType),
			ResourceUuid: stringPtrIfNotEmpty(resourceUUID),
			TagUuids:     tagUUIDs,
		},
	}, nil
}

func infrastructureChangedBoolPtr(cmd *cobra.Command, flag string) *bool {
	value, _ := cmd.Flags().GetBool(flag)
	if !cmd.Flags().Changed(flag) {
		return nil
	}
	return &value
}

func isInfrastructureDryRun(cmd *cobra.Command) bool {
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	return dryRun
}

func printInfrastructureDryRun(cmd *cobra.Command, data interface{}) error {
	output, _ := cmd.Flags().GetString("output")
	if output == "" || output == "text" {
		output = "yaml"
	}
	return utils.PrintDryRun(data, output)
}

func printInfrastructureResult(cmd *cobra.Command, text string, result interface{}) error {
	return printResourceResult(cmd, text, result)
}
