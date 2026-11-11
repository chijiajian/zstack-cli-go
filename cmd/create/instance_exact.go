package create

import (
	"fmt"
	"slices"
	"strings"

	"github.com/chijiajian/zstack-cli-go/cmd/vmops"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

func createInstanceByUUID(cmd *cobra.Command, name string) error {
	flags := cmd.Flags()
	allowed := []string{"uuid-only", "request-timeout", "output", "yes", "dry-run", "context", "error-format", "image", "l3-network", "instance-offering", "cpu", "memory", "description", "resource-uuid", "strategy"}
	var unsupported string
	flags.Visit(func(flag *pflag.Flag) {
		if !slices.Contains(allowed, flag.Name) {
			unsupported = flag.Name
		}
	})
	if unsupported != "" {
		return vmops.Invalid(fmt.Sprintf("--%s is unavailable in exact UUID mode", unsupported))
	}
	image, _ := flags.GetString("image")
	networks, _ := flags.GetStringSlice("l3-network")
	offering, _ := flags.GetString("instance-offering")
	cpu, _ := flags.GetInt64("cpu")
	memory, _ := flags.GetString("memory")
	description, _ := flags.GetString("description")
	uuid, _ := flags.GetString("resource-uuid")
	strategy, _ := flags.GetString("strategy")
	if strings.TrimSpace(name) == "" || image == "" || len(networks) == 0 || uuid == "" {
		return vmops.Invalid("name, image UUID, L3 network UUID and resource UUID are required")
	}
	if !flags.Changed("strategy") || strategy != "CreateStopped" && strategy != "InstantStart" {
		return vmops.Invalid("explicit --strategy CreateStopped or InstantStart is required")
	}
	p := param.CreateVmInstanceParamDetail{
		Name: name, ImageUuid: &image, L3NetworkUuids: networks, ResourceUuid: &uuid, Strategy: &strategy,
	}
	if flags.Changed("description") {
		p.Description = &description
	}
	if offering != "" {
		if flags.Changed("cpu") || flags.Changed("memory") {
			return vmops.Invalid("use either an instance offering or explicit CPU and memory")
		}
		p.InstanceOfferingUuid = &offering
	} else {
		size, err := utils.ParseMemorySize(memory)
		if err != nil || size <= 0 || cpu <= 0 || cpu > 2147483647 {
			return vmops.Invalid("positive CPU and memory are required without an instance offering")
		}
		count := int(cpu)
		p.CpuNum, p.MemorySize = &count, &size
	}
	return vmops.Create(cmd, param.CreateVmInstanceParam{Params: p})
}
