package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runCLIWithConfig(t, filepath.Join(t.TempDir(), "config.yaml"), args...)
}

func runCLIWithConfig(t *testing.T, configPath string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(cliBinary, args...)
	cmd.Env = append(os.Environ(), "ZSTACK_CONFIG="+configPath)

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestRootHelpRegistersLoginOnce(t *testing.T) {
	out, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("expected help to succeed, got %v\n%s", err, out)
	}

	if got := countHelpCommand(out, "login", "Login to ZStack API server"); got != 1 {
		t.Fatalf("expected one login command in help, got %d\n%s", got, out)
	}
}

func countHelpCommand(out string, command string, description string) int {
	count := 0
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != command {
			continue
		}
		if strings.Contains(line, description) {
			count++
		}
	}
	return count
}

func TestVersionFlag(t *testing.T) {
	out, err := runCLI(t, "--version")
	if err != nil {
		t.Fatalf("expected --version to succeed, got %v\n%s", err, out)
	}

	if !strings.Contains(out, "zstack-cli version:") {
		t.Fatalf("expected version output, got\n%s", out)
	}
}

func TestUnauthenticatedGetReturnsNonZero(t *testing.T) {
	out, err := runCLI(t, "get", "images")
	if err == nil {
		t.Fatalf("expected unauthenticated get to fail\n%s", out)
	}

	if !strings.Contains(out, "Not logged in") {
		t.Fatalf("expected login error, got\n%s", out)
	}
}

func TestLoginWithoutEndpointReturnsNonZero(t *testing.T) {
	out, err := runCLI(t, "login")
	if err == nil {
		t.Fatalf("expected login without endpoint to fail\n%s", out)
	}

	if !strings.Contains(out, "no endpoint specified") {
		t.Fatalf("expected endpoint error, got\n%s", out)
	}
}

func TestLoginAccessKeySavesContext(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	out, err := runCLIWithConfig(t, configPath, "login", "127.0.0.1", "--access-key-id", "ak-test", "--access-key-secret", "sk-test")
	if err != nil {
		t.Fatalf("expected access key login to save config without contacting server, got %v\n%s", err, out)
	}
	if !strings.Contains(out, "Saved access key authentication") {
		t.Fatalf("expected access key save message, got\n%s", out)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config to be written: %v", err)
	}
	configText := string(data)
	for _, expected := range []string{
		"auth_type: access-key",
		"access_key_id: ak-test",
		"access_key_secret: sk-test",
	} {
		if !strings.Contains(configText, expected) {
			t.Fatalf("expected config to contain %q, got\n%s", expected, configText)
		}
	}

	view, err := runCLIWithConfig(t, configPath, "config", "view")
	if err != nil {
		t.Fatalf("expected config view to succeed, got %v\n%s", err, view)
	}
	if !strings.Contains(view, "auth_type: access-key") {
		t.Fatalf("expected config view to show access-key auth type, got\n%s", view)
	}
	if strings.Contains(view, "ak-test") || strings.Contains(view, "sk-test") {
		t.Fatalf("expected access key credentials to be masked, got\n%s", view)
	}
}

func TestLoginAccessKeyRequiresIDAndSecret(t *testing.T) {
	out, err := runCLI(t, "login", "127.0.0.1", "--access-key-id", "ak-test")
	if err == nil {
		t.Fatalf("expected incomplete access key login to fail\n%s", out)
	}
	if !strings.Contains(out, "both --access-key-id and --access-key-secret are required") {
		t.Fatalf("expected access key validation error, got\n%s", out)
	}
}

func TestCreateParentFileDoesNotSucceedSilently(t *testing.T) {
	out, err := runCLI(t, "create", "--file", "test/image/create_image.yaml")
	if err == nil {
		t.Fatalf("expected create --file on parent command to fail\n%s", out)
	}

	if !strings.Contains(out, "create --file is not supported on the parent command") {
		t.Fatalf("expected explicit create --file error, got\n%s", out)
	}
}

func TestCreateImageMissingRequiredFlagReturnsNonZero(t *testing.T) {
	out, err := runCLI(t, "create", "image", "demo")
	if err == nil {
		t.Fatalf("expected create image without required flags to fail\n%s", out)
	}

	if !strings.Contains(out, "required flag --url not set") {
		t.Fatalf("expected missing URL error, got\n%s", out)
	}
}

func TestCreateImageHelpUsesImageStorageFlag(t *testing.T) {
	out, err := runCLI(t, "create", "image", "--help")
	if err != nil {
		t.Fatalf("expected help to succeed, got %v\n%s", err, out)
	}

	if strings.Contains(out, "--backup-storage") {
		t.Fatalf("help still references removed --backup-storage flag\n%s", out)
	}
	if !strings.Contains(out, "--image-storage") {
		t.Fatalf("help should reference --image-storage\n%s", out)
	}
}

func TestDeletePluralAlias(t *testing.T) {
	out, err := runCLI(t, "delete", "images", "demo")
	if err == nil {
		t.Fatalf("expected unauthenticated delete to fail\n%s", out)
	}

	if strings.Contains(out, "invalid argument") || strings.Contains(out, "unknown command") {
		t.Fatalf("expected images alias to resolve to delete image, got\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "not logged in") {
		t.Fatalf("expected login error, got\n%s", out)
	}
}

func TestP0CommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("expected help to succeed, got %v\n%s", err, out)
	}

	for _, command := range []string{"update", "wait", "describe"} {
		if !strings.Contains(out, command) {
			t.Fatalf("expected root help to include %q\n%s", command, out)
		}
	}
}

func TestUpdateHelpIncludesP0Resources(t *testing.T) {
	out, err := runCLI(t, "update", "--help")
	if err != nil {
		t.Fatalf("expected update help to succeed, got %v\n%s", err, out)
	}

	for _, resource := range []string{
		"instance", "image", "disk-offering", "instance-offering",
		"volume", "l3-network", "vip", "eip", "security-group",
	} {
		if !strings.Contains(out, resource) {
			t.Fatalf("expected update help to include %q\n%s", resource, out)
		}
	}
}

func TestUpdateDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "update", "image", "demo", "--name", "renamed", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected update dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"resource": "image"`) || !strings.Contains(out, `"target": "demo"`) || !strings.Contains(out, `"name": "renamed"`) {
		t.Fatalf("expected dry-run request output, got\n%s", out)
	}
}

func TestUpdateWithoutFieldsReturnsNonZero(t *testing.T) {
	out, err := runCLI(t, "update", "image", "demo")
	if err == nil {
		t.Fatalf("expected update without fields to fail\n%s", out)
	}

	if !strings.Contains(out, "at least one update field must be specified") {
		t.Fatalf("expected explicit missing fields error, got\n%s", out)
	}
}

func TestWaitRequiresForFlag(t *testing.T) {
	out, err := runCLI(t, "wait", "instance", "demo")
	if err == nil {
		t.Fatalf("expected wait without --for to fail\n%s", out)
	}

	if !strings.Contains(out, `required flag(s) "for" not set`) {
		t.Fatalf("expected required --for error, got\n%s", out)
	}
}

func TestVolumeCommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "volume", "--help")
	if err != nil {
		t.Fatalf("expected volume help to succeed, got %v\n%s", err, out)
	}

	for _, command := range []string{"attach", "detach"} {
		if !strings.Contains(out, command) {
			t.Fatalf("expected volume help to include %q\n%s", command, out)
		}
	}
}

func TestCreateVolumeDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "create", "volume", "data-01", "--size", "10G", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected create volume dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"name": "data-01"`) || !strings.Contains(out, `"diskSize": 10737418240`) {
		t.Fatalf("expected create volume dry-run request, got\n%s", out)
	}
}

func TestVolumeAttachDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "volume", "attach", "data-01", "--instance", "vm-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected volume attach dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"volume": "data-01"`) || !strings.Contains(out, `"instance": "vm-01"`) {
		t.Fatalf("expected volume attach dry-run request, got\n%s", out)
	}
}

func TestDeleteAndExpungeVolumeDryRunDoesNotRequireLogin(t *testing.T) {
	for _, args := range [][]string{
		{"delete", "volume", "data-01", "--dry-run", "-o", "json"},
		{"expunge", "volume", "data-01", "--dry-run", "-o", "json"},
	} {
		out, err := runCLI(t, args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(args[:2], " "), err, out)
		}
		if !strings.Contains(out, `"volume": "data-01"`) {
			t.Fatalf("expected volume dry-run request, got\n%s", out)
		}
	}
}

func TestSnapshotCommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "snapshot", "--help")
	if err != nil {
		t.Fatalf("expected snapshot help to succeed, got %v\n%s", err, out)
	}

	for _, command := range []string{"create", "revert"} {
		if !strings.Contains(out, command) {
			t.Fatalf("expected snapshot help to include %q\n%s", command, out)
		}
	}
}

func TestSnapshotCreateDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "snapshot", "create", "volume-01", "--name", "snap-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected snapshot create dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"volume": "volume-01"`) || !strings.Contains(out, `"name": "snap-01"`) {
		t.Fatalf("expected snapshot create dry-run request, got\n%s", out)
	}
}

func TestSnapshotRevertDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "snapshot", "revert", "snap-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected snapshot revert dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"snapshot": "snap-01"`) {
		t.Fatalf("expected snapshot revert dry-run request, got\n%s", out)
	}
}

func TestDeleteSnapshotDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "delete", "snapshot", "snap-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected delete snapshot dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"snapshot": "snap-01"`) {
		t.Fatalf("expected delete snapshot dry-run request, got\n%s", out)
	}
}

func TestInstanceAttachCommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "instance", "--help")
	if err != nil {
		t.Fatalf("expected instance help to succeed, got %v\n%s", err, out)
	}

	for _, command := range []string{
		"attach-l3", "detach-l3", "attach-iso", "detach-iso", "attach-volume", "detach-volume", "update",
	} {
		if !strings.Contains(out, command) {
			t.Fatalf("expected instance help to include %q\n%s", command, out)
		}
	}
}

func TestInstanceAttachDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"instance", "attach-l3", "vm-01", "--l3-network", "l3-01", "--dry-run", "-o", "json"},
			expected: []string{`"instance": "vm-01"`, `"target": "l3-01"`},
		},
		{
			args:     []string{"instance", "detach-l3", "vm-01", "--l3-network", "l3-01", "--dry-run", "-o", "json"},
			expected: []string{`"instance": "vm-01"`, `"target": "l3-01"`},
		},
		{
			args:     []string{"instance", "attach-iso", "vm-01", "--iso", "iso-01", "--dry-run", "-o", "json"},
			expected: []string{`"instance": "vm-01"`, `"target": "iso-01"`},
		},
		{
			args:     []string{"instance", "detach-iso", "vm-01", "--iso", "iso-01", "--dry-run", "-o", "json"},
			expected: []string{`"instance": "vm-01"`, `"target": "iso-01"`},
		},
		{
			args:     []string{"instance", "attach-volume", "vm-01", "--volume", "data-01", "--dry-run", "-o", "json"},
			expected: []string{`"instance": "vm-01"`, `"target": "data-01"`},
		},
		{
			args:     []string{"instance", "detach-volume", "vm-01", "--volume", "data-01", "--dry-run", "-o", "json"},
			expected: []string{`"instance": "vm-01"`, `"target": "data-01"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s to succeed without login, got %v\n%s", strings.Join(tc.args[:3], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestInstanceAttachRequiresTargetFlag(t *testing.T) {
	out, err := runCLI(t, "instance", "attach-l3", "vm-01")
	if err == nil {
		t.Fatalf("expected missing --l3-network to fail\n%s", out)
	}

	if !strings.Contains(out, "required flag --l3-network not set") {
		t.Fatalf("expected missing l3-network error, got\n%s", out)
	}
}

func TestInstanceUpdateDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "instance", "update", "vm-01", "--name", "renamed-vm", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected instance update dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"instance": "vm-01"`) || !strings.Contains(out, `"name": "renamed-vm"`) {
		t.Fatalf("expected instance update dry-run request, got\n%s", out)
	}
}

func TestInstanceUpdateWithoutFieldsReturnsNonZero(t *testing.T) {
	out, err := runCLI(t, "instance", "update", "vm-01")
	if err == nil {
		t.Fatalf("expected instance update without fields to fail\n%s", out)
	}

	if !strings.Contains(out, "at least one update field must be specified") {
		t.Fatalf("expected explicit missing fields error, got\n%s", out)
	}
}

func TestVipAndEipCreateDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"create", "vip", "vip-01", "--l3-network", "l3-01", "--dry-run", "-o", "json"},
			expected: []string{`"name": "vip-01"`, `"l3NetworkUuid": "l3-01"`},
		},
		{
			args:     []string{"create", "eip", "eip-01", "--vip", "vip-01", "--dry-run", "-o", "json"},
			expected: []string{`"name": "eip-01"`, `"vipUuid": "vip-01"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:3], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestEipCommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "eip", "--help")
	if err != nil {
		t.Fatalf("expected eip help to succeed, got %v\n%s", err, out)
	}

	for _, command := range []string{"attach", "detach"} {
		if !strings.Contains(out, command) {
			t.Fatalf("expected eip help to include %q\n%s", command, out)
		}
	}
}

func TestEipAttachDetachDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"eip", "attach", "eip-01", "--nic", "nic-01", "--dry-run", "-o", "json"},
			expected: []string{`"eip": "eip-01"`, `"nic": "nic-01"`},
		},
		{
			args:     []string{"eip", "detach", "eip-01", "--dry-run", "-o", "json"},
			expected: []string{`"eip": "eip-01"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestDeleteVipAndEipDryRunsDoNotRequireLogin(t *testing.T) {
	for _, args := range [][]string{
		{"delete", "vip", "vip-01", "--dry-run", "-o", "json"},
		{"delete", "eip", "eip-01", "--dry-run", "-o", "json"},
	} {
		out, err := runCLI(t, args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(args[:2], " "), err, out)
		}
		if !strings.Contains(out, `"target": "`) {
			t.Fatalf("expected delete network dry-run request, got\n%s", out)
		}
	}
}

func TestSecurityGroupCommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "security-group", "--help")
	if err != nil {
		t.Fatalf("expected security-group help to succeed, got %v\n%s", err, out)
	}

	for _, command := range []string{"attach-l3", "detach-l3", "add-rule", "delete-rule", "add-nic", "remove-nic"} {
		if !strings.Contains(out, command) {
			t.Fatalf("expected security-group help to include %q\n%s", command, out)
		}
	}
}

func TestSecurityGroupCreateDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "create", "security-group", "sg-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected security-group create dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"name": "sg-01"`) {
		t.Fatalf("expected security-group create dry-run request, got\n%s", out)
	}
}

func TestSecurityGroupOperationDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"security-group", "attach-l3", "sg-01", "--l3-network", "l3-01", "--dry-run", "-o", "json"},
			expected: []string{`"securityGroup": "sg-01"`, `"target": "l3-01"`},
		},
		{
			args:     []string{"security-group", "detach-l3", "sg-01", "--l3-network", "l3-01", "--dry-run", "-o", "json"},
			expected: []string{`"securityGroup": "sg-01"`, `"target": "l3-01"`},
		},
		{
			args:     []string{"security-group", "add-rule", "sg-01", "--type", "Ingress", "--protocol", "TCP", "--start-port", "22", "--end-port", "22", "--cidr", "0.0.0.0/0", "--dry-run", "-o", "json"},
			expected: []string{`"securityGroup": "sg-01"`, `"type": "Ingress"`, `"protocol": "TCP"`},
		},
		{
			args:     []string{"security-group", "delete-rule", "rule-01", "--dry-run", "-o", "json"},
			expected: []string{`"rule": "rule-01"`},
		},
		{
			args:     []string{"security-group", "add-nic", "sg-01", "--nic", "nic-01", "--dry-run", "-o", "json"},
			expected: []string{`"securityGroup": "sg-01"`, `"target": "nic-01"`},
		},
		{
			args:     []string{"security-group", "remove-nic", "sg-01", "--nic", "nic-01", "--dry-run", "-o", "json"},
			expected: []string{`"securityGroup": "sg-01"`, `"target": "nic-01"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestDeleteSecurityGroupDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "delete", "security-group", "sg-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected delete security-group dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"resource": "security-group"`) || !strings.Contains(out, `"target": "sg-01"`) {
		t.Fatalf("expected delete security-group dry-run request, got\n%s", out)
	}
}

func TestPortForwardingCommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "port-forwarding", "--help")
	if err != nil {
		t.Fatalf("expected port-forwarding help to succeed, got %v\n%s", err, out)
	}

	for _, command := range []string{"attach", "detach"} {
		if !strings.Contains(out, command) {
			t.Fatalf("expected port-forwarding help to include %q\n%s", command, out)
		}
	}
}

func TestPortForwardingCreateDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "create", "port-forwarding", "pf-01", "--vip", "vip-01", "--protocol", "TCP", "--vip-port", "22", "--private-port", "22", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected port-forwarding create dry-run to succeed without login, got %v\n%s", err, out)
	}

	for _, expected := range []string{`"name": "pf-01"`, `"vipUuid": "vip-01"`, `"protocolType": "TCP"`, `"vipPortStart": 22`} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
		}
	}
}

func TestPortForwardingAttachDetachDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"port-forwarding", "attach", "pf-01", "--nic", "nic-01", "--dry-run", "-o", "json"},
			expected: []string{`"rule": "pf-01"`, `"nic": "nic-01"`},
		},
		{
			args:     []string{"port-forwarding", "detach", "pf-01", "--dry-run", "-o", "json"},
			expected: []string{`"rule": "pf-01"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestDeletePortForwardingDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "delete", "port-forwarding", "pf-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected delete port-forwarding dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"resource": "port-forwarding"`) || !strings.Contains(out, `"target": "pf-01"`) {
		t.Fatalf("expected delete port-forwarding dry-run request, got\n%s", out)
	}
}

func TestLoadBalancerCommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "load-balancer", "--help")
	if err != nil {
		t.Fatalf("expected load-balancer help to succeed, got %v\n%s", err, out)
	}

	for _, command := range []string{"add-nic", "remove-nic"} {
		if !strings.Contains(out, command) {
			t.Fatalf("expected load-balancer help to include %q\n%s", command, out)
		}
	}
}

func TestLoadBalancerCreateDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"create", "load-balancer", "lb-01", "--vip", "vip-01", "--dry-run", "-o", "json"},
			expected: []string{`"name": "lb-01"`, `"vipUuid": "vip-01"`},
		},
		{
			args:     []string{"create", "load-balancer-listener", "listener-01", "--load-balancer", "lb-01", "--protocol", "TCP", "--port", "80", "--dry-run", "-o", "json"},
			expected: []string{`"loadBalancer": "lb-01"`, `"name": "listener-01"`, `"loadBalancerPort": 80`, `"protocol": "TCP"`},
		},
		{
			args:     []string{"create", "load-balancer-server-group", "group-01", "--load-balancer", "lb-01", "--dry-run", "-o", "json"},
			expected: []string{`"loadBalancer": "lb-01"`, `"name": "group-01"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:3], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestLoadBalancerNicDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"load-balancer", "add-nic", "listener-01", "--nic", "nic-01", "--dry-run", "-o", "json"},
			expected: []string{`"listener": "listener-01"`, `"nic": "nic-01"`},
		},
		{
			args:     []string{"load-balancer", "remove-nic", "listener-01", "--nic", "nic-01", "--dry-run", "-o", "json"},
			expected: []string{`"listener": "listener-01"`, `"nic": "nic-01"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestDeleteLoadBalancerDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		resource string
		target   string
	}{
		{args: []string{"delete", "load-balancer", "lb-01", "--dry-run", "-o", "json"}, resource: "load-balancer", target: "lb-01"},
		{args: []string{"delete", "load-balancer-listener", "listener-01", "--dry-run", "-o", "json"}, resource: "load-balancer-listener", target: "listener-01"},
		{args: []string{"delete", "load-balancer-server-group", "group-01", "--dry-run", "-o", "json"}, resource: "load-balancer-server-group", target: "group-01"},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		if !strings.Contains(out, fmt.Sprintf(`"resource": "%s"`, tc.resource)) || !strings.Contains(out, fmt.Sprintf(`"target": "%s"`, tc.target)) {
			t.Fatalf("expected delete load-balancer dry-run request, got\n%s", out)
		}
	}
}

func TestVmScriptCommandsAreRegistered(t *testing.T) {
	out, err := runCLI(t, "vm-script", "--help")
	if err != nil {
		t.Fatalf("expected vm-script help to succeed, got %v\n%s", err, out)
	}
	if !strings.Contains(out, "execute") {
		t.Fatalf("expected vm-script help to include execute\n%s", out)
	}

	out, err = runCLI(t, "get", "vm-script-records", "--help")
	if err != nil {
		t.Fatalf("expected get vm-script-records help to succeed, got %v\n%s", err, out)
	}
	if !strings.Contains(out, "VM script execution records") {
		t.Fatalf("expected vm-script-records help, got\n%s", out)
	}
}

func TestVmScriptCreateDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "create", "vm-script", "script-01", "--content", "echo hello", "--platform", "Linux", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected vm-script create dry-run to succeed without login, got %v\n%s", err, out)
	}

	for _, expected := range []string{`"name": "script-01"`, `"scriptContent": "echo hello"`, `"platform": "Linux"`, `"encodingType": "PlainText"`, `"scriptType": "Shell"`} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
		}
	}
}

func TestVmScriptUpdateDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "update", "vm-script", "script-01", "--content", "echo updated", "--platform", "Linux", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected vm-script update dry-run to succeed without login, got %v\n%s", err, out)
	}

	for _, expected := range []string{`"resource": "vm-script"`, `"target": "script-01"`, `"scriptContent": "echo updated"`, `"platform": "Linux"`} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
		}
	}
}

func TestVmScriptExecuteDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "vm-script", "execute", "script-01", "--instance", "vm-01", "--runtime-params", `{"name":"demo"}`, "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected vm-script execute dry-run to succeed without login, got %v\n%s", err, out)
	}

	for _, expected := range []string{`"script": "script-01"`, `"vmInstanceUuids": [`, `"vm-01"`, `"runtimeParams": "{\"name\":\"demo\"}"`} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
		}
	}
}

func TestDeleteVmScriptDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "delete", "vm-script", "script-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected delete vm-script dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"resource": "vm-script"`) || !strings.Contains(out, `"target": "script-01"`) {
		t.Fatalf("expected delete vm-script dry-run request, got\n%s", out)
	}
}

func TestInfrastructureCommandsAreRegistered(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{args: []string{"host", "--help"}, expected: []string{"add-kvm"}},
		{args: []string{"primary-storage", "--help"}, expected: []string{"add-nfs", "attach-cluster", "detach-cluster"}},
		{args: []string{"backup-storage", "--help"}, expected: []string{"add-image-store", "attach-zone", "detach-zone"}},
		{args: []string{"l2-network", "--help"}, expected: []string{"attach-cluster", "detach-cluster"}},
		{args: []string{"l3-network", "--help"}, expected: []string{"add-ip-range", "add-dns", "remove-dns", "add-host-route", "remove-host-route"}},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s help to succeed, got %v\n%s", strings.Join(tc.args[:1], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected help output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestInfrastructureCreateDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"create", "zone", "zone-01", "--default", "--dry-run", "-o", "json"},
			expected: []string{`"name": "zone-01"`, `"isDefault": true`},
		},
		{
			args:     []string{"create", "cluster", "cluster-01", "--zone", "zone-01", "--hypervisor", "KVM", "--dry-run", "-o", "json"},
			expected: []string{`"name": "cluster-01"`, `"zoneUuid": "zone-01"`, `"hypervisorType": "KVM"`},
		},
		{
			args:     []string{"create", "l2-network", "l2-01", "--zone", "zone-01", "--vlan", "100", "--dry-run", "-o", "json"},
			expected: []string{`"name": "l2-01"`, `"zoneUuid": "zone-01"`, `"vlan": 100`},
		},
		{
			args:     []string{"create", "l3-network", "l3-01", "--l2-network", "l2-01", "--dry-run", "-o", "json"},
			expected: []string{`"name": "l3-01"`, `"l2NetworkUuid": "l2-01"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:3], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestInfrastructureOperationDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"host", "add-kvm", "host-01", "--cluster", "cluster-01", "--management-ip", "10.0.0.10", "--password", "secret", "--dry-run", "-o", "json"},
			expected: []string{`"name": "host-01"`, `"clusterUuid": "cluster-01"`, `"managementIp": "10.0.0.10"`},
		},
		{
			args:     []string{"primary-storage", "add-nfs", "ps-01", "--url", "nfs://10.0.0.2/export", "--zone", "zone-01", "--dry-run", "-o", "json"},
			expected: []string{`"name": "ps-01"`, `"url": "nfs://10.0.0.2/export"`, `"zoneUuid": "zone-01"`},
		},
		{
			args:     []string{"primary-storage", "attach-cluster", "ps-01", "--cluster", "cluster-01", "--dry-run", "-o", "json"},
			expected: []string{`"resource": "primary-storage"`, `"target": "ps-01"`, `"related": "cluster-01"`},
		},
		{
			args:     []string{"backup-storage", "add-image-store", "bs-01", "--url", "/zstack_bs", "--hostname", "10.0.0.3", "--password", "secret", "--dry-run", "-o", "json"},
			expected: []string{`"name": "bs-01"`, `"hostname": "10.0.0.3"`, `"url": "/zstack_bs"`},
		},
		{
			args:     []string{"backup-storage", "attach-zone", "bs-01", "--zone", "zone-01", "--dry-run", "-o", "json"},
			expected: []string{`"resource": "backup-storage"`, `"target": "bs-01"`, `"related": "zone-01"`},
		},
		{
			args:     []string{"l2-network", "attach-cluster", "l2-01", "--cluster", "cluster-01", "--dry-run", "-o", "json"},
			expected: []string{`"resource": "l2-network"`, `"target": "l2-01"`, `"related": "cluster-01"`},
		},
		{
			args:     []string{"l3-network", "add-ip-range", "l3-01", "--name", "range-01", "--cidr", "192.168.10.0/24", "--gateway", "192.168.10.1", "--dry-run", "-o", "json"},
			expected: []string{`"resource": "l3-network"`, `"target": "l3-01"`, `"name": "range-01"`, `"networkCidr": "192.168.10.0/24"`},
		},
		{
			args:     []string{"l3-network", "add-dns", "l3-01", "--dns", "223.5.5.5", "--dry-run", "-o", "json"},
			expected: []string{`"resource": "l3-network"`, `"target": "l3-01"`, `"related": "223.5.5.5"`},
		},
		{
			args:     []string{"l3-network", "add-host-route", "l3-01", "--prefix", "10.10.0.0/16", "--nexthop", "192.168.10.1", "--dry-run", "-o", "json"},
			expected: []string{`"resource": "l3-network"`, `"target": "l3-01"`, `"prefix": "10.10.0.0/16"`, `"nexthop": "192.168.10.1"`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestInfrastructureUpdateDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{args: []string{"update", "zone", "zone-01", "--name", "zone-renamed", "--dry-run", "-o", "json"}, expected: []string{`"resource": "zone"`, `"name": "zone-renamed"`}},
		{args: []string{"update", "cluster", "cluster-01", "--description", "updated", "--dry-run", "-o", "json"}, expected: []string{`"resource": "cluster"`, `"description": "updated"`}},
		{args: []string{"update", "host", "host-01", "--management-ip", "10.0.0.11", "--dry-run", "-o", "json"}, expected: []string{`"resource": "host"`, `"managementIp": "10.0.0.11"`}},
		{args: []string{"update", "primary-storage", "ps-01", "--url", "nfs://10.0.0.2/new", "--dry-run", "-o", "json"}, expected: []string{`"resource": "primary-storage"`, `"url": "nfs://10.0.0.2/new"`}},
		{args: []string{"update", "backup-storage", "bs-01", "--name", "bs-renamed", "--dry-run", "-o", "json"}, expected: []string{`"resource": "backup-storage"`, `"name": "bs-renamed"`}},
		{args: []string{"update", "l2-network", "l2-01", "--description", "updated", "--dry-run", "-o", "json"}, expected: []string{`"resource": "l2-network"`, `"description": "updated"`}},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestInfrastructureDeleteDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		resource string
		target   string
	}{
		{args: []string{"delete", "zone", "zone-01", "--dry-run", "-o", "json"}, resource: "zone", target: "zone-01"},
		{args: []string{"delete", "cluster", "cluster-01", "--dry-run", "-o", "json"}, resource: "cluster", target: "cluster-01"},
		{args: []string{"delete", "host", "host-01", "--dry-run", "-o", "json"}, resource: "host", target: "host-01"},
		{args: []string{"delete", "primary-storage", "ps-01", "--dry-run", "-o", "json"}, resource: "primary-storage", target: "ps-01"},
		{args: []string{"delete", "backup-storage", "bs-01", "--dry-run", "-o", "json"}, resource: "backup-storage", target: "bs-01"},
		{args: []string{"delete", "l2-network", "l2-01", "--dry-run", "-o", "json"}, resource: "l2-network", target: "l2-01"},
		{args: []string{"delete", "l3-network", "l3-01", "--dry-run", "-o", "json"}, resource: "l3-network", target: "l3-01"},
		{args: []string{"delete", "ip-range", "range-01", "--dry-run", "-o", "json"}, resource: "ip-range", target: "range-01"},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		if !strings.Contains(out, fmt.Sprintf(`"resource": "%s"`, tc.resource)) || !strings.Contains(out, fmt.Sprintf(`"target": "%s"`, tc.target)) {
			t.Fatalf("expected infrastructure delete dry-run request, got\n%s", out)
		}
	}
}

func TestImageTemplateCommandsAreRegistered(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{args: []string{"image", "--help"}, expected: []string{"recover", "sync", "create-from-volume", "create-from-snapshot"}},
		{args: []string{"image-group", "--help"}, expected: []string{"create-from-image"}},
		{args: []string{"get", "image-groups", "--help"}, expected: []string{"List image groups"}},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s help to succeed, got %v\n%s", strings.Join(tc.args[:1], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected help output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestImageTemplateDryRunsDoNotRequireLogin(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{
			args:     []string{"image", "recover", "image-01", "--backup-storage", "bs-01", "--dry-run", "-o", "json"},
			expected: []string{`"image": "image-01"`, `"backupStorageUuids": [`},
		},
		{
			args:     []string{"image", "sync", "image-01", "--image-storage", "bs-01", "--dry-run", "-o", "json"},
			expected: []string{`"image": "image-01"`, `"storage": "bs-01"`},
		},
		{
			args:     []string{"image", "create-from-volume", "volume-01", "--name", "tmpl-01", "--backup-storage", "bs-01", "--dry-run", "-o", "json"},
			expected: []string{`"source": "volume-01"`, `"name": "tmpl-01"`, `"backupStorageUuids": [`},
		},
		{
			args:     []string{"image", "create-from-snapshot", "snap-01", "--name", "tmpl-02", "--dry-run", "-o", "json"},
			expected: []string{`"source": "snap-01"`, `"name": "tmpl-02"`},
		},
		{
			args:     []string{"image-group", "create-from-image", "image-01", "--name", "group-01", "--data-volume-template", "tmpl-01", "--dry-run", "-o", "json"},
			expected: []string{`"image": "image-01"`, `"name": "group-01"`, `"dataVolumeTemplateUuids": [`},
		},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s dry-run to succeed without login, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestExpungeImageGroupDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "expunge", "image-group", "group-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected expunge image-group dry-run to succeed without login, got %v\n%s", err, out)
	}

	if !strings.Contains(out, `"imageGroup": "group-01"`) {
		t.Fatalf("expected expunge image-group dry-run request, got\n%s", out)
	}
}

func TestP3OpsCommandsAreRegistered(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{args: []string{"global-config", "--help"}, expected: []string{"set"}},
		{args: []string{"get", "alarms", "--help"}, expected: []string{"List alarms"}},
		{args: []string{"get", "events", "--help"}, expected: []string{"List event records"}},
		{args: []string{"get", "audits", "--help"}, expected: []string{"List audit records"}},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s help to succeed, got %v\n%s", strings.Join(tc.args[:1], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected help output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestP3IAMReadCommandsAreRegistered(t *testing.T) {
	cases := []struct {
		args     []string
		expected []string
	}{
		{args: []string{"get", "accounts", "--help"}, expected: []string{"List ZStack accounts"}},
		{args: []string{"get", "users", "--help"}, expected: []string{"List ZStack account users"}},
		{args: []string{"get", "projects", "--help"}, expected: []string{"List ZStack IAM2 projects"}},
		{args: []string{"get", "organizations", "--help"}, expected: []string{"List ZStack IAM2 organizations"}},
		{args: []string{"get", "roles", "--help"}, expected: []string{"List ZStack IAM roles"}},
		{args: []string{"get", "policies", "--help"}, expected: []string{"List ZStack IAM policies"}},
	}

	for _, tc := range cases {
		out, err := runCLI(t, tc.args...)
		if err != nil {
			t.Fatalf("expected %s help to succeed, got %v\n%s", strings.Join(tc.args[:2], " "), err, out)
		}
		for _, expected := range tc.expected {
			if !strings.Contains(out, expected) {
				t.Fatalf("expected help output to contain %s, got\n%s", expected, out)
			}
		}
	}
}

func TestGlobalConfigSetDryRunDoesNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "global-config", "set", "quota.vm.num", "--value", "100", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected global-config set dry-run to succeed without login, got %v\n%s", err, out)
	}

	for _, expected := range []string{`"category": "quota"`, `"name": "vm.num"`, `"value": "100"`} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected dry-run output to contain %s, got\n%s", expected, out)
		}
	}
}

func TestAlarmDryRunsDoNotRequireLogin(t *testing.T) {
	out, err := runCLI(t, "create", "alarm", "alarm-01", "--namespace", "ZStack/VM", "--metric", "CPUUsedUtilization", "--operator", "GreaterThanThreshold", "--threshold", "80", "--label", "VMUuid=vm-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected create alarm dry-run to succeed without login, got %v\n%s", err, out)
	}
	for _, expected := range []string{`"name": "alarm-01"`, `"namespace": "ZStack/VM"`, `"metricName": "CPUUsedUtilization"`, `"threshold": 80`, `"key": "VMUuid"`} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected create alarm dry-run output to contain %s, got\n%s", expected, out)
		}
	}

	out, err = runCLI(t, "delete", "alarm", "alarm-01", "--dry-run", "-o", "json")
	if err != nil {
		t.Fatalf("expected delete alarm dry-run to succeed without login, got %v\n%s", err, out)
	}
	if !strings.Contains(out, `"resource": "alarm"`) || !strings.Contains(out, `"target": "alarm-01"`) {
		t.Fatalf("expected delete alarm dry-run request, got\n%s", out)
	}
}
