package common

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
)

// 全局客户端实例
var (
	zsClient    *client.ZSClient
	clientMutex sync.Mutex
)

// SetClient 设置全局客户端实例
func SetClient(client *client.ZSClient) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	zsClient = client
}

// GetClient 获取全局客户端实例
func GetClient() *client.ZSClient {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	return zsClient
}

// AddQueryFlags 添加通用查询标志到命令
func AddQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("q", "q", []string{}, "Query condition, can be specified multiple times")
	cmd.Flags().IntP("limit", "l", 0, "Maximum number of results to return")
	cmd.Flags().IntP("start", "s", 0, "Starting index for results")
	cmd.Flags().Bool("count", false, "Return count of matching records only")
	cmd.Flags().Bool("reply-with-count", false, "Include total count in response")
	cmd.Flags().String("sort", "", "Sort results by field (e.g. '+name' or '-createDate')")
	cmd.Flags().String("group-by", "", "Group results by field")
	//cmd.Flags().StringSlice("fields", []string{}, "Fields to return in response, comma-separated")
	cmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml, or text")
	cmd.Flags().StringSlice("fields", nil, "Fields to display (use comma without spaces: --fields name,uuid,type or multiple flags: --fields name --fields type)")
}

// BuildQueryParams 构建查询参数
func BuildQueryParams(cmd *cobra.Command, args []string, nameField string) (*param.QueryParam, error) {
	queryParam := param.NewQueryParam()

	// 处理单个资源名称查询（如果提供了参数）
	if len(args) == 1 && nameField != "" {
		resourceName := args[0]
		queryParam.AddQ(fmt.Sprintf("%s=%s", nameField, resourceName))
	}

	// 获取查询条件
	qConditions, _ := cmd.Flags().GetStringArray("q")
	for _, q := range qConditions {
		queryParam.AddQ(q)
	}

	// 获取分页参数
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 {
		queryParam.Limit(limit)
	}

	start, _ := cmd.Flags().GetInt("start")
	if start > 0 {
		queryParam.Start(start)
	}

	// 获取其他查询参数
	count, _ := cmd.Flags().GetBool("count")
	if count {
		queryParam.Count(true)
	}

	replyWithCount, _ := cmd.Flags().GetBool("reply-with-count")
	if replyWithCount {
		queryParam.ReplyWithCount(true)
	}

	sort, _ := cmd.Flags().GetString("sort")
	if sort != "" {
		queryParam.Sort(sort)
	}

	groupBy, _ := cmd.Flags().GetString("group-by")
	if groupBy != "" {
		queryParam.GroupBy(groupBy)
	}

	fields, _ := cmd.Flags().GetStringSlice("fields")
	if len(fields) > 0 {
		queryParam.Fields(fields)
	}

	return &queryParam, nil
}
