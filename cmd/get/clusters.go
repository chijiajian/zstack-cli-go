/*
Copyright © 2025 zstack.io
*/
package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

// clustersCmd 表示 clusters 命令
var clustersCmd = &cobra.Command{
	Use:   "clusters [name]",
	Short: "Get ZStack clusters",
	Long: `Display one or many ZStack clusters.

Examples:
  # List all clusters
  zstack-cli get clusters

  # Get a specific cluster by name
  zstack-cli get clusters my-cluster

  # Query clusters with specific conditions
  zstack-cli get clusters --q "name=Cluster1"

  # Query clusters with multiple conditions
  zstack-cli get clusters --q "name=Cluster1" --q "state=Enabled"

  # Limit the number of results and paginate
  zstack-cli get clusters --limit 10 --start 0

  # Output in different formats
  zstack-cli get clusters --output json
  zstack-cli get clusters --output yaml
  zstack-cli get clusters --output text`,
	Run: func(cobraCmd *cobra.Command, args []string) {
		// 1. 获取已通过 session 认证的全局客户端实例
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

		// 3. 使用已认证的客户端执行查询
		clusters, err := zsClient.QueryCluster(*queryParam)
		if err != nil {
			fmt.Printf("Query failed: %s\n", err)

			// 添加更详细的错误诊断
			fmt.Println("\nDebug: Error details:")
			fmt.Printf("Debug: Error type: %T\n", err)
			fmt.Printf("Debug: Query parameters: %+v\n", queryParam)

			// 如果错误是 NotSupportedError，给出更具体的建议
			if err.Error() == "not supported" {
				fmt.Println("\nThe 'QueryCluster' API is not supported. This could be due to:")
				fmt.Println("1. Your ZStack version is not compatible with this SDK")
				fmt.Println("2. Your user account doesn't have permission to query clusters")
				fmt.Println("3. There's an issue with the SDK implementation")
				fmt.Println("\nPlease check your ZStack version and user permissions.")
			}

			return
		}

		// 4. 使用 output 包格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 如果没有找到集群，显示适当的消息
		if len(clusters) == 0 {
			fmt.Println("No clusters found.")
			return
		}

		// 使用 output 包进行输出
		err = output.PrintWithFields(clusters, format, fields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
			return
		}
	},
}

func init() {
	GetCmd.AddCommand(clustersCmd)

	// 使用公共包中的函数添加查询标志
	common.AddQueryFlags(clustersCmd)
}
