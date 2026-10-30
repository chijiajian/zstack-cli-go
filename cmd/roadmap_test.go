package cmd

import (
	"strings"
	"testing"
)

func TestRoadmapCommandCoverage(t *testing.T) {
	groups := map[string][]string{
		"update":          {"instance", "image", "disk-offering", "instance-offering", "volume", "l3-network", "vip", "eip", "security-group", "vm-script", "zone", "cluster", "host", "primary-storage", "backup-storage", "l2-network"},
		"create":          {"instance", "image", "volume", "disk-offering", "instance-offering", "vm-script", "vip", "eip", "security-group", "port-forwarding", "load-balancer", "load-balancer-listener", "load-balancer-server-group", "zone", "cluster", "l2-network", "l3-network", "alarm"},
		"delete":          {"instance", "image", "volume", "snapshot", "vm-script", "vip", "eip", "security-group", "port-forwarding", "load-balancer", "load-balancer-listener", "load-balancer-server-group", "zone", "cluster", "host", "primary-storage", "backup-storage", "l2-network", "l3-network", "ip-range", "alarm"},
		"expunge":         {"instance", "image", "volume", "image-group"},
		"wait":            {"instance", "image", "volume", "job"},
		"describe":        {"instance", "volume", "long-job"},
		"instance":        {"start", "stop", "restart", "pause", "resume", "recover", "update", "attach-l3", "detach-l3", "attach-iso", "detach-iso", "attach-volume", "detach-volume"},
		"volume":          {"attach", "detach", "recover", "resize"},
		"snapshot":        {"create", "revert"},
		"vm-script":       {"execute"},
		"eip":             {"attach", "detach"},
		"security-group":  {"attach-l3", "detach-l3", "add-rule", "delete-rule", "add-nic", "remove-nic"},
		"port-forwarding": {"attach", "detach"},
		"load-balancer":   {"add-nic", "remove-nic"},
		"host":            {"add-kvm"},
		"job":             {"resume", "clean"},
		"project":         {"recover"},
		"primary-storage": {"add-nfs", "add-local", "attach-cluster", "detach-cluster"},
		"backup-storage":  {"add-image-store", "attach-zone", "detach-zone"},
		"l2-network":      {"attach-cluster", "detach-cluster"},
		"l3-network":      {"add-ip-range", "add-dns", "remove-dns", "add-host-route", "remove-host-route"},
		"image":           {"recover", "sync", "create-from-volume", "create-from-snapshot"},
		"image-group":     {"create-from-image"},
		"global-config":   {"set"},
		"config":          {"view", "get-contexts", "current-context", "use-context", "rename-context", "delete-context"},
	}
	groups["create"] = append(groups["create"], "account", "user", "role", "policy", "project", "organization")
	groups["update"] = append(groups["update"], "account", "user", "role", "project", "organization")
	groups["delete"] = append(groups["delete"], "account", "user", "role", "policy", "project", "organization")
	groups["expunge"] = append(groups["expunge"], "project")
	for parent, children := range groups {
		for _, child := range children {
			t.Run(parent+"/"+child, func(t *testing.T) {
				command, remaining, err := rootCmd.Find([]string{parent, child})
				if err != nil || len(remaining) != 0 || command == nil || !command.Runnable() {
					t.Fatalf("missing command %s %s: %v %v", parent, child, remaining, err)
				}
				if parent == "delete" || parent == "expunge" || parent == "snapshot" && child == "revert" {
					if command.Flags().Lookup("yes") == nil && command.InheritedFlags().Lookup("yes") == nil {
						t.Fatal("dangerous operation missing --yes")
					}
				}
			})
		}
	}
	command, rest, err := rootCmd.Find([]string{"logout"})
	if err != nil || len(rest) != 0 || command.Name() != "logout" {
		t.Fatal("logout missing")
	}
}
func TestQueriesRetainSharedFlags(t *testing.T) {
	get, _, _ := rootCmd.Find([]string{"get"})
	for _, cmd := range get.Commands() {
		if cmd.Hidden || !cmd.Runnable() {
			continue
		}
		for _, name := range strings.Fields("q limit start sort fields output all-pages count reply-with-count request-timeout") {
			if cmd.Flags().Lookup(name) == nil && cmd.InheritedFlags().Lookup(name) == nil {
				t.Errorf("get %s missing --%s", cmd.Name(), name)
			}
		}
	}
}
