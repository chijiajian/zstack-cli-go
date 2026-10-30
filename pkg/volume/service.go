// Package volume contains volume operation policies independent of CLI parsing.
package volume

import (
	"fmt"

	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

type Reader interface {
	GetVolume(string) (*view.VolumeInventoryView, error)
}
type Resizer interface {
	ResizeDataVolume(string, param.ResizeDataVolumeParam) (*view.VolumeInventoryView, error)
	ResizeRootVolume(string, param.ResizeRootVolumeParam) (*view.VolumeInventoryView, error)
}
type Recoverer interface {
	RecoverDataVolume(string, param.RecoverDataVolumeParam) (*view.VolumeInventoryView, error)
}

func Resize(reader Reader, writer Resizer, uuid string, size int64) (*view.VolumeInventoryView, error) {
	if size <= 0 {
		return nil, fmt.Errorf("size must be positive")
	}
	current, err := reader.GetVolume(uuid)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("volume detail missing")
	}
	if current.Type != "Data" && current.Type != "Root" {
		return nil, fmt.Errorf("unsupported volume type %q", current.Type)
	}
	if current.Status != "Ready" {
		return nil, fmt.Errorf("volume must be Ready to resize, got %q", current.Status)
	}
	if current.Size <= 0 {
		return nil, fmt.Errorf("volume current size is invalid")
	}
	if size <= current.Size {
		return nil, fmt.Errorf("new size %d must be greater than current size %d; shrinking is not supported", size, current.Size)
	}
	if current.Type == "Data" {
		return writer.ResizeDataVolume(uuid, param.ResizeDataVolumeParam{Params: param.ResizeDataVolumeParamDetail{Size: size}})
	}
	p := param.ResizeRootVolumeParam{Params: param.ResizeRootVolumeParamDetail{Size: size}}
	if current.VmInstanceUuid != "" {
		p.Params.VmInstanceUuid = &current.VmInstanceUuid
	}
	return writer.ResizeRootVolume(uuid, p)
}

func Recover(reader Reader, writer Recoverer, uuid string) (*view.VolumeInventoryView, error) {
	current, err := reader.GetVolume(uuid)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("volume detail missing")
	}
	if current.Type != "Data" {
		return nil, fmt.Errorf("only data volumes can be recovered independently; use instance recover for a root volume")
	}
	if current.Status != "Deleted" {
		return nil, fmt.Errorf("volume must be Deleted to recover, got %q", current.Status)
	}
	return writer.RecoverDataVolume(uuid, param.RecoverDataVolumeParam{})
}
