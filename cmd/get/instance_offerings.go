package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

// 定义格式化的实例规格结构体
type FormattedInstanceOffering struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	CpuNum            int    `json:"cpuNum" yaml:"cpuNum" header:"CPU"`
	MemorySize        string `json:"memorySize" yaml:"memorySize" header:"MEMORY"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	AllocatorStrategy string `json:"allocatorStrategy" yaml:"allocatorStrategy" header:"ALLOCATOR STRATEGY"`
	State             string `json:"state" yaml:"state" header:"STATE"`
}

// instanceOfferingsCmd 表示 instance-offerings 命令
var instanceOfferingsCmd = &cobra.Command{
	Use:   "instance-offerings [name]",
	Short: "List instance offerings",
	Long:  `List all instance offerings in the ZStack cloud platform.`,
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
		instanceOfferings, err := zsClient.QueryInstaceOffering(*queryParam)
		if err != nil {
			fmt.Printf("Error querying instance offerings: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedInstanceOffering
		for _, offering := range instanceOfferings {
			formatted := FormattedInstanceOffering{
				Name:              offering.Name,
				UUID:              offering.UUID,
				CpuNum:            offering.CpuNum,
				MemorySize:        utils.FormatMemorySize(offering.MemorySize),
				Type:              offering.Type,
				AllocatorStrategy: offering.AllocatorStrategy,
				State:             offering.State,
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
	GetCmd.AddCommand(instanceOfferingsCmd)

	common.AddQueryFlags(instanceOfferingsCmd)
}
