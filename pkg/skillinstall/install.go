package skillinstall

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/chijiajian/zstack-cli-go/skills"
)

const Name = "zstack-operations"

var Directories = map[string][2]string{
	"deepcode":    {".agents/skills", ".deepcode/skills"},
	"custom":      {"", ""},
	"codex":       {".agents/skills", ".agents/skills"},
	"claude-code": {".claude/skills", ".claude/skills"},
	"cursor":      {".cursor/skills", ".cursor/skills"},
	"copilot":     {".copilot/skills", ".github/skills"},
}

type Result struct {
	Destination string `json:"destination"`
	Backup      string `json:"backup,omitempty"`
	DryRun      bool   `json:"dryRun"`
}

func Target(agent, scope, project, dest, home string) (string, error) {
	dirs, ok := Directories[agent]
	if !ok {
		return "", fmt.Errorf("unknown agent %q", agent)
	}
	if agent == "custom" && dest == "" {
		return "", fmt.Errorf("custom agent requires --dest pointing to its Skill parent directory")
	}
	if scope != "user" && scope != "project" {
		return "", fmt.Errorf("scope must be user or project")
	}
	if dest != "" {
		return filepath.Abs(filepath.Join(dest, Name))
	}
	base, dir := home, dirs[0]
	if scope == "project" {
		if project == "" {
			return "", fmt.Errorf("project scope requires --project")
		}
		info, err := os.Stat(project)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("project must be an existing directory")
		}
		base, dir = project, dirs[1]
	}
	return filepath.Abs(filepath.Join(base, filepath.FromSlash(dir), Name))
}
func Install(target string, replace, dryRun bool) (Result, error) {
	result := Result{Destination: target, DryRun: dryRun}
	check := func() error {
		info, err := os.Lstat(target)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("destination must be a directory, not a symlink")
		}
		if !replace {
			return fmt.Errorf("destination exists; use --replace to update with a backup")
		}
		return nil
	}
	if err := check(); err != nil {
		return result, err
	}
	if dryRun {
		return result, nil
	}
	parent := filepath.Dir(target)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return result, err
	}
	lock := target + ".install.lock"
	f, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return result, err
	}
	f.Close()
	defer os.Remove(lock)
	if err := check(); err != nil {
		return result, err
	}
	stage, err := os.MkdirTemp(parent, ".zstack-skill-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(stage)
	source, err := fs.Sub(skills.Files, Name)
	if err != nil {
		return result, err
	}
	err = fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		destination := filepath.Join(stage, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(destination, 0700)
		}
		data, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, data, 0600)
	})
	if err != nil {
		return result, err
	}
	if _, err := os.Lstat(target); err == nil {
		backup, err := os.MkdirTemp(parent, Name+".backup-")
		if err != nil {
			return result, err
		}
		if err = os.Remove(backup); err != nil {
			return result, err
		}
		if err = os.Rename(target, backup); err != nil {
			return result, err
		}
		result.Backup = backup
	}
	if err = os.Rename(stage, target); err != nil {
		if result.Backup != "" {
			if rollbackErr := os.Rename(result.Backup, target); rollbackErr != nil {
				return result, fmt.Errorf("install failed: %v; restore backup %s: %w", err, result.Backup, rollbackErr)
			}
		}
		return result, err
	}
	return result, nil
}
