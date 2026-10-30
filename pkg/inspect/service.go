// Package inspect assembles diagnostic views without depending on Cobra or rendering.
package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type Reader interface {
	Read(context.Context, string, url.Values, any) error
}
type Service struct{ Reader Reader }
type relationRef struct{ key, field, path string }
type Document struct {
	Kind     string         `json:"kind" yaml:"kind"`
	Resource map[string]any `json:"resource" yaml:"resource"`
	Related  map[string]any `json:"related" yaml:"related"`
	Warnings []string       `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

func (s Service) one(ctx context.Context, path, uuid string) (map[string]any, error) {
	var result struct {
		Inventories []map[string]any `json:"inventories"`
	}
	endpoint := path + "/" + url.PathEscape(uuid)
	var query url.Values
	// Some ZStack versions expose primary storage only through the query API.
	if path == "v1/primary-storage" {
		endpoint = path
		query = url.Values{"q": {"uuid=" + uuid}, "limit": {"2"}}
	}
	if err := s.Reader.Read(ctx, endpoint, query, &result); err != nil {
		return nil, err
	}
	if len(result.Inventories) != 1 {
		return nil, fmt.Errorf("expected one resource for %s/%s, got %d", path, uuid, len(result.Inventories))
	}
	actual, _ := result.Inventories[0]["uuid"].(string)
	if actual == "" || actual != uuid {
		return nil, fmt.Errorf("resource identity mismatch for %s/%s: got UUID %q", path, uuid, actual)
	}
	normalizeNumbers(result.Inventories[0])
	return result.Inventories[0], nil
}

func (s Service) Describe(ctx context.Context, kind, uuid string) (Document, error) {
	path := "v1/vm-instances"
	if kind == "volume" {
		path = "v1/volumes"
	} else if kind != "instance" {
		return Document{}, fmt.Errorf("unsupported detail kind %q", kind)
	}
	resource, err := s.one(ctx, path, uuid)
	if err != nil {
		return Document{}, err
	}
	doc := Document{Kind: kind, Resource: resource, Related: map[string]any{}}
	refs := []relationRef{}
	if kind == "instance" {
		refs = append(refs, relationRef{"host", "hostUuid", "v1/hosts"}, relationRef{"image", "imageUuid", "v1/images"}, relationRef{"instanceOffering", "instanceOfferingUuid", "v1/instance-offerings"})
		networks := map[string]any{}
		if nics, ok := resource["vmNics"].([]any); ok {
			for _, entry := range nics {
				nic, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				id, _ := nic["l3NetworkUuid"].(string)
				if id == "" {
					continue
				}
				if _, done := networks[id]; done {
					continue
				}
				network, err := s.one(ctx, "v1/l3-networks", id)
				if err != nil {
					networks[id] = nil
					doc.Warnings = append(doc.Warnings, fmt.Sprintf("l3-network %s unavailable: %v", id, err))
				} else {
					networks[id] = network
				}
			}
		}
		doc.Related["l3Networks"] = networks
	} else {
		refs = append(refs, relationRef{"instance", "vmInstanceUuid", "v1/vm-instances"}, relationRef{"primaryStorage", "primaryStorageUuid", "v1/primary-storage"}, relationRef{"diskOffering", "diskOfferingUuid", "v1/disk-offerings"})
	}
	for _, ref := range refs {
		id, _ := resource[ref.field].(string)
		if id == "" {
			continue
		}
		related, err := s.one(ctx, ref.path, id)
		if err != nil {
			doc.Related[ref.key] = nil
			doc.Warnings = append(doc.Warnings, fmt.Sprintf("%s %s unavailable: %v", ref.key, id, err))
			continue
		}
		doc.Related[ref.key] = related
	}
	// Cancellation is an operation failure, not a partial permissions warning.
	if err := ctx.Err(); err != nil {
		return Document{}, err
	}
	return doc, nil
}

func normalizeNumbers(value any) any {
	switch value := value.(type) {
	case json.Number:
		if integer, err := value.Int64(); err == nil {
			return integer
		}
		if unsigned, err := strconv.ParseUint(string(value), 10, 64); err == nil {
			return unsigned
		}
		if decimal, err := value.Float64(); err == nil {
			return decimal
		}
		return value
	case map[string]any:
		for key, item := range value {
			value[key] = normalizeNumbers(item)
		}
	case []any:
		for i, item := range value {
			value[i] = normalizeNumbers(item)
		}
	}
	return value
}
