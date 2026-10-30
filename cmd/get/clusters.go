/*
Copyright © 2025 zstack.io
*/
package get

import (
	"fmt"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
)

type FormattedCluster struct {
	Name           string `json:"name" yaml:"name" header:"NAME"`
	UUID           string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description    string `json:"description" yaml:"description" header:"DESCRIPTION"`
	State          string `json:"state" yaml:"state" header:"STATE"`
	HypervisorType string `json:"hypervisorType" yaml:"hypervisorType" header:"HYPERVISOR TYPE"`
	ZoneUuid       string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	Type           string `json:"type" yaml:"type" header:"TYPE"`
	Architecture   string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
}

var clustersCmd = &cobra.Command{
	Use:   "clusters [name]",
	Short: "Get ZStack clusters",
	Long: `Display one or many ZStack clusters.

Examples:
  # List all clusters
  zstack-cli get clusters

  # Get a specific cluster by name
  zstack-cli get clusters my-cluster

  # Query clusters with specific conditions
  zstack-cli get clusters --q "name=Cluster1"

  # Query clusters with multiple conditions
  zstack-cli get clusters --q "name=Cluster1" --q "state=Enabled"

  # Limit the number of results and paginate
  zstack-cli get clusters --limit 10 --start 0

  # Output in different formats
  zstack-cli get clusters --output json
  zstack-cli get clusters --output yaml
  zstack-cli get clusters --output text`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.ClusterInventoryView](cobraCmd, queryParam, "v1/clusters")
		clusters := result.Items
		if err != nil {
			return fmt.Errorf("query clusters: %w", err)
		}

		var formattedResults []FormattedCluster
		for _, cluster := range clusters {
			formattedResults = append(formattedResults, FormattedCluster{
				Name:           cluster.Name,
				UUID:           cluster.UUID,
				Description:    cluster.Description,
				State:          cluster.State,
				HypervisorType: cluster.HypervisorType,
				ZoneUuid:       cluster.ZoneUuid,
				Type:           cluster.Type,
				Architecture:   cluster.Architecture,
			})
		}

		err = common.PrintQuery(cobraCmd, formattedResults, result.Total)
		if err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
	},
}

func init() {
	GetCmd.AddCommand(clustersCmd)

	common.AddQueryFlags(clustersCmd)
}
