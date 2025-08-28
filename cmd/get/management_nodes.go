package get

import (
	"fmt"
	"strings"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

// managementNodesCmd 表示 management-nodes 命令
var managementNodesCmd = &cobra.Command{
	Use:   "management-nodes [hostname]",
	Short: "Query management nodes",
	Long:  `Query management nodes in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cobraCmd *cobra.Command, args []string) {
		// 打印调试信息
		fmt.Printf("Debug: Command arguments: %v\n", args)

		// 1. 创建客户端
		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		// 2. 创建查询参数
		queryParam, err := common.BuildQueryParams(cobraCmd, args, "hostName")
		if err != nil {
			fmt.Printf("Error building query parameters: %s\n", err)
			return
		}

		// 打印查询参数调试信息
		fmt.Printf("Debug: Query parameters: %+v\n", queryParam)

		// 3. 调用 API
		nodes, err := zsClient.QueryManagementNode(*queryParam)
		if err != nil {
			fmt.Printf("Error querying management nodes: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)

		// 获取并处理字段过滤参数
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")
		processedFields := []string{}

		// 处理每个字段，分割逗号分隔的值并去除空格
		for _, field := range fields {
			parts := strings.Split(field, ",")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					processedFields = append(processedFields, trimmed)
				}
			}
		}

		// 输出处理后的字段列表，用于调试
		if len(processedFields) > 0 {
			fmt.Printf("Debug: Processed fields: %v\n", processedFields)
		}

		// 准备格式化的数据
		var formattedResults []FormattedManagementNode
		for _, node := range nodes {
			formatted := FormattedManagementNode{
				UUID:      node.UUID,
				HostName:  node.HostName,
				JoinDate:  node.JoinDate.Format(time.RFC3339),
				HeartBeat: node.HeartBeat.Format(time.RFC3339),
			}
			formattedResults = append(formattedResults, formatted)
		}

		// 使用支持字段过滤的 PrintWithFields 函数
		err = output.PrintWithFields(formattedResults, format, processedFields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
		}
	},
}

// FormattedManagementNode 是格式化后的管理节点结构
type FormattedManagementNode struct {
	UUID      string `json:"uuid"`      // 资源UUID，唯一标识资源
	HostName  string `json:"hostName"`  // 主机名
	JoinDate  string `json:"joinDate"`  // 加入日期
	HeartBeat string `json:"heartBeat"` // 心跳时间
}

func init() {
	GetCmd.AddCommand(managementNodesCmd)

	// 添加通用查询标志
	common.AddQueryFlags(managementNodesCmd)

}
