package skillinstall

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTargetsAndInstall(t *testing.T) {
	root := t.TempDir()
	for agent, dirs := range Directories {
		for i, scope := range []string{"user", "project"} {
			got, err := Target(agent, scope, root, "", root)
			want := filepath.Join(root, filepath.FromSlash(dirs[i]), Name)
			if err != nil || got != want {
				t.Fatalf("%s %s %s %v", agent, scope, got, err)
			}
		}
	}
	if _, err := Target("unknown", "user", "", "", root); err == nil {
		t.Fatal("accepted unknown agent")
	}
	if _, err := Target("codex", "project", "", "", root); err == nil {
		t.Fatal("missing project accepted")
	}
	target := filepath.Join(root, "install", Name)
	if _, err := Install(target, false, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(target)); !os.IsNotExist(err) {
		t.Fatal("preview created directory")
	}
	if _, err := Install(target, false, false); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(target, "SKILL.md"))
	if err != nil || len(content) == 0 {
		t.Fatal("missing skill")
	}
	if _, err := os.Stat(filepath.Join(target, "scripts", "check_cli.py")); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(target, false, false); err == nil {
		t.Fatal("overwrote existing install")
	}
	marker := filepath.Join(target, "local-customization")
	os.WriteFile(marker, []byte("preserve"), 0600)
	result, err := Install(target, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(result.Backup, "local-customization")); err != nil || string(data) != "preserve" {
		t.Fatal("backup lost")
	}
	if _, err := os.Stat(target + ".install.lock"); !os.IsNotExist(err) {
		t.Fatal("lock leaked")
	}
	lock, err := os.Create(target + ".install.lock")
	if err != nil {
		t.Fatal(err)
	}
	lock.Close()
	if _, err := Install(target, true, false); err == nil {
		t.Fatal("concurrent install accepted")
	}
}
