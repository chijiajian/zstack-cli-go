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

type FormattedVmScriptRecord struct {
	UUID           string    `json:"uuid" yaml:"uuid" header:"UUID"`
	Name           string    `json:"name" yaml:"name" header:"NAME"`
	ScriptUUID     string    `json:"scriptUuid" yaml:"scriptUuid" header:"SCRIPT UUID"`
	RecordName     string    `json:"recordName" yaml:"recordName" header:"RECORD NAME"`
	Status         string    `json:"status" yaml:"status" header:"STATUS"`
	Executor       string    `json:"executor" yaml:"executor" header:"EXECUTOR"`
	ExecutionCount int       `json:"executionCount" yaml:"executionCount" header:"EXECUTION COUNT"`
	Version        int       `json:"version" yaml:"version" header:"VERSION"`
	EncodingType   string    `json:"encodingType" yaml:"encodingType" header:"ENCODING TYPE"`
	ScriptTimeout  int       `json:"scriptTimeout" yaml:"scriptTimeout" header:"SCRIPT TIMEOUT (SEC)"`
	StartTime      time.Time `json:"startTime" yaml:"startTime" header:"START TIME"`
	EndTime        time.Time `json:"endTime" yaml:"endTime" header:"END TIME"`
}

var vmScriptRecordsCmd = &cobra.Command{
	Use:   "vm-script-records [record-name]",
	Short: "List VM script execution records",
	Long:  `List VM script execution records in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "recordName")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.GuestVmScriptExecutedRecordInventoryView](cobraCmd, queryParam, "v1/scripts/records")
		records := result.Items
		if err != nil {
			return fmt.Errorf("Error querying VM script execution records: %w", err)
		}

		formattedResults := make([]FormattedVmScriptRecord, 0, len(records))
		for _, record := range records {
			formattedResults = append(formattedResults, FormattedVmScriptRecord{
				UUID:           record.UUID,
				Name:           record.Name,
				ScriptUUID:     record.ScriptUuid,
				RecordName:     record.RecordName,
				Status:         record.Status,
				Executor:       record.Executor,
				ExecutionCount: record.ExecutionCount,
				Version:        record.Version,
				EncodingType:   record.EncodingType,
				ScriptTimeout:  record.ScriptTimeout,
				StartTime:      record.StartTime,
				EndTime:        record.EndTime,
			})
		}

		if err := common.PrintQuery(cobraCmd, formattedResults, result.Total); err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
	},
}

func init() {
	GetCmd.AddCommand(vmScriptRecordsCmd)
	common.AddQueryFlags(vmScriptRecordsCmd)
}
