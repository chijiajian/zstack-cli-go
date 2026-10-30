package utils

// cmdutils/vm_output.go
import (
	"fmt"
	"strings"

	sdkView "github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

// VMRow
type VMRow struct {
	Name   string `json:"name"   yaml:"name"   header:"NAME"`
	UUID   string `json:"uuid"   yaml:"uuid"   header:"UUID"`
	State  string `json:"state"  yaml:"state"  header:"STATE"`
	CPU    int    `json:"cpu"    yaml:"cpu"    header:"CPU"`
	IPs    string `json:"ips"    yaml:"ips"    header:"IPS"`
	Memory string `json:"memory" yaml:"memory" header:"MEMORY"`
}

// ConvertVMs
func ConvertVMs(vms []sdkView.VmInstanceInventoryView) []VMRow {
	var rows []VMRow
	for _, vm := range vms {
		var ips []string
		for _, nic := range vm.VmNics {
			if nic.Ip != "" {
				ips = append(ips, nic.Ip)
			}
		}

		rows = append(rows, VMRow{
			Name:   vm.Name,
			UUID:   vm.UUID,
			CPU:    vm.CpuNum,
			Memory: FormatMemorySize(vm.MemorySize),
			State:  vm.State,
			IPs:    strings.Join(ips, ", "),
		})
	}
	return rows
}

// PrintVMs Print VMs 的list
func PrintVMs(vms []sdkView.VmInstanceInventoryView, format OutputFormat, fields []string) error {
	rows := ConvertVMs(vms)
	return PrintWithFields(rows, format, fields)
}

// PrintSummary
func PrintSummary(successes, failures, skipped []sdkView.VmInstanceInventoryView) {
	fmt.Printf("\nSummary: %d started, %d failed, %d skipped\n",
		len(successes), len(failures), len(skipped))

	if len(failures) > 0 {
		fmt.Println("Failures:")
		for _, f := range failures {
			fmt.Printf("  - %s (%s)\n", f.Name, f.UUID)
		}
	}
	if len(skipped) > 0 {
		fmt.Println("Skipped (not Stopped):")
		for _, s := range skipped {
			fmt.Printf("  - %s (%s) state=%s\n", s.Name, s.UUID, s.State)
		}
	}
}
