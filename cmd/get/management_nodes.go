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
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
)

var managementNodesCmd = &cobra.Command{
	Use:   "management-nodes [hostname]",
	Short: "Query management nodes",
	Long:  `Query management nodes in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "hostName")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.ManagementNodeInventoryView](cobraCmd, queryParam, "v1/management-nodes")
		nodes := result.Items
		if err != nil {
			return fmt.Errorf("Error querying management nodes: %w", err)
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

		err = common.PrintQuery(cobraCmd, formattedResults, result.Total)
		if err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
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
