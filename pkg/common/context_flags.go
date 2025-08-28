package common

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
)

// AddBasicContextFlags 添加基本的上下文相关标志
func AddBasicContextFlags(cmd *cobra.Command) {
	cmd.Flags().String("zone", "", "Filter resources by zone name or UUID")
	cmd.Flags().String("cluster", "", "Filter resources by cluster name or UUID")
	cmd.Flags().String("host", "", "Filter resources by host name or UUID")
}

// ProcessBasicContextFlags 处理基本上下文标志并添加到查询参数中
func ProcessBasicContextFlags(cmd *cobra.Command, queryParam *param.QueryParam) error {
	// 获取客户端
	zsClient := client.GetClient()
	if zsClient == nil {
		return fmt.Errorf("not logged in, please run 'zstack-cli login' first")
	}

	// 处理区域过滤
	zone, _ := cmd.Flags().GetString("zone")
	if zone != "" {
		// 尝试获取区域UUID
		zoneUUID, err := client.GetZoneUUIDByName(zsClient, zone)
		if err != nil {
			return fmt.Errorf("failed to find zone '%s': %v", zone, err)
		}
		queryParam.AddQ(fmt.Sprintf("zoneUuid=%s", zoneUUID))
	}

	// 处理集群过滤
	cluster, _ := cmd.Flags().GetString("cluster")
	if cluster != "" {
		// 尝试获取集群UUID
		clusterUUID, err := client.GetClusterUUIDByName(zsClient, cluster)
		if err != nil {
			return fmt.Errorf("failed to find cluster '%s': %v", cluster, err)
		}
		queryParam.AddQ(fmt.Sprintf("clusterUuid=%s", clusterUUID))
	}

	// 处理主机过滤
	host, _ := cmd.Flags().GetString("host")
	if host != "" {
		// 尝试获取主机UUID
		hostUUID, err := client.GetHostUUIDByName(zsClient, host)
		if err != nil {
			return fmt.Errorf("failed to find host '%s': %v", host, err)
		}
		queryParam.AddQ(fmt.Sprintf("hostUuid=%s", hostUUID))
	}

	return nil
}
