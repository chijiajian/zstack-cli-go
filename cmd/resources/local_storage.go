package resources

import (
	"fmt"
	"path"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

var primaryStorageAddLocalCmd = &cobra.Command{
	Use: "add-local <name>", Short: "Add a local primary storage", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		request, err := buildAddLocalPrimaryStorageParam(cmd, args[0])
		if err != nil {
			return err
		}
		if isInfrastructureDryRun(cmd) {
			return printInfrastructureDryRun(cmd, request)
		}
		cli, err := client.GetClientContext(cmd.Context())
		if err != nil {
			return err
		}
		zone, err := client.ResolveOneResource(cli, "zone", request.Params.ZoneUuid)
		if err != nil {
			return err
		}
		request.Params.ZoneUuid = zone.UUID
		var result view.PrimaryStorageInventoryView
		// SDK v0.0.10's convenience method sends an empty body. Use its typed
		// request and shared transport until the method accepts parameters.
		if err := cli.Post("v1/primary-storage/local-storage", request, &result); err != nil {
			return err
		}
		return printInfrastructureResult(cmd, fmt.Sprintf("primary-storage/%s added", result.Name), result)
	},
}

func buildAddLocalPrimaryStorageParam(cmd *cobra.Command, name string) (param.AddLocalPrimaryStorageParam, error) {
	base, err := buildAddNFSPrimaryStorageParam(cmd, name)
	if err != nil {
		return param.AddLocalPrimaryStorageParam{}, err
	}
	if strings.TrimSpace(name) == "" || strings.TrimSpace(base.Params.ZoneUuid) == "" {
		return param.AddLocalPrimaryStorageParam{}, fmt.Errorf("name and zone must be nonempty")
	}
	if !path.IsAbs(base.Params.Url) {
		return param.AddLocalPrimaryStorageParam{}, fmt.Errorf("local storage --url must be an absolute host directory")
	}
	p := base.Params
	return param.AddLocalPrimaryStorageParam{BaseParam: base.BaseParam, Params: param.AddLocalPrimaryStorageParamDetail{
		Name: p.Name, Url: p.Url, ZoneUuid: p.ZoneUuid, Description: p.Description, Type: p.Type, ResourceUuid: p.ResourceUuid, TagUuids: p.TagUuids,
	}}, nil
}

func addPrimaryStorageFlags(cmd *cobra.Command) {
	cmd.Flags().String("url", "", "Storage URL or absolute local host directory")
	cmd.Flags().String("zone", "", "Zone name or UUID")
	cmd.Flags().StringP("description", "d", "", "Detailed description of the primary storage")
	cmd.Flags().String("type", "", "Primary storage type")
	cmd.Flags().String("resource-uuid", "", "Resource UUID")
	cmd.Flags().StringSlice("tag", nil, "Tag UUID list")
	cmd.Flags().StringSlice("system-tag", nil, "System tag list")
	cmd.Flags().StringSlice("user-tag", nil, "User tag list")
}
