package resources

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

type failedResultWriter struct{ err error }

func (w failedResultWriter) Write([]byte) (int, error) { return 0, w.err }
func TestResourceResultsReportOutputFailure(t *testing.T) {
	for name, printResult := range map[string]func(*cobra.Command, string, interface{}) error{
		"volume": printVolumeResult, "instance": printInstanceResult, "image": printImageResult, "snapshot": printSnapshotResult, "eip": printEipResult, "vip-related": printInfrastructureResult, "global-config": printGlobalConfigResult, "load-balancer": printLoadBalancerResult, "port-forwarding": printPortForwardingResult, "security-group": printSecurityGroupResult, "vm-script": printVmScriptResult,
	} {
		t.Run(name, func(t *testing.T) {
			for _, format := range []string{"", "text", "name"} {
				cmd := &cobra.Command{}
				cmd.Flags().String("output", format, "")
				cmd.Flags().StringSlice("fields", nil, "")
				want := errors.New("output unavailable")
				cmd.SetOut(failedResultWriter{want})
				if err := printResult(cmd, "volume/demo resized", nil); !errors.Is(err, want) {
					t.Fatalf("format %q: got %v", format, err)
				}
			}
		})
	}

}
