package skillinstall

import (
	"archive/zip"
	"bytes"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkBuddyExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "skill bundle.zip")
	if err := ExportWorkBuddy(path); err != nil {
		t.Fatal(err)
	}
	original, _ := os.ReadFile(path)
	if err := ExportWorkBuddy(path); err == nil {
		t.Fatal("overwrote existing archive")
	}
	current, _ := os.ReadFile(path)
	if !bytes.Equal(current, original) {
		t.Fatal("existing archive changed")
	}
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	names := map[string]bool{}
	for _, f := range archive.File {
		names[f.Name] = true
		if f.Name != Name+"/SKILL.md" {
			continue
		}
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(r)
		r.Close()
		parts := bytes.SplitN(data, []byte("\n---\n"), 2)
		if len(parts) != 2 {
			t.Fatal("invalid frontmatter")
		}
		var meta map[string]interface{}
		if err = yaml.Unmarshal(bytes.TrimPrefix(parts[0], []byte("---\n")), &meta); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"name", "display_name", "display_name_en", "description", "description_zh", "description_en", "author", "version"} {
			if meta[key] == nil || meta[key] == "" {
				t.Errorf("missing %s", key)
			}
		}
		if !bytes.Contains(parts[1], []byte("zstack-cli")) {
			t.Fatal("workflow lost")
		}
	}
	for _, file := range []string{"SKILL.md", "references/queries.md", "scripts/check_cli.py"} {
		if !names[Name+"/"+file] {
			t.Error("missing", file)
		}
	}
}
func TestAdditionalAgentTargets(t *testing.T) {
	home := t.TempDir()
	for _, tc := range []struct{ agent, scope, dest, want string }{
		{"deepcode", "user", "", filepath.Join(home, ".agents", "skills", Name)},
		{"deepcode", "project", "", filepath.Join(home, ".deepcode", "skills", Name)},
		{"custom", "user", filepath.Join(home, "other agent"), filepath.Join(home, "other agent", Name)},
	} {
		got, err := Target(tc.agent, tc.scope, home, tc.dest, home)
		if err != nil || got != tc.want {
			t.Fatalf("%s got %s %v", tc.agent, got, err)
		}
	}
	if _, err := Target("custom", "user", "", "", home); err == nil {
		t.Fatal("custom agent without destination accepted")
	}
}

func TestWorkBuddyManifestLineEndings(t *testing.T) {
	source := "---\nname: zstack-operations\ndescription: test\nmetadata:\n  version: 1.2.0\n---\n# Instructions\nRun zstack-cli.\n"
	unix, err := workBuddyManifest([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	windows, err := workBuddyManifest(bytes.ReplaceAll([]byte(source), []byte("\n"), []byte("\r\n")))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(unix, windows) {
		t.Fatal("line endings changed exported metadata or instructions")
	}
}
