// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// CreateCmd
var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create one or many resources",
	Long:  `Create one or many ZStack resources.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 && fileFlag == "" {
			cmd.Help()
			return
		}
	},
	PersistentPreRunE: preRunCheckFile,
}

func init() {

	CreateCmd.PersistentFlags().StringVarP(&fileFlag, "file", "f", "", "Filename, directory, or URL to files containing resource definitions")
	CreateCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "Output format (json|yaml)")
	CreateCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Only print the object that would be sent, without sending it")
	CreateCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
}

func preRunCheckFile(cmd *cobra.Command, args []string) error {

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

		cmd.Run = func(cmd *cobra.Command, args []string) {

		}
	}
	return nil
}
