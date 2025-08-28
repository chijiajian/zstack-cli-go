// cmd/create/create.go
package create

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	fileFlag    string
	outputFlag  string
	dryRunFlag  bool
	verboseFlag bool
)

// CreateCmd 表示 create 命令
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create one or many resources",
	Long:  `Create one or many ZStack resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有子命令或文件标志，显示帮助
		if len(args) == 0 && fileFlag == "" {
			cmd.Help()
			return
		}
	},
	PersistentPreRunE: preRunCheckFile,
}

// 初始化命令
func init() {
	// 添加通用标志
	CreateCmd.PersistentFlags().StringVarP(&fileFlag, "file", "f", "", "Filename, directory, or URL to files containing resource definitions")
	CreateCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "Output format (json|yaml)")
	CreateCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Only print the object that would be sent, without sending it")
	CreateCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
}

// 预运行函数，检查文件标志
func preRunCheckFile(cmd *cobra.Command, args []string) error {
	// 如果指定了文件标志，处理文件
	if fileFlag != "" {
		if dryRunFlag {
			fmt.Println("Dry run mode: would process file", fileFlag)
			return nil
		}

		if verboseFlag {
			fmt.Printf("Processing file: %s\n", fileFlag)
		}

		err := utils.ProcessFilePath(fileFlag)
		if err != nil {
			return err
		}

		// 如果成功处理了文件，可以选择跳过常规命令执行
		cmd.Run = func(cmd *cobra.Command, args []string) {
			// 文件已处理，无需进一步操作
		}
	}
	return nil
}
