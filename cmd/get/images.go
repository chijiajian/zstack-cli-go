package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

// 定义格式化的镜像结构体，只包含第一层属性
type FormattedImage struct {
	Name         string `json:"name" yaml:"name" header:"NAME"`
	UUID         string `json:"uuid" yaml:"uuid" header:"UUID"`
	State        string `json:"state" yaml:"state" header:"STATE"`
	Status       string `json:"status" yaml:"status" header:"STATUS"`
	Size         string `json:"size" yaml:"size" header:"SIZE"`
	ActualSize   string `json:"actualSize" yaml:"actualSize" header:"ACTUAL SIZE"`
	Format       string `json:"format" yaml:"format" header:"FORMAT"`
	MediaType    string `json:"mediaType" yaml:"mediaType" header:"MEDIA TYPE"`
	Platform     string `json:"platform" yaml:"platform" header:"PLATFORM"`
	Architecture string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	Type         string `json:"type" yaml:"type" header:"TYPE"`
	GuestOsType  string `json:"guestOsType" yaml:"guestOsType" header:"GUEST OS TYPE"`
}

// imagesCmd 表示 images 命令
var imagesCmd = &cobra.Command{
	Use:   "images [name]",
	Short: "List images",
	Long:  `List all images in the ZStack cloud platform.`,
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
		images, err := zsClient.QueryImage(*queryParam)
		if err != nil {
			fmt.Printf("Error querying images: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")
		// 准备格式化的数据
		var formattedResults []FormattedImage
		for _, image := range images {
			formatted := FormattedImage{
				Name:         image.Name,
				UUID:         image.UUID,
				State:        image.State,
				Status:       image.Status,
				Size:         utils.FormatDiskSize(image.Size),
				ActualSize:   utils.FormatDiskSize(image.ActualSize),
				Format:       image.Format,
				MediaType:    image.MediaType,
				Platform:     image.Platform,
				Architecture: string(image.Architecture),
				Type:         image.Type,
				GuestOsType:  image.GuestOsType,
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
	GetCmd.AddCommand(imagesCmd)

	common.AddQueryFlags(imagesCmd)
}
