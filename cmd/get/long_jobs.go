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

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
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
	JobResult   string `json:"jobResult,omitempty" yaml:"jobResult,omitempty" header:"JOB RESULT"`
}

var longJobsCmd = &cobra.Command{
	Use:     "long-jobs [name]",
	Aliases: []string{"long-job", "jobs", "job"},
	Short:   "List long-running async jobs",
	Long:    `List all long-running async jobs in the ZStack cloud platform.`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.LongJobInventoryView](cobraCmd, queryParam, "v1/longjobs")
		jobs := result.Items
		if err != nil {
			return err
		}

		var formattedResults []FormattedLongJob
		for _, job := range jobs {
			formatted := FormattedLongJob{
				UUID:        job.UUID,
				JobResult:   job.JobResult,
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

		err = common.PrintQuery(cobraCmd, formattedResults, result.Total)
		if err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
	},
}

func init() {
	GetCmd.AddCommand(longJobsCmd)
	common.AddQueryFlags(longJobsCmd)
	longJobsCmd.Flags().Bool("pagination", false, "Use pagination when querying long jobs")
}
