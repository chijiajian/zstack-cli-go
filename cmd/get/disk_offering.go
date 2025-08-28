package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

// 定义格式化的磁盘规格结构体，添加适当的标签以便表格正确显示
type FormattedDiskOffering struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	DiskSize          string `json:"diskSize" yaml:"diskSize" header:"DISK SIZE"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	AllocatorStrategy string `json:"allocatorStrategy" yaml:"allocatorStrategy" header:"ALLOCATOR STRATEGY"`
	State             string `json:"state" yaml:"state" header:"STATE"`
	//Description       string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

// diskOfferingsCmd 表示 disk-offerings 命令
var diskOfferingsCmd = &cobra.Command{
	Use:   "disk-offerings [name]",
	Short: "List disk offerings",
	Long:  `List all disk offerings in the ZStack cloud platform.`,
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
		diskOfferings, err := zsClient.QueryDiskOffering(*queryParam)
		if err != nil {
			fmt.Printf("Error querying disk offerings: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedDiskOffering
		for _, offering := range diskOfferings {
			formatted := FormattedDiskOffering{
				Name:              offering.Name,
				UUID:              offering.UUID,
				DiskSize:          utils.FormatDiskSize(offering.DiskSize),
				Type:              offering.Type,
				AllocatorStrategy: offering.AllocatorStrategy,
				State:             offering.State,
				//	Description:       offering.Description,
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
	GetCmd.AddCommand(diskOfferingsCmd)

	common.AddQueryFlags(diskOfferingsCmd)
}
