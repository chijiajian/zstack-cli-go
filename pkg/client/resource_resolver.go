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
	"context"
	"fmt"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"strings"

	sdkClient "github.com/zstackio/zstack-sdk-go-v2/pkg/client"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

type ResolveStrategy string

const (
	ResolveAuto      ResolveStrategy = "auto"
	ResolveUUID      ResolveStrategy = "uuid"
	ResolveExactName ResolveStrategy = "exact-name"
	ResolveFuzzyName ResolveStrategy = "fuzzy-name"
)

type ResolveOptions struct {
	Strategy      ResolveStrategy
	AllowMultiple bool
}

type ResolvedResource struct {
	JobResult    string `json:"jobResult,omitempty" yaml:"jobResult,omitempty"`
	ResourceType string `json:"resourceType" yaml:"resourceType"`
	UUID         string `json:"uuid" yaml:"uuid"`
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	State        string `json:"state,omitempty" yaml:"state,omitempty"`
	Status       string `json:"status,omitempty" yaml:"status,omitempty"`
}

func ResolveResource(cli *sdkClient.ZSClient, resourceType string, identifier string, options ResolveOptions) ([]ResolvedResource, error) {
	ctx := context.Background()
	if cli == nil {
		return nil, fmt.Errorf("client is nil")
	}
	query := func(ctx context.Context, path string, params *param.QueryParam, result any) error {
		return cli.ZSHttpClient.List(ctx, path, params, result)
	}
	return resolveResource(ctx, query, resourceType, identifier, options)
}

type resourceQuery func(context.Context, string, *param.QueryParam, any) error

func resolveResource(ctx context.Context, query resourceQuery, resourceType, identifier string, options ResolveOptions) ([]ResolvedResource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(identifier) == "" {
		return nil, fmt.Errorf("%s identifier is required", resourceType)
	}

	normalizedType, err := normalizeResolvableResourceType(resourceType)
	if err != nil {
		return nil, err
	}

	strategy := options.Strategy
	if strategy == "" {
		strategy = ResolveAuto
	}

	var matches []ResolvedResource
	switch strategy {
	case ResolveUUID:
		matches, err = queryResolvableResource(ctx, query, normalizedType, fmt.Sprintf("uuid=%s", identifier))
	case ResolveExactName:
		matches, err = queryResolvableResource(ctx, query, normalizedType, fmt.Sprintf("name=%s", identifier))
	case ResolveFuzzyName:
		matches, err = queryResolvableResource(ctx, query, normalizedType, fmt.Sprintf("name~=%s", identifier))
	case ResolveAuto:
		matches, err = queryResolvableResource(ctx, query, normalizedType, fmt.Sprintf("uuid=%s", identifier))
		if err == nil && len(matches) == 0 {
			matches, err = queryResolvableResource(ctx, query, normalizedType, fmt.Sprintf("name=%s", identifier))
		}
	default:
		return nil, fmt.Errorf("unsupported resolve strategy %q", strategy)
	}
	if err != nil {
		return nil, err
	}
	for _, match := range matches {
		if strings.TrimSpace(match.UUID) == "" {
			return nil, fmt.Errorf("%s query returned a resource without UUID", normalizedType)
		}
	}
	if len(matches) == 0 {
		return nil, clierr.Wrap(clierr.NotFound, fmt.Errorf("%s with name or UUID %q not found", normalizedType, identifier))
	}
	if len(matches) > 1 && !options.AllowMultiple {
		return nil, clierr.Wrap(clierr.Conflict, fmt.Errorf("multiple %s resources matched %q; use a more specific name or UUID", normalizedType, identifier))
	}
	return matches, nil
}

func ResolveOneResource(cli *sdkClient.ZSClient, resourceType string, identifier string) (ResolvedResource, error) {
	matches, err := ResolveResource(cli, resourceType, identifier, ResolveOptions{})
	if err != nil {
		return ResolvedResource{}, err
	}
	return matches[0], nil
}

func GetResolvedResource(cli *sdkClient.ZSClient, resourceType string, uuid string) (ResolvedResource, error) {
	matches, err := ResolveResource(cli, resourceType, uuid, ResolveOptions{Strategy: ResolveUUID})
	if err != nil {
		return ResolvedResource{}, err
	}
	return matches[0], nil
}

func normalizeResolvableResourceType(resourceType string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(resourceType))
	normalized = strings.ReplaceAll(normalized, "_", "-")

	switch normalized {
	case "account", "user", "role", "policy", "project", "organization":
		return normalized, nil
	case "zone":
		return "zone", nil
	case "cluster":
		return "cluster", nil
	case "host", "kvm-host", "kvmhost":
		return "host", nil
	case "instance", "vm", "vm-instance", "virtual-machine", "virtualmachine":
		return "instance", nil
	case "image":
		return "image", nil
	case "disk-offering", "diskoffering":
		return "disk-offering", nil
	case "instance-offering", "instanceoffering":
		return "instance-offering", nil
	case "volume", "vol":
		return "volume", nil
	case "snapshot", "volume-snapshot", "volumesnapshot":
		return "volume-snapshot", nil
	case "l3-network", "l3network", "network":
		return "l3-network", nil
	case "l2-network", "l2network":
		return "l2-network", nil
	case "ip-range", "iprange":
		return "ip-range", nil
	case "primary-storage", "primarystorage", "ps":
		return "primary-storage", nil
	case "backup-storage", "backupstorage", "bs", "image-storage":
		return "backup-storage", nil
	case "nic", "vm-nic", "vmnic":
		return "nic", nil
	case "vip":
		return "vip", nil
	case "eip":
		return "eip", nil
	case "port-forwarding", "port-forwarding-rule", "pf":
		return "port-forwarding", nil
	case "load-balancer", "loadbalancer", "lb":
		return "load-balancer", nil
	case "load-balancer-listener", "loadbalancerlistener", "lb-listener", "listener":
		return "load-balancer-listener", nil
	case "load-balancer-server-group", "loadbalancerservergroup", "lb-server-group", "server-group":
		return "load-balancer-server-group", nil
	case "vm-script", "vmscript", "guest-vm-script", "script":
		return "vm-script", nil
	case "image-group", "imagegroup":
		return "image-group", nil
	case "alarm":
		return "alarm", nil
	case "security-group", "securitygroup", "sg":
		return "security-group", nil
	case "long-job", "longjob", "job":
		return "long-job", nil
	default:
		return "", fmt.Errorf("unsupported resource type %q", resourceType)
	}
}

func queryResolvableResource(ctx context.Context, queryFn resourceQuery, resourceType string, query string) ([]ResolvedResource, error) {
	queryParam := param.NewQueryParam()
	queryParam.AddQ(query)

	if path, ok := map[string]string{"account": "v1/accounts", "user": "v1/accounts/users", "role": "v1/identities/roles", "policy": "v1/accounts/policies", "project": "v1/iam2/projects", "organization": "v1/iam2/organizations"}[resourceType]; ok {
		var result []struct {
			UUID  string `json:"uuid"`
			Name  string `json:"name"`
			State string `json:"state"`
		}
		if err := queryFn(ctx, path, &queryParam, &result); err != nil {
			return nil, err
		}
		matches := make([]ResolvedResource, 0, len(result))
		for _, item := range result {
			matches = append(matches, ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State})
		}
		return matches, nil
	}
	switch resourceType {
	case "zone":
		items, err := queryWithContext[view.ZoneInventoryView](ctx, queryFn, "v1/zones", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.ZoneInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "cluster":
		items, err := queryWithContext[view.ClusterInventoryView](ctx, queryFn, "v1/clusters", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.ClusterInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "host":
		items, err := queryWithContext[view.HostInventoryView](ctx, queryFn, "v1/hosts", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.HostInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State, Status: item.Status}
		}), nil
	case "instance":
		items, err := queryWithContext[view.VmInstanceInventoryView](ctx, queryFn, "v1/vm-instances", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.VmInstanceInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "image":
		items, err := queryWithContext[view.ImageInventoryView](ctx, queryFn, "v1/images", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.ImageInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State, Status: item.Status}
		}), nil
	case "disk-offering":
		items, err := queryWithContext[view.DiskOfferingInventoryView](ctx, queryFn, "v1/disk-offerings", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.DiskOfferingInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "instance-offering":
		items, err := queryWithContext[view.InstanceOfferingInventoryView](ctx, queryFn, "v1/instance-offerings", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.InstanceOfferingInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "volume":
		items, err := queryWithContext[view.VolumeInventoryView](ctx, queryFn, "v1/volumes", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.VolumeInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State, Status: item.Status}
		}), nil
	case "volume-snapshot":
		items, err := queryWithContext[view.VolumeSnapshotInventoryView](ctx, queryFn, "v1/volume-snapshots", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.VolumeSnapshotInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State, Status: item.Status}
		}), nil
	case "l3-network":
		items, err := queryWithContext[view.L3NetworkInventoryView](ctx, queryFn, "v1/l3-networks", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.L3NetworkInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "l2-network":
		items, err := queryWithContext[view.L2NetworkInventoryView](ctx, queryFn, "v1/l2-networks", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.L2NetworkInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name}
		}), nil
	case "ip-range":
		items, err := queryWithContext[view.IpRangeInventoryView](ctx, queryFn, "v1/l3-networks/ip-ranges", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.IpRangeInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name}
		}), nil
	case "primary-storage":
		items, err := queryWithContext[view.PrimaryStorageInventoryView](ctx, queryFn, "v1/primary-storage", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.PrimaryStorageInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State, Status: item.Status}
		}), nil
	case "backup-storage":
		items, err := queryWithContext[view.BackupStorageInventoryView](ctx, queryFn, "v1/backup-storage", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.BackupStorageInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State, Status: item.Status}
		}), nil
	case "nic":
		items, err := queryWithContext[view.VmNicInventoryView](ctx, queryFn, "v1/vm-instances/nics", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.VmNicInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "vip":
		items, err := queryWithContext[view.VipInventoryView](ctx, queryFn, "v1/vips", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.VipInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "eip":
		items, err := queryWithContext[view.EipInventoryView](ctx, queryFn, "v1/eips", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.EipInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "port-forwarding":
		items, err := queryWithContext[view.PortForwardingRuleInventoryView](ctx, queryFn, "v1/port-forwarding", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.PortForwardingRuleInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "load-balancer":
		items, err := queryWithContext[view.LoadBalancerInventoryView](ctx, queryFn, "v1/load-balancers", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.LoadBalancerInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "load-balancer-listener":
		items, err := queryWithContext[view.LoadBalancerListenerInventoryView](ctx, queryFn, "v1/load-balancers/listeners", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.LoadBalancerListenerInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name}
		}), nil
	case "load-balancer-server-group":
		items, err := queryWithContext[view.LoadBalancerServerGroupInventoryView](ctx, queryFn, "v1/load-balancers/servergroups", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.LoadBalancerServerGroupInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name}
		}), nil
	case "vm-script":
		items, err := queryWithContext[view.GuestVmScriptInventoryView](ctx, queryFn, "v1/scripts", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.GuestVmScriptInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name}
		}), nil
	case "image-group":
		items, err := queryWithContext[view.ImageGroupInventoryView](ctx, queryFn, "v1/imagegroups", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.ImageGroupInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, Status: item.Status}
		}), nil
	case "alarm":
		items, err := queryWithContext[view.AlarmInventoryView](ctx, queryFn, "v1/zwatch/alarms", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.AlarmInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State, Status: item.Status}
		}), nil
	case "security-group":
		items, err := queryWithContext[view.SecurityGroupInventoryView](ctx, queryFn, "v1/security-groups", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.SecurityGroupInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State}
		}), nil
	case "long-job":
		items, err := queryWithContext[view.LongJobInventoryView](ctx, queryFn, "v1/longjobs", &queryParam)
		if err != nil {
			return nil, err
		}
		return mapResolved(items, resourceType, func(item view.LongJobInventoryView) ResolvedResource {
			return ResolvedResource{ResourceType: resourceType, UUID: item.UUID, Name: item.Name, State: item.State, JobResult: item.JobResult}
		}), nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q", resourceType)
	}
}

func mapResolved[T any](items []T, resourceType string, mapper func(T) ResolvedResource) []ResolvedResource {
	resolved := make([]ResolvedResource, 0, len(items))
	for _, item := range items {
		next := mapper(item)
		next.ResourceType = resourceType
		resolved = append(resolved, next)
	}
	return resolved
}

func queryWithContext[T any](ctx context.Context, queryFn resourceQuery, path string, query *param.QueryParam) ([]T, error) {
	var items []T
	if err := queryFn(ctx, path, query, &items); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
