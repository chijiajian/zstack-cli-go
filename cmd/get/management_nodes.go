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

package get

import (
	"fmt"
	"strings"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

var managementNodesCmd = &cobra.Command{
	Use:   "management-nodes [hostname]",
	Short: "Query management nodes",
	Long:  `Query management nodes in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cobraCmd *cobra.Command, args []string) {

		fmt.Printf("Debug: Command arguments: %v\n", args)

		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "hostName")
		if err != nil {
			fmt.Printf("Error building query parameters: %s\n", err)
			return
		}

		fmt.Printf("Debug: Query parameters: %+v\n", queryParam)

		nodes, err := zsClient.QueryManagementNode(*queryParam)
		if err != nil {
			fmt.Printf("Error querying management nodes: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")
		processedFields := []string{}

		for _, field := range fields {
			parts := strings.Split(field, ",")
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					processedFields = append(processedFields, trimmed)
				}
			}
		}

		if len(processedFields) > 0 {
			fmt.Printf("Debug: Processed fields: %v\n", processedFields)
		}

		var formattedResults []FormattedManagementNode
		for _, node := range nodes {
			formatted := FormattedManagementNode{
				UUID:      node.UUID,
				HostName:  node.HostName,
				JoinDate:  node.JoinDate.Format(time.RFC3339),
				HeartBeat: node.HeartBeat.Format(time.RFC3339),
			}
			formattedResults = append(formattedResults, formatted)
		}

		err = utils.PrintWithFields(formattedResults, format, processedFields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
		}
	},
}

type FormattedManagementNode struct {
	UUID      string `json:"uuid"`
	HostName  string `json:"hostName"`
	JoinDate  string `json:"joinDate"`
	HeartBeat string `json:"heartBeat"`
}

func init() {
	GetCmd.AddCommand(managementNodesCmd)
	common.AddQueryFlags(managementNodesCmd)

}
