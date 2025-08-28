package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

// 定义格式化的二层网络结构体
type FormattedL2Network struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	Vlan              int    `json:"vlan" yaml:"vlan" header:"VLAN"`
	PhysicalInterface string `json:"physicalInterface" yaml:"physicalInterface" header:"PHYSICAL INTERFACE"`
	ZoneUuid          string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
}

// l2NetworksCmd 表示 l2-networks 命令
var l2NetworksCmd = &cobra.Command{
	Use:   "l2-networks [name]",
	Short: "List L2 networks",
	Long:  `List all Layer 2 networks in the ZStack cloud platform.`,
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
		l2Networks, err := zsClient.QueryL2Network(*queryParam)
		if err != nil {
			fmt.Printf("Error querying L2 networks: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedL2Network
		for _, network := range l2Networks {
			formatted := FormattedL2Network{
				Name:              network.Name,
				UUID:              network.UUID,
				Type:              network.Type,
				Vlan:              network.Vlan,
				PhysicalInterface: network.PhysicalInterface,
				ZoneUuid:          network.ZoneUuid,
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
	GetCmd.AddCommand(l2NetworksCmd)

	common.AddQueryFlags(l2NetworksCmd)
}
