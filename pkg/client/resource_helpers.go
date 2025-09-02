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

	sdkClient "github.com/terraform-zstack-modules/zstack-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/view"
)

// GetZoneUUIDByName
func GetZoneUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {

	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))

	zones, err := cli.QueryZone(queryParam)
	if err != nil {
		return "", err
	}

	if len(zones) > 0 {
		return zones[0].UUID, nil
	}

	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	zones, err = cli.QueryZone(queryParam)
	if err != nil {
		return "", err
	}

	if len(zones) > 0 {
		return zones[0].UUID, nil
	}

	return "", fmt.Errorf("zone with name or UUID '%s' not found", nameOrUUID)
}

// GetClusterUUIDByName
func GetClusterUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {

	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))

	clusters, err := cli.QueryCluster(queryParam)
	if err != nil {
		return "", err
	}

	if len(clusters) > 0 {
		return clusters[0].Uuid, nil
	}

	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	clusters, err = cli.QueryCluster(queryParam)
	if err != nil {
		return "", err
	}

	if len(clusters) > 0 {
		return clusters[0].Uuid, nil
	}

	return "", fmt.Errorf("cluster with name or UUID '%s' not found", nameOrUUID)
}

// GetHostUUIDByName
func GetHostUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {

	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))

	hosts, err := cli.QueryHost(queryParam)
	if err != nil {
		return "", err
	}

	if len(hosts) > 0 {
		return hosts[0].UUID, nil
	}

	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	hosts, err = cli.QueryHost(queryParam)
	if err != nil {
		return "", err
	}

	if len(hosts) > 0 {
		return hosts[0].UUID, nil
	}

	return "", fmt.Errorf("host with name or UUID '%s' not found", nameOrUUID)
}

// GetBackupStorageUUIDByName
func GetBackupStorageUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {

	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name=%s", nameOrUUID))

	backupStorages, err := cli.QueryBackupStorage(queryParam)
	if err != nil {
		return "", err
	}

	if len(backupStorages) > 0 {
		return backupStorages[0].UUID, nil
	}

	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	backupStorages, err = cli.QueryBackupStorage(queryParam)
	if err != nil {
		return "", err
	}

	if len(backupStorages) > 0 {
		return backupStorages[0].UUID, nil
	}

	return "", fmt.Errorf("backup storage with name or UUID '%s' not found", nameOrUUID)
}

// GetImageUUIDByName
func GetImageUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {

	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name~=%s", nameOrUUID))

	images, err := cli.QueryImage(queryParam)
	if err != nil {
		return "", err
	}

	if len(images) > 0 {
		return images[0].UUID, nil
	}

	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid~=%s", nameOrUUID))

	images, err = cli.QueryImage(queryParam)
	if err != nil {
		return "", err
	}

	if len(images) > 0 {
		return images[0].UUID, nil
	}

	return "", fmt.Errorf("image with name or UUID '%s' not found", nameOrUUID)
}

// GetInstanceUUIDByName
func GetInstanceUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {

	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name~=%s", nameOrUUID))

	instances, err := cli.QueryInstaceOffering(queryParam)
	if err != nil {
		return "", err
	}

	if len(instances) > 0 {
		return instances[0].UUID, nil
	}

	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	instances, err = cli.QueryInstaceOffering(queryParam)
	if err != nil {
		return "", err
	}

	if len(instances) > 0 {
		return instances[0].UUID, nil
	}

	return "", fmt.Errorf("instance with name or UUID '%s' not found", nameOrUUID)
}

// GetL3NetworkUUIDByName
func GetL3NetworkUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {

	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name~=%s", nameOrUUID))

	l3Networks, err := cli.QueryL3Network(queryParam)
	if err != nil {
		return "", err
	}

	if len(l3Networks) > 0 {
		return l3Networks[0].UUID, nil
	}

	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	l3Networks, err = cli.QueryL3Network(queryParam)
	if err != nil {
		return "", err
	}

	if len(l3Networks) > 0 {
		return l3Networks[0].UUID, nil
	}

	return "", fmt.Errorf("L3 network with name or UUID '%s' not found", nameOrUUID)
}

// GetPrimaryStorageUUIDByName
func GetPrimaryStorageUUIDByName(cli *sdkClient.ZSClient, nameOrUUID string) (string, error) {

	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name~=%s", nameOrUUID))

	primaryStorages, err := cli.QueryPrimaryStorage(queryParam)
	if err != nil {
		return "", err
	}

	if len(primaryStorages) > 0 {
		return primaryStorages[0].UUID, nil
	}

	queryParam = param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))

	primaryStorages, err = cli.QueryPrimaryStorage(queryParam)
	if err != nil {
		return "", err
	}

	if len(primaryStorages) > 0 {
		return primaryStorages[0].UUID, nil
	}

	return "", fmt.Errorf("primary storage with name or UUID '%s' not found", nameOrUUID)
}

// GetReadyImagesByNameOrUUID
func GetReadyImagesByNameOrUUID(cli *sdkClient.ZSClient, nameOrUUID string) ([]view.ImageView, error) {
	queryParam := param.NewQueryParam()

	queryParam.AddQ(fmt.Sprintf("name~=%s", nameOrUUID))
	images, err := cli.QueryImage(queryParam)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		queryParam = param.NewQueryParam()
		queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))
		images, err = cli.QueryImage(queryParam)
		if err != nil {
			return nil, err
		}
	}

	readyImages := []view.ImageView{}
	for _, img := range images {
		if img.Status == "Ready" {
			readyImages = append(readyImages, img)
		}
	}

	return readyImages, nil
}

// GetReadyImagesByNameOrUUID
func GetDeletedImagesByNameOrUUID(cli *sdkClient.ZSClient, nameOrUUID string) ([]view.ImageView, error) {
	queryParam := param.NewQueryParam()

	queryParam.AddQ(fmt.Sprintf("name~=%s", nameOrUUID))
	images, err := cli.QueryImage(queryParam)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		queryParam = param.NewQueryParam()
		queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))
		images, err = cli.QueryImage(queryParam)
		if err != nil {
			return nil, err
		}
	}

	readyImages := []view.ImageView{}
	for _, img := range images {
		if img.Status == "Deleted" {
			readyImages = append(readyImages, img)
		}
	}

	return readyImages, nil
}

// GetReadyVMsByNameOrUUID
func GetReadyVMsByNameOrUUID(cli *sdkClient.ZSClient, nameOrUUID string) ([]view.VmInstanceInventoryView, error) {
	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name~=%s", nameOrUUID))

	vms, err := cli.QueryVmInstance(queryParam)
	if err != nil {
		return nil, err
	}

	if len(vms) == 0 {
		queryParam = param.NewQueryParam()
		queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))
		vms, err = cli.QueryVmInstance(queryParam)
		if err != nil {
			return nil, err
		}
	}

	readyVMs := []view.VmInstanceInventoryView{}
	for _, vm := range vms {
		if vm.State == "Running" || vm.State == "Stopped" {
			readyVMs = append(readyVMs, vm)
		}
	}

	return readyVMs, nil
}

// GetReadyVMsByNameOrUUID
func GetDestroyedVMsByNameOrUUID(cli *sdkClient.ZSClient, nameOrUUID string) ([]view.VmInstanceInventoryView, error) {
	queryParam := param.NewQueryParam()
	queryParam.AddQ(fmt.Sprintf("name~=%s", nameOrUUID))

	vms, err := cli.QueryVmInstance(queryParam)
	if err != nil {
		return nil, err
	}

	if len(vms) == 0 {
		queryParam = param.NewQueryParam()
		queryParam.AddQ(fmt.Sprintf("uuid=%s", nameOrUUID))
		vms, err = cli.QueryVmInstance(queryParam)
		if err != nil {
			return nil, err
		}
	}

	readyVMs := []view.VmInstanceInventoryView{}
	for _, vm := range vms {
		if vm.State == "Destroyed" {
			readyVMs = append(readyVMs, vm)
		}
	}

	return readyVMs, nil
}
