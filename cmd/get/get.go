// cmd/get/get.go
package get

import (
	"github.com/spf13/cobra"
)

// GetCmd 表示 get 命令
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Display one or many resources",
	Long:  `Display one or many ZStack resources.`,
}
