package get

import (
	"fmt"
	"strconv"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

// 定义格式化的Host结构体，添加适当的标签以便表格正确显示
type FormattedHost struct {
	Name            string `json:"name" yaml:"name" header:"NAME"`
	UUID            string `json:"uuid" yaml:"uuid" header:"UUID"`
	ManagementIp    string `json:"managementIp" yaml:"managementIp" header:"MANAGEMENT IP"`
	HypervisorType  string `json:"hypervisorType" yaml:"hypervisorType" header:"HYPERVISOR TYPE"`
	State           string `json:"state" yaml:"state" header:"STATE"`
	Status          string `json:"status" yaml:"status" header:"STATUS"`
	ClusterUuid     string `json:"clusterUuid" yaml:"clusterUuid" header:"CLUSTER UUID"`
	ZoneUuid        string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	TotalCpu        string `json:"totalCpu" yaml:"totalCpu" header:"TOTAL CPU"`
	AvailableCpu    string `json:"availableCpu" yaml:"availableCpu" header:"AVAILABLE CPU"`
	TotalMemory     string `json:"totalMemory" yaml:"totalMemory" header:"TOTAL MEMORY"`
	AvailableMemory string `json:"availableMemory" yaml:"availableMemory" header:"AVAILABLE MEMORY"`
	CpuSockets      int    `json:"cpuSockets" yaml:"cpuSockets" header:"CPU SOCKETS"`
	CpuNum          int    `json:"cpuNum" yaml:"cpuNum" header:"CPU NUM"`
	Architecture    string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	Description     string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

// hostsCmd 表示 hosts 命令
var hostsCmd = &cobra.Command{
	Use:   "hosts [name]",
	Short: "List physical hosts",
	Long:  `List all physical hosts in the ZStack cloud platform.`,
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
		hosts, err := zsClient.QueryHost(*queryParam)
		if err != nil {
			fmt.Printf("Error querying hosts: %s\n", err)
			return
		}

		// 如果没有找到主机，显示适当的消息
		if len(hosts) == 0 {
			fmt.Println("No hosts found.")
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedHost
		for _, host := range hosts {
			formatted := FormattedHost{
				Name:            host.Name,
				UUID:            host.UUID,
				ManagementIp:    host.ManagementIp,
				HypervisorType:  host.HypervisorType,
				State:           host.State,
				Status:          host.Status,
				ClusterUuid:     host.ClusterUuid,
				ZoneUuid:        host.ZoneUuid,
				TotalCpu:        utils.FormatCpuCapacity(host.TotalCpuCapacity),
				AvailableCpu:    utils.FormatCpuCapacity(host.AvailableCpuCapacity),
				TotalMemory:     utils.FormatMemorySize(host.TotalMemoryCapacity),
				AvailableMemory: utils.FormatMemorySize(host.AvailableMemoryCapacity),
				CpuSockets:      host.CpuSockets,
				CpuNum:          host.CpuNum,
				Architecture:    host.Architecture,
				Description:     host.Description,
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

// 格式化CPU容量为易读格式
func formatCpuCapacity(cpuHz int64) string {
	cpuGHz := float64(cpuHz) / 1000000000
	return strconv.FormatFloat(cpuGHz, 'f', 2, 64) + " GHz"
}

func init() {
	GetCmd.AddCommand(hostsCmd)
	common.AddQueryFlags(hostsCmd)
}
