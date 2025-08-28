package get

import (
	"fmt"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

// 定义格式化的虚拟路由器结构体
type FormattedVirtualRouter struct {
	Name                      string `json:"name" yaml:"name" header:"NAME"`
	UUID                      string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description               string `json:"description" yaml:"description" header:"DESCRIPTION"`
	ApplianceVmType           string `json:"applianceVmType" yaml:"applianceVmType" header:"APPLIANCE VM TYPE"`
	ManagementNetworkUuid     string `json:"managementNetworkUuid" yaml:"managementNetworkUuid" header:"MANAGEMENT NETWORK UUID"`
	DefaultRouteL3NetworkUuid string `json:"defaultRouteL3NetworkUuid" yaml:"defaultRouteL3NetworkUuid" header:"DEFAULT ROUTE L3 NETWORK UUID"`
	Status                    string `json:"status" yaml:"status" header:"STATUS"`
	AgentPort                 int    `json:"agentPort" yaml:"agentPort" header:"AGENT PORT"`
	ZoneUuid                  string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	ClusterUUID               string `json:"clusterUuid" yaml:"clusterUuid" header:"CLUSTER UUID"`
	ImageUUID                 string `json:"imageUuid" yaml:"imageUuid" header:"IMAGE UUID"`
	HostUuid                  string `json:"hostUuid" yaml:"hostUuid" header:"HOST UUID"`
	LastHostUUID              string `json:"lastHostUuid" yaml:"lastHostUuid" header:"LAST HOST UUID"`
	InstanceOfferingUUID      string `json:"instanceOfferingUuid" yaml:"instanceOfferingUuid" header:"INSTANCE OFFERING UUID"`
	RootVolumeUuid            string `json:"rootVolumeUuid" yaml:"rootVolumeUuid" header:"ROOT VOLUME UUID"`
	Platform                  string `json:"platform" yaml:"platform" header:"PLATFORM"`
	DefaultL3NetworkUuid      string `json:"defaultL3NetworkUuid" yaml:"defaultL3NetworkUuid" header:"DEFAULT L3 NETWORK UUID"`
	Type                      string `json:"type" yaml:"type" header:"TYPE"`
	HypervisorType            string `json:"hypervisorType" yaml:"hypervisorType" header:"HYPERVISOR TYPE"`
	MemorySize                string `json:"memorySize" yaml:"memorySize" header:"MEMORY SIZE"`
	CPUNum                    int    `json:"cpuNum" yaml:"cpuNum" header:"CPU NUM"`
	CPUSpeed                  int64  `json:"cpuSpeed" yaml:"cpuSpeed" header:"CPU SPEED"`
	AllocatorStrategy         string `json:"allocatorStrategy" yaml:"allocatorStrategy" header:"ALLOCATOR STRATEGY"`
	State                     string `json:"state" yaml:"state" header:"STATE"`
	HaStatus                  string `json:"haStatus" yaml:"haStatus" header:"HA STATUS"`
	Architecture              string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	IPs                       string `json:"ips" yaml:"ips" header:"IPS"`
}

// virtualRoutersCmd 表示 virtual-routers 命令
var virtualRoutersCmd = &cobra.Command{
	Use:   "virtual-routers [name]",
	Short: "List virtual routers",
	Long:  `List all virtual routers in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cobraCmd *cobra.Command, args []string) {
		// 1. 创建客户端
		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		// 2. 创建查询参数
		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			fmt.Printf("Error building query parameters: %s\n", err)
			return
		}

		// 3. 调用 API
		virtualRouters, err := zsClient.QueryVirtualRouterVm(*queryParam)
		if err != nil {
			fmt.Printf("Error querying virtual routers: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedVirtualRouter
		for _, vr := range virtualRouters {
			// 收集所有IP地址
			var ips []string
			for _, nic := range vr.VMNics {
				if nic.IP != "" {
					ips = append(ips, nic.IP)
				}
			}

			formatted := FormattedVirtualRouter{
				Name:        vr.Name,
				UUID:        vr.UUID,
				Description: vr.Description,

				ApplianceVmType:           vr.ApplianceVmType,
				ManagementNetworkUuid:     vr.ManagementNetworkUuid,
				DefaultRouteL3NetworkUuid: vr.DefaultRouteL3NetworkUuid,
				Status:                    vr.Status,
				AgentPort:                 vr.AgentPort,
				ZoneUuid:                  vr.ZoneUuid,
				ClusterUUID:               vr.ClusterUUID,
				ImageUUID:                 vr.ImageUUID,
				HostUuid:                  vr.HostUuid,
				LastHostUUID:              vr.LastHostUUID,
				InstanceOfferingUUID:      vr.InstanceOfferingUUID,
				RootVolumeUuid:            vr.RootVolumeUuid,
				Platform:                  vr.Platform,
				DefaultL3NetworkUuid:      vr.DefaultL3NetworkUuid,
				Type:                      vr.Type,
				HypervisorType:            vr.HypervisorType,
				MemorySize:                utils.FormatMemorySize(vr.MemorySize),
				CPUNum:                    vr.CPUNum,
				CPUSpeed:                  vr.CPUSpeed,
				AllocatorStrategy:         vr.AllocatorStrategy,
				State:                     vr.State,
				HaStatus:                  vr.HaStatus,
				Architecture:              vr.Architecture,
				IPs:                       strings.Join(ips, ", "),
			}
			formattedResults = append(formattedResults, formatted)
		}

		// 使用 output 包进行输出
		err = output.PrintWithFields(formattedResults, format, fields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
			return
		}

	},
}

func init() {
	GetCmd.AddCommand(virtualRoutersCmd)
	common.AddQueryFlags(virtualRoutersCmd)
}
