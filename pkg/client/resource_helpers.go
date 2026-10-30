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

package client

import (
	"fmt"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/types"
	sdkClient "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

func GetZoneUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	resource, err := ResolveOneResource(cli, "zone", nameOrUUID)
	if err != nil {
		return "", err
	}
	return resource.UUID, nil
}

func GetClusterUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	resource, err := ResolveOneResource(cli, "cluster", nameOrUUID)
	if err != nil {
		return "", err
	}
	return resource.UUID, nil
}

func GetHostUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	resource, err := ResolveOneResource(cli, "host", nameOrUUID)
	if err != nil {
		return "", err
	}
	return resource.UUID, nil
}

func GetBackupStorageUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	resource, err := ResolveOneResource(cli, "backup-storage", nameOrUUID)
	if err != nil {
		return "", err
	}
	return resource.UUID, nil
}

func GetImageUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	resource, err := ResolveOneResource(cli, "image", nameOrUUID)
	if err != nil {
		return "", err
	}
	if resource.State != types.ImageStateEnabled || resource.Status != types.ImageStatusReady {
		return "", fmt.Errorf("image %q must be Enabled and Ready", nameOrUUID)
	}
	return resource.UUID, nil
}

func GetInstanceOfferingUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	resource, err := ResolveOneResource(cli, "instance-offering", nameOrUUID)
	if err != nil {
		return "", err
	}
	return resource.UUID, nil
}

func GetL3NetworkUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	resource, err := ResolveOneResource(cli, "l3-network", nameOrUUID)
	if err != nil {
		return "", err
	}
	return resource.UUID, nil
}

func GetPrimaryStorageUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {
	resource, err := ResolveOneResource(cli, "primary-storage", nameOrUUID)
	if err != nil {
		return "", err
	}
	return resource.UUID, nil
}

// GetReadyImagesByNameOrUUID
func GetReadyImagesByNameOrUUID(cli *sdkClient.ZSClient, nameOrUUID string) ([]view.ImageInventoryView, error) {
	queryParam := param.NewQueryParam()

	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))
	images, err := cli.QueryImage(&queryParam)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		queryParam = param.NewQueryParam()
		queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))
		images, err = cli.QueryImage(&queryParam)
		if err != nil {
			return nil, err
		}
	}

	readyImages := []view.ImageInventoryView{}
	for _, img := range images {
		if strings.TrimSpace(img.UUID) == "" {
			return nil, fmt.Errorf("image query returned a resource without UUID")
		}
		if types.IsImageReady(img.Status) {
			readyImages = append(readyImages, img)
		}
	}

	return readyImages, nil
}

// GetReadyImagesByNameOrUUID
func GetDeletedImagesByNameOrUUID(cli *sdkClient.ZSClient, nameOrUUID string) ([]view.ImageInventoryView, error) {
	queryParam := param.NewQueryParam()

	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))
	images, err := cli.QueryImage(&queryParam)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		queryParam = param.NewQueryParam()
		queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))
		images, err = cli.QueryImage(&queryParam)
		if err != nil {
			return nil, err
		}
	}

	deletedImages := []view.ImageInventoryView{}
	for _, img := range images {
		if strings.TrimSpace(img.UUID) == "" {
			return nil, fmt.Errorf("image query returned a resource without UUID")
		}
		if img.Status == types.ImageStatusDeleted {
			deletedImages = append(deletedImages, img)
		}
	}

	return deletedImages, nil
}

// GetReadyVMsByNameOrUUID
func GetReadyVMsByNameOrUUID(cli *sdkClient.ZSClient, nameOrUUID string) ([]view.VmInstanceInventoryView, error) {
	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	vms, err := cli.QueryVmInstance(&queryParam)
	if err != nil {
		return nil, err
	}

	if len(vms) == 0 {
		queryParam = param.NewQueryParam()
		queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))
		vms, err = cli.QueryVmInstance(&queryParam)
		if err != nil {
			return nil, err
		}
	}

	activeVMs := []view.VmInstanceInventoryView{}
	for _, vm := range vms {
		if strings.TrimSpace(vm.UUID) == "" {
			return nil, fmt.Errorf("instance query returned a resource without UUID")
		}
		if types.IsVMActive(vm.State) {
			activeVMs = append(activeVMs, vm)
		}
	}

	return activeVMs, nil
}

// GetReadyVMsByNameOrUUID
func GetDestroyedVMsByNameOrUUID(cli *sdkClient.ZSClient, nameOrUUID string) ([]view.VmInstanceInventoryView, error) {
	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	vms, err := cli.QueryVmInstance(&queryParam)
	if err != nil {
		return nil, err
	}

	if len(vms) == 0 {
		queryParam = param.NewQueryParam()
		queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))
		vms, err = cli.QueryVmInstance(&queryParam)
		if err != nil {
			return nil, err
		}
	}

	destroyedVMs := []view.VmInstanceInventoryView{}
	for _, vm := range vms {
		if strings.TrimSpace(vm.UUID) == "" {
			return nil, fmt.Errorf("instance query returned a resource without UUID")
		}
		if vm.State == types.VMStateDestroyed {
			destroyedVMs = append(destroyedVMs, vm)
		}
	}

	return destroyedVMs, nil
}
