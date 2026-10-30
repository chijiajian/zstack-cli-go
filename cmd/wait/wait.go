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

package wait

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/waiter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var WaitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait for a resource to reach a state",
	Long:  "Wait for a ZStack resource to reach the requested state or status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return nil
	},
}

func init() {
	WaitCmd.AddCommand(
		newWaitResourceCmd("instance", nil, "Wait for a VM instance"),
		newWaitResourceCmd("image", nil, "Wait for an image"),
		newWaitResourceCmd("volume", nil, "Wait for a volume"),
		newWaitResourceCmd("long-job", []string{"job", "longjob"}, "Wait for a long-running async job"),
	)
}

func newWaitResourceCmd(resourceType string, aliases []string, short string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     resourceType + " <name-or-uuid>",
		Aliases: aliases,
		Short:   short,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWait(cmd, resourceType, args[0])
		},
	}
	defaultTarget := ""
	if resourceType == "long-job" {
		defaultTarget = "Succeeded"
	}
	cmd.Flags().String("for", defaultTarget, "Required state or status")
	cmd.Flags().StringP("output", "o", "text", "Output format: text, table, json, or yaml")
	cmd.Flags().Bool("progress", false, "Print state changes to stderr")
	cmd.Flags().Duration("timeout", 10*time.Minute, "Maximum time to wait")
	cmd.Flags().Duration("interval", 5*time.Second, "Polling interval")
	if resourceType != "long-job" {
		_ = cmd.MarkFlagRequired("for")
	}
	return cmd
}

func runWait(cmd *cobra.Command, resourceType string, identifier string) error {
	target, _ := cmd.Flags().GetString("for")
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("required flag --for not set")
	}
	timeout, _ := cmd.Flags().GetDuration("timeout")
	interval, _ := cmd.Flags().GetDuration("interval")
	if timeout <= 0 {
		return fmt.Errorf("--timeout must be greater than 0")
	}
	if interval <= 0 {
		return fmt.Errorf("--interval must be greater than 0")
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "text", "table", "json", "yaml":
	default:
		return fmt.Errorf("unsupported output format %q", output)
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()
	cli, err := zstackclient.NewReadClient(ctx)
	if err != nil {
		return err
	}
	resolved, err := cli.Resolve(ctx, resourceType, identifier, zstackclient.ResolveAuto)
	if err != nil {
		return err
	}
	convert := func(r zstackclient.ResolvedResource) waiter.Observation {
		return waiter.Observation{ResourceType: r.ResourceType, UUID: r.UUID, Name: r.Name, State: r.State, Status: r.Status, JobResult: r.JobResult}
	}
	evaluate := waiter.Match
	if resourceType == "long-job" {
		evaluate = waiter.Job
	}
	progress, _ := cmd.Flags().GetBool("progress")
	var notify func(waiter.Observation) error
	if progress {
		notify = func(o waiter.Observation) error {
			_, err := fmt.Fprintf(cmd.ErrOrStderr(), "%s/%s: %s\n", o.ResourceType, o.UUID, o.Summary())
			if err != nil {
				return fmt.Errorf("write wait progress: %w", err)
			}
			return nil
		}
	}
	result, err := waiter.Run(ctx, convert(resolved), target, interval, func(ctx context.Context) (waiter.Observation, error) {
		r, e := cli.Resolve(ctx, resourceType, resolved.UUID, zstackclient.ResolveUUID)
		return convert(r), e
	}, evaluate, notify)
	if err != nil {
		return err
	}
	switch output {
	case "json":
		return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
	case "yaml":
		return yaml.NewEncoder(cmd.OutOrStdout()).Encode(result)
	default:
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s/%s reached %q (%s)\n", resourceType, result.Name, strings.TrimSpace(target), result.Summary())
		return err
	}
}
