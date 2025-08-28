// cmd/create/image.go
package create

import (
	"fmt"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
)

var imageCmd = &cobra.Command{
	Use:   "image NAME",
	Short: "Create a new image",
	Long: `Create a new image in ZStack.

Examples:
  # Create a root volume template image
  zstack-cli create image my-image --url http://example.com/image.qcow2 --backup-storage bs-uuid1,bs-uuid2 --media-type RootVolumeTemplate

  # Create an ISO image
  zstack-cli create image my-iso --url http://example.com/image.iso --backup-storage bs-uuid1 --media-type ISO --platform Linux
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// 获取标志值
		url, _ := cmd.Flags().GetString("url")
		backupStorageStr, _ := cmd.Flags().GetString("backup-storage")
		mediaType, _ := cmd.Flags().GetString("media-type")
		format, _ := cmd.Flags().GetString("format")
		platform, _ := cmd.Flags().GetString("platform")
		guestOsType, _ := cmd.Flags().GetString("guest-os-type")
		architecture, _ := cmd.Flags().GetString("architecture")
		description, _ := cmd.Flags().GetString("description")
		resourceUuid, _ := cmd.Flags().GetString("resource-uuid")
		system, _ := cmd.Flags().GetBool("system")
		virtio, _ := cmd.Flags().GetBool("virtio")

		// 验证必要参数
		if url == "" {
			fmt.Println("Error: required flag --url not set")
			cmd.Help()
			return
		}

		if backupStorageStr == "" {
			fmt.Println("Error: required flag --backup-storage not set")
			cmd.Help()
			return
		}

		if mediaType == "" {
			fmt.Println("Error: required flag --media-type not set")
			cmd.Help()
			return
		}

		// 解析备份存储UUID列表
		backupStorageUuids := strings.Split(backupStorageStr, ",")

		// 构建API参数
		imageParam := param.AddImageParam{
			BaseParam: param.BaseParam{},
			Params: param.AddImageDetailParam{
				Name:               name,
				Description:        description,
				Url:                url,
				MediaType:          param.MediaType(mediaType),
				GuestOsType:        guestOsType,
				System:             system,
				Format:             param.ImageFormat(format),
				Platform:           platform,
				BackupStorageUuids: backupStorageUuids,
				ResourceUuid:       resourceUuid,
				Architecture:       param.Architecture(architecture),
				Virtio:             virtio,
			},
		}

		// 如果是dry-run模式，打印参数并返回
		if dryRunFlag {
			output.PrintDryRun(imageParam, outputFlag)
			return
		}

		// 创建客户端并发送请求
		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		// 发送请求
		fmt.Printf("Creating image '%s'...\n", name)
		result, err := zsClient.AddImage(imageParam)
		if err != nil {
			fmt.Printf("Error creating image: %s\n", err.Error())
			return
		}

		// 根据输出格式打印结果
		output.PrintOperationResult("Image", result, outputFlag)
	},
}

func init() {
	// 添加 image 命令到 create 命令
	CreateCmd.AddCommand(imageCmd)

	// 添加标志
	imageCmd.Flags().String("url", "", "URL of the image (required)")
	imageCmd.Flags().String("backup-storage", "", "Comma-separated list of backup storage UUIDs (required)")
	imageCmd.Flags().String("media-type", "", "Media type of the image: RootVolumeTemplate, ISO, or DataVolumeTemplate (required)")
	imageCmd.Flags().String("format", "qcow2", "Format of the image (e.g., qcow2, raw)")
	imageCmd.Flags().String("platform", "", "Platform of the image (Linux, Windows, WindowsVirtio, Other, Paravirtualization)")
	imageCmd.Flags().String("guest-os-type", "", "Guest OS type of the image")
	imageCmd.Flags().String("architecture", "x86_64", "Architecture of the image (x86_64, aarch64, mips64el)")
	imageCmd.Flags().String("description", "", "Description of the image")
	imageCmd.Flags().String("resource-uuid", "", "Resource UUID for the image")
	imageCmd.Flags().Bool("system", false, "Whether it is a system image")
	imageCmd.Flags().Bool("virtio", false, "Whether to use virtio drivers")
}
