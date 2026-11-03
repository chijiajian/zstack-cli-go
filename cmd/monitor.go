package cmd

import (
	"bytes"
	"fmt"
	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"github.com/chijiajian/zstack-cli-go/pkg/monitor"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"io"
	"text/tabwriter"
	"time"
)

func init() {
	group := &cobra.Command{Use: "monitor", Short: "Monitor cloud capacity and host load"}
	usage := &cobra.Command{Use: "usage", Short: "Show cloud capacity and sampled host CPU and memory utilization", Args: cobra.NoArgs, RunE: runUsage}
	usage.Flags().String("section", "all", "Report section: all, capacity or load")
	group.AddCommand(usage)
	rootCmd.AddCommand(group)
}
func runUsage(cmd *cobra.Command, _ []string) error {
	section, _ := cmd.Flags().GetString("section")
	format, _ := cmd.Flags().GetString("output")
	if section != "all" && section != "capacity" && section != "load" {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("unknown section %q", section))
	}
	if format != "table" && format != "text" && format != "json" && format != "yaml" {
		return clierr.Wrap(clierr.Usage, fmt.Errorf("unsupported monitor output %q", format))
	}
	cli, err := client.GetClientContext(cmd.Context())
	if err != nil {
		return err
	}
	var report monitor.Report
	if section != "load" {
		report.Capacity, err = monitor.ReadCapacity(cli)
		if err != nil {
			return fmt.Errorf("read capacity: %w", err)
		}
	}
	if section != "capacity" {
		report.Load, err = monitor.ReadLoad(cli, time.Now())
		if err != nil {
			return fmt.Errorf("read host load: %w", err)
		}
	}
	if format == "json" || format == "yaml" {
		err = utils.Print(report, utils.ParseFormat(format))
	} else {
		err = printUsage(cmd.OutOrStdout(), report)
	}
	if err != nil {
		return clierr.Wrap(clierr.Output, err)
	}
	return nil
}
func percent(v *float64) string {
	if v == nil {
		return "N/A"
	}
	return fmt.Sprintf("%.2f%%", *v)
}
func printUsage(out io.Writer, r monitor.Report) error {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', 0)
	if len(r.Capacity) > 0 {
		fmt.Fprintln(w, "CAPACITY\tUSED\tTOTAL\tAVAILABLE\tUSED %")
		for _, c := range r.Capacity {
			size := func(n int64) string {
				if c.Unit == "bytes" {
					return fmt.Sprintf("%.2f GiB", float64(n)/(1<<30))
				}
				return fmt.Sprintf("%d vCPU", n)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", c.Metric, size(c.Used), size(c.Total), size(c.Available), percent(c.Percent))
		}
	}
	if l := r.Load; l != nil {
		stamp := "N/A"
		if l.SampleTime != nil {
			stamp = l.SampleTime.Local().Format(time.RFC3339)
		}
		fmt.Fprintln(w, "HOST LOAD\tCPU\tMEMORY\tSAMPLED/CONNECTED\tSAMPLE TIME")
		fmt.Fprintf(w, "Mean\t%s\t%s\t%d/%d\t%s\n", percent(l.CPU), percent(l.Memory), l.SampledHosts, l.ConnectedHosts, stamp)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	_, err := buf.WriteTo(out)
	return err
}
