package client

import (
	"context"

	sdk "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

// UpdateVMInstance sends the typed request envelope. The SDK's generated
// UpdateVmInstance wraps its detail in map[string]interface{}, which disables
// omitempty for nested fields in the SDK serializer.
func UpdateVMInstance(ctx context.Context, cli *sdk.ZSClient, uuid string, params param.UpdateVmInstanceParam) (*view.VmInstanceInventoryView, error) {
	var result view.VmInstanceInventoryView
	if err := cli.ZSHttpClient.PutWithRespKey(ctx, "v1/vm-instances", uuid, "", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
