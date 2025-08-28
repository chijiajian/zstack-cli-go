package get

import (
	"fmt"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

// 定义格式化的云盘结构体
type FormattedVolume struct {
	Name               string  `json:"name" yaml:"name" header:"NAME"`
	UUID               string  `json:"uuid" yaml:"uuid" header:"UUID"`
	Description        string  `json:"description" yaml:"description" header:"DESCRIPTION"`
	PrimaryStorageUUID string  `json:"primaryStorageUuid" yaml:"primaryStorageUuid" header:"PRIMARY STORAGE UUID"`
	VMInstanceUUID     string  `json:"vmInstanceUuid" yaml:"vmInstanceUuid" header:"VM INSTANCE UUID"`
	LastVmInstanceUuid string  `json:"lastVmInstanceUuid" yaml:"lastVmInstanceUuid" header:"LAST VM INSTANCE UUID"`
	DiskOfferingUUID   string  `json:"diskOfferingUuid" yaml:"diskOfferingUuid" header:"DISK OFFERING UUID"`
	RootImageUUID      string  `json:"rootImageUuid" yaml:"rootImageUuid" header:"ROOT IMAGE UUID"`
	InstallPath        string  `json:"installPath" yaml:"installPath" header:"INSTALL PATH"`
	Type               string  `json:"type" yaml:"type" header:"TYPE"`
	Format             string  `json:"format" yaml:"format" header:"FORMAT"`
	Size               string  `json:"size" yaml:"size" header:"SIZE"`
	ActualSize         string  `json:"actualSize" yaml:"actualSize" header:"ACTUAL SIZE"`
	DeviceID           float32 `json:"deviceId" yaml:"deviceId" header:"DEVICE ID"`
	State              string  `json:"state" yaml:"state" header:"STATE"`
	Status             string  `json:"status" yaml:"status" header:"STATUS"`
	IsShareable        bool    `json:"isShareable" yaml:"isShareable" header:"IS SHAREABLE"`
	LastDetachDate     string  `json:"lastDetachDate" yaml:"lastDetachDate" header:"LAST DETACH DATE"`
	Attached           string  `json:"attached" yaml:"attached" header:"ATTACHED"`
}

// volumesCmd 表示 volumes 命令
var volumesCmd = &cobra.Command{
	Use:   "disks [name]",
	Short: "List disks",
	Long:  `List all volumes (disks) in the ZStack cloud platform.`,
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

		// 检查是否请求分页
		usePagination, _ := cobraCmd.Flags().GetBool("pagination")
		var volumes []view.VolumeView
		var total int

		if usePagination {
			// 3a. 调用分页API
			volumes, total, err = zsClient.PageVolume(*queryParam)
			if err != nil {
				fmt.Printf("Error querying volumes: %s\n", err)
				return
			}
			fmt.Printf("Total: %d\n", total)
		} else {
			// 3b. 调用非分页API
			volumes, err = zsClient.QueryVolume(*queryParam)
			if err != nil {
				fmt.Printf("Error querying volumes: %s\n", err)
				return
			}
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedVolume
		for _, volume := range volumes {
			// 判断是否已挂载
			attached := "No"
			if volume.VMInstanceUUID != "" {
				attached = "Yes"
			}

			// 格式化最后分离时间
			lastDetachDate := ""
			var zeroTime time.Time
			if volume.LastDetachDate != zeroTime {
				lastDetachDate = volume.LastDetachDate.Format("2006-01-02 15:04:05")
			}

			formatted := FormattedVolume{
				Name:               volume.Name,
				UUID:               volume.UUID,
				Description:        volume.Description,
				PrimaryStorageUUID: volume.PrimaryStorageUUID,
				VMInstanceUUID:     volume.VMInstanceUUID,
				LastVmInstanceUuid: volume.LastVmInstanceUuid,
				DiskOfferingUUID:   volume.DiskOfferingUUID,
				RootImageUUID:      volume.RootImageUUID,
				InstallPath:        volume.InstallPath,
				Type:               volume.Type,
				Format:             volume.Format,
				Size:               utils.FormatDiskSize(int64(volume.Size)),
				ActualSize:         utils.FormatDiskSize(int64(volume.ActualSize)),
				DeviceID:           volume.DeviceID,
				State:              volume.State,
				Status:             volume.Status,
				IsShareable:        volume.IsShareable,
				LastDetachDate:     lastDetachDate,
				Attached:           attached,
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
	GetCmd.AddCommand(volumesCmd)
	common.AddQueryFlags(volumesCmd)
	volumesCmd.Flags().Bool("pagination", false, "Use pagination when querying volumes")
}
