package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

// 定义格式化的EIP结构体，添加适当的标签以便表格正确显示
type FormattedEip struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	VipIp       string `json:"vipIp" yaml:"vipIp" header:"VIP IP"`
	GuestIp     string `json:"guestIp" yaml:"guestIp" header:"GUEST IP"`
	State       string `json:"state" yaml:"state" header:"STATE"`
	VmNicUuid   string `json:"vmNicUuid" yaml:"vmNicUuid" header:"VM NIC UUID"`
	VipUuid     string `json:"vipUuid" yaml:"vipUuid" header:"VIP UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
}

// eipsCmd 表示 eips 命令
var eipsCmd = &cobra.Command{
	Use:   "eips [name]",
	Short: "List elastic IPs",
	Long:  `List all elastic IPs (EIPs) in the ZStack cloud platform.`,
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
		eips, err := zsClient.QueryEip(*queryParam)
		if err != nil {
			fmt.Printf("Error querying elastic IPs: %s\n", err)
			return
		}

		// 如果没有找到EIP，显示适当的消息
		if len(eips) == 0 {
			fmt.Println("No elastic IPs found.")
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedEip
		for _, eip := range eips {
			formatted := FormattedEip{
				Name:        eip.Name,
				UUID:        eip.UUID,
				VipIp:       eip.VipIp,
				GuestIp:     eip.GuestIp,
				State:       eip.State,
				VmNicUuid:   eip.VmNicUuid,
				VipUuid:     eip.VipUuid,
				Description: eip.Description,
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
	GetCmd.AddCommand(eipsCmd)

	common.AddQueryFlags(eipsCmd)
}
