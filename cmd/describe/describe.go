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

package describe

import (
	"context"
	"fmt"
	"net/url"
	"time"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

var DescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a resource",
	Long:  "Describe detailed information for a ZStack resource.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return nil
	},
}

var longJobCmd = &cobra.Command{
	Use:     "long-job <uuid>",
	Aliases: []string{"job", "longjob"},
	Short:   "Describe a long-running async job",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("request-timeout")
		if timeout <= 0 {
			return fmt.Errorf("request-timeout must be positive")
		}
		output, _ := cmd.Flags().GetString("output")
		switch output {
		case "table", "text", "json", "yaml":
		default:
			return fmt.Errorf("unsupported output format %q", output)
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
		defer cancel()
		cli, err := zstackclient.NewReadClient(ctx)
		if err != nil {
			return err
		}
		var response struct {
			Inventories []view.LongJobInventoryView `json:"inventories"`
		}
		if err := cli.Read(ctx, "v1/longjobs/"+url.PathEscape(args[0]), nil, &response); err != nil {
			return err
		}
		if len(response.Inventories) != 1 {
			return fmt.Errorf("expected one long-job %s, got %d", args[0], len(response.Inventories))
		}
		job := response.Inventories[0]
		if job.UUID == "" || job.UUID != args[0] {
			return fmt.Errorf("long-job identity mismatch: expected %q, got %q", args[0], job.UUID)
		}

		fields, _ := cmd.Flags().GetStringSlice("fields")
		if len(fields) > 0 {
			projected, err := utils.ProjectObject(job, fields)
			if err != nil {
				return err
			}
			return utils.Print(projected, utils.ParseFormat(output))
		}
		return utils.Print(job, utils.ParseFormat(output))
	},
}

func init() {
	DescribeCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, yaml, or text")
	DescribeCmd.PersistentFlags().StringSlice("fields", nil, "Fields to display")
	longJobCmd.Flags().Duration("request-timeout", 30*time.Second, "Deadline for authentication and job detail")
	DescribeCmd.AddCommand(longJobCmd)
}
