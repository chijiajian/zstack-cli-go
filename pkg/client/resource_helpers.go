package client

import (
	"fmt"

	sdkClient "github.com/terraform-zstack-modules/zstack-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
)

// GetZoneUUIDByName 根据区域名称获取UUID，如果输入已经是UUID则直接返回
func GetZoneUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	// 首先尝试通过名称查询
	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))

	zones, err := cli.QueryZone(queryParam)
	if err != nil {
		return "", err
	}

	if len(zones) > 0 {
		return zones[0].UUID, nil
	}

	// 如果按名称查询没有结果，尝试按UUID查询
	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	zones, err = cli.QueryZone(queryParam)
	if err != nil {
		return "", err
	}

	if len(zones) > 0 {
		return zones[0].UUID, nil
	}

	return "", fmt.Errorf("zone with name or UUID '%s' not found", nameOrUUID)
}

// GetClusterUUIDByName 根据集群名称获取UUID，如果输入已经是UUID则直接返回
func GetClusterUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	// 首先尝试通过名称查询
	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))

	clusters, err := cli.QueryCluster(queryParam)
	if err != nil {
		return "", err
	}

	if len(clusters) > 0 {
		return clusters[0].Uuid, nil
	}

	// 如果按名称查询没有结果，尝试按UUID查询
	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	clusters, err = cli.QueryCluster(queryParam)
	if err != nil {
		return "", err
	}

	if len(clusters) > 0 {
		return clusters[0].Uuid, nil
	}

	return "", fmt.Errorf("cluster with name or UUID '%s' not found", nameOrUUID)
}

// GetHostUUIDByName 根据主机名称获取UUID，如果输入已经是UUID则直接返回
func GetHostUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	// 首先尝试通过名称查询
	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))

	hosts, err := cli.QueryHost(queryParam)
	if err != nil {
		return "", err
	}

	if len(hosts) > 0 {
		return hosts[0].UUID, nil
	}

	// 如果按名称查询没有结果，尝试按UUID查询
	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	hosts, err = cli.QueryHost(queryParam)
	if err != nil {
		return "", err
	}

	if len(hosts) > 0 {
		return hosts[0].UUID, nil
	}

	return "", fmt.Errorf("host with name or UUID '%s' not found", nameOrUUID)
}
