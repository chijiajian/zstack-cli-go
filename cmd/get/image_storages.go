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
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

// 定义格式化的备份存储结构体，添加适当的标签以便表格正确显示
type FormattedBackupStorage struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	URL               string `json:"url" yaml:"url" header:"URL"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	State             string `json:"state" yaml:"state" header:"STATE"`
	Status            string `json:"status" yaml:"status" header:"STATUS"`
	TotalCapacity     string `json:"totalCapacity" yaml:"totalCapacity" header:"TOTAL CAPACITY"`
	AvailableCapacity string `json:"availableCapacity" yaml:"availableCapacity" header:"AVAILABLE CAPACITY"`
}

// 转换字节为可读的容量表示
func formatCapacity(bytes int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
		PB
	)

	switch {
	case bytes >= int64(PB):
		return fmt.Sprintf("%.2f PB", float64(bytes)/PB)
	case bytes >= int64(TB):
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= int64(GB):
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= int64(MB):
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= int64(KB):
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// imageStoragesCmd 表示 image-storages 命令
var imageStoragesCmd = &cobra.Command{
	Use:     "image-storages [name]",
	Aliases: []string{"image-storage", "backup-storage", "backup-storages"},
	Short:   "Get ZStack image storages (backup storages)",
	Long: `Display one or many ZStack image storages (backup storages).

Examples:
  # List all image storages
  zstack-cli get image-storages

  # Get a specific image storage by name
  zstack-cli get image-storages my-image-storage

  # Query image storages with specific conditions
  zstack-cli get image-storages --q "name=ImageStore1"

  # Query image storages with multiple conditions
  zstack-cli get image-storages --q "name=ImageStore1" --q "state=Enabled"

  # Limit the number of results and paginate
  zstack-cli get image-storages --limit 10 --start 0

  # Output in different formats
  zstack-cli get image-storages --output json
  zstack-cli get image-storages --output yaml
  zstack-cli get image-storages --output text`,
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
		var backupStorages []view.BackupStorageInventoryView
		backupStorages, err = zsClient.QueryBackupStorage(*queryParam)

		if err != nil {
			fmt.Printf("Query failed: %s\n", err)
			fmt.Println("\nDebug: Error details:")
			fmt.Printf("Debug: Error type: %T\n", err)
			fmt.Printf("Debug: Query parameters: %+v\n", queryParam)
			return
		}
		// 如果没有找到镜像存储，显示适当的消息
		if len(backupStorages) == 0 {
			fmt.Println("No image storages found.")
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedBackupStorage
		for _, storage := range backupStorages {
			formatted := FormattedBackupStorage{
				Name:              storage.Name,
				UUID:              storage.UUID,
				URL:               storage.Url,
				Type:              storage.Type,
				State:             storage.State,
				Status:            storage.Status,
				TotalCapacity:     formatCapacity(storage.TotalCapacity),
				AvailableCapacity: formatCapacity(storage.AvailableCapacity),
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
	GetCmd.AddCommand(imageStoragesCmd)

	// 使用公共包中的函数添加查询标志
	common.AddQueryFlags(imageStoragesCmd)
}
