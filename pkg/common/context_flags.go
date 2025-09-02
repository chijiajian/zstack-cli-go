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
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
)

func AddBasicContextFlags(cmd *cobra.Command) {
	cmd.Flags().String("zone", "", "Filter resources by zone name or UUID")
	cmd.Flags().String("cluster", "", "Filter resources by cluster name or UUID")
	cmd.Flags().String("host", "", "Filter resources by host name or UUID")
}

func ProcessBasicContextFlags(cmd *cobra.Command, queryParam *param.QueryParam) error {

	zsClient := client.GetClient()
	if zsClient == nil {
		return fmt.Errorf("not logged in, please run 'zstack-cli login' first")
	}

	zone, _ := cmd.Flags().GetString("zone")
	if zone != "" {

		zoneUUID, err := client.GetZoneUUIDByName(zsClient, zone)
		if err != nil {
			return fmt.Errorf("failed to find zone '%s': %v", zone, err)
		}
		queryParam.AddQ(fmt.Sprintf("zoneUuid=%s", zoneUUID))
	}

	cluster, _ := cmd.Flags().GetString("cluster")
	if cluster != "" {

		clusterUUID, err := client.GetClusterUUIDByName(zsClient, cluster)
		if err != nil {
			return fmt.Errorf("failed to find cluster '%s': %v", cluster, err)
		}
		queryParam.AddQ(fmt.Sprintf("clusterUuid=%s", clusterUUID))
	}

	host, _ := cmd.Flags().GetString("host")
	if host != "" {

		hostUUID, err := client.GetHostUUIDByName(zsClient, host)
		if err != nil {
			return fmt.Errorf("failed to find host '%s': %v", host, err)
		}
		queryParam.AddQ(fmt.Sprintf("hostUuid=%s", hostUUID))
	}

	return nil
}
