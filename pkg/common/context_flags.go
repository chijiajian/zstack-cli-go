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

package common

import (
	"context"
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

func AddBasicContextFlags(cmd *cobra.Command) {
	cmd.Flags().String("zone", "", "Filter resources by zone name or UUID")
	cmd.Flags().String("cluster", "", "Filter resources by cluster name or UUID")
	cmd.Flags().String("host", "", "Filter resources by host name or UUID")
}

func ProcessBasicContextFlags(cmd *cobra.Command, queryParam *param.QueryParam) error {
	for _, filter := range []string{"zone", "cluster", "host"} {
		if err := ResolveFilter(cmd, queryParam, filter, filter, filter+"Uuid"); err != nil {
			return err
		}
	}
	return nil
}

func ResolveFilter(cmd *cobra.Command, params *param.QueryParam, flag, resource, field string) error {
	identifier, _ := cmd.Flags().GetString(flag)
	if identifier == "" {
		return nil
	}
	timeout, _ := cmd.Flags().GetDuration("request-timeout")
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()
	reader, err := client.NewReadClient(ctx)
	if err != nil {
		return err
	}
	resolved, err := reader.Resolve(ctx, resource, identifier, client.ResolveAuto)
	if err != nil {
		return fmt.Errorf("resolve --%s: %w", flag, err)
	}
	params.AddQ(field + "=" + resolved.UUID)
	return nil
}
