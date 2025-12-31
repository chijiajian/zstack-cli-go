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

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

type FormattedLongJob struct {
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Name        string `json:"name" yaml:"name" header:"NAME"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
	State       string `json:"state" yaml:"state" header:"STATE"`
	JobName     string `json:"jobName" yaml:"jobName" header:"JOB NAME"`
	TargetName  string `json:"targetResourceUuid" yaml:"targetResourceUuid" header:"TARGET"`
	ExecuteTime int64  `json:"executeTime" yaml:"executeTime" header:"EXECUTE TIME"`
	CreateDate  string `json:"createDate" yaml:"createDate" header:"CREATED"`
}

var longJobsCmd = &cobra.Command{
	Use:     "long-jobs [name]",
	Aliases: []string{"long-job", "jobs", "job"},
	Short:   "List long-running async jobs",
	Long:    `List all long-running async jobs in the ZStack cloud platform.`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cobraCmd *cobra.Command, args []string) {

		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			fmt.Printf("Error building query parameters: %s\n", err)
			return
		}

		usePagination, _ := cobraCmd.Flags().GetBool("pagination")
		var jobs []view.LongJobInventoryView
		var total int

		if usePagination {
			jobs, total, err = zsClient.PageLongJob(*queryParam)
			if err != nil {
				fmt.Printf("Error querying long jobs: %s\n", err)
				return
			}
			fmt.Printf("Total: %d\n", total)
		} else {
			jobs, err = zsClient.QueryLongJob(*queryParam)
			if err != nil {
				fmt.Printf("Error querying long jobs: %s\n", err)
				return
			}
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedLongJob
		for _, job := range jobs {
			formatted := FormattedLongJob{
				UUID:        job.UUID,
				Name:        job.Name,
				Description: job.Description,
				State:       string(job.State),
				JobName:     job.JobName,
				TargetName:  job.TargetResourceUuid,
				ExecuteTime: job.ExecuteTime,
				CreateDate:  job.CreateDate.Format("2006-01-02 15:04:05"),
			}
			formattedResults = append(formattedResults, formatted)
		}

		err = utils.PrintWithFields(formattedResults, format, fields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
			return
		}
	},
}

func init() {
	GetCmd.AddCommand(longJobsCmd)
	common.AddQueryFlags(longJobsCmd)
	longJobsCmd.Flags().Bool("pagination", false, "Use pagination when querying long jobs")
}
