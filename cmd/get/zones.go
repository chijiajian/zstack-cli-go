package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

// 定义格式化的可用区结构体
type FormattedZone struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`

	State string `json:"state" yaml:"state" header:"STATE"`
	Type  string `json:"type" yaml:"type" header:"TYPE"`
}

// zonesCmd 表示 zones 命令
var zonesCmd = &cobra.Command{
	Use:   "zones [name]",
	Short: "List zones",
	Long:  `List all zones in the ZStack cloud platform.`,
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

		// 3. 调用API
		zones, err := zsClient.QueryZone(*queryParam)
		if err != nil {
			fmt.Printf("Error querying zones: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedZone
		for _, zone := range zones {
			formatted := FormattedZone{
				Name:        zone.Name,
				UUID:        zone.UUID,
				Description: zone.Description,

				State: zone.State,
				Type:  zone.Type,
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
	GetCmd.AddCommand(zonesCmd)
	common.AddQueryFlags(zonesCmd)
}
