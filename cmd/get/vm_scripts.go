package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

// 定义格式化的虚拟机脚本结构体
type FormattedVmInstanceScript struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`

	ScriptContent string `json:"scriptContent" yaml:"scriptContent" header:"SCRIPT CONTENT"`
	RenderParams  string `json:"renderParams" yaml:"renderParams" header:"RENDER PARAMS"`
	Platform      string `json:"platform" yaml:"platform" header:"PLATFORM"`
	ScriptType    string `json:"scriptType" yaml:"scriptType" header:"SCRIPT TYPE"`
	ScriptTimeout int    `json:"scriptTimeout" yaml:"scriptTimeout" header:"SCRIPT TIMEOUT (SEC)"`
	EncodingType  string `json:"encodingType" yaml:"encodingType" header:"ENCODING TYPE"`
}

// vmScriptsCmd 表示 vm-scripts 命令
var vmScriptsCmd = &cobra.Command{
	Use:   "vm-scripts [name]",
	Short: "List VM instance scripts",
	Long:  `List all VM instance scripts in the ZStack cloud platform.`,
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

		// 3. 调用 API
		scripts, err := zsClient.QueryVmInstanceScript(*queryParam)
		if err != nil {
			fmt.Printf("Error querying VM instance scripts: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 检查是否需要显示完整的脚本内容
		showFullScript, _ := cobraCmd.Flags().GetBool("show-full-script")

		// 准备格式化的数据
		var formattedResults []FormattedVmInstanceScript
		for _, script := range scripts {
			// 如果不显示完整脚本内容，则截断显示
			scriptContent := script.ScriptContent
			if !showFullScript && len(scriptContent) > 50 {
				scriptContent = scriptContent[:50] + "..."
			}

			formatted := FormattedVmInstanceScript{
				Name:        script.Name,
				UUID:        script.UUID,
				Description: script.Description,

				ScriptContent: scriptContent,
				RenderParams:  script.RenderParams,
				Platform:      script.Platform,
				ScriptType:    script.ScriptType,
				ScriptTimeout: script.ScriptTimeout,
				EncodingType:  script.EncodingType,
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
	GetCmd.AddCommand(vmScriptsCmd)
	common.AddQueryFlags(vmScriptsCmd)
	vmScriptsCmd.Flags().Bool("show-full-script", false, "Show full script content instead of truncated version")
}
