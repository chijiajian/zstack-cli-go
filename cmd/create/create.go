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
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		if fileFlag != "" {
			return fmt.Errorf("create --file is not supported on the parent command; use 'zstack-cli create <resource> --file %s'", fileFlag)
		}
		if len(args) == 0 && fileFlag == "" {
			cmd.Help()
			return nil
		}
		return nil
	},
}

func init() {

	CreateCmd.PersistentFlags().StringVarP(&fileFlag, "file", "f", "", "Filename, directory, or URL to files containing resource definitions")
	CreateCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "Output format (json|yaml)")
	CreateCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Only print the object that would be sent, without sending it")
	CreateCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
}

func exactOneArgUnlessFile(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		return cobra.MaximumNArgs(1)(cmd, args)
	}
	return cobra.ExactArgs(1)(cmd, args)
}

func stringPtrIfNotEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func intPtrFromInt64IfNotZero(value int64) *int {
	if value == 0 {
		return nil
	}
	intValue := int(value)
	return &intValue
}

func int64PtrIfNotZero(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

// Keep machine-readable output a single document; human progress must also
// propagate writer errors instead of hiding a partially successful operation.
func printCreateProgress(cmd *cobra.Command, message string, args ...any) error {
	format, _ := cmd.Flags().GetString("output")
	if format == "json" || format == "yaml" || format == "name" {
		return nil
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), message, args...)
	return clierr.Wrap(clierr.Output, err)
}
