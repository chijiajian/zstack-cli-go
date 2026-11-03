package skillinstall

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"

	"github.com/chijiajian/zstack-cli-go/skills"
	"gopkg.in/yaml.v3"
)

// ExportWorkBuddy adapts marketplace metadata while retaining the shared workflow.
func ExportWorkBuddy(path string) error {
	source, err := fs.Sub(skills.Files, Name)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	err = fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		if path == "SKILL.md" {
			parts := bytes.SplitN(data, []byte("\n---\n"), 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid Skill metadata")
			}
			var meta map[string]interface{}
			if err = yaml.Unmarshal(bytes.TrimPrefix(parts[0], []byte("---\n")), &meta); err != nil {
				return err
			}
			version, ok := meta["metadata"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("missing Skill content version")
			}
			meta["version"] = version["version"]
			meta["author"] = "ZStack CLI contributors"
			meta["display_name"] = "ZStack 运维"
			meta["display_name_en"] = "ZStack Operations"
			meta["description_zh"] = meta["description"]
			meta["description_en"] = "Operate ZStack through zstack-cli for queries, monitoring and resource lifecycle tasks."
			front, err := yaml.Marshal(meta)
			if err != nil {
				return err
			}
			data = append(append([]byte("---\n"), front...), append([]byte("---\n"), parts[1]...)...)
		}
		entry, err := archive.Create(Name + "/" + path)
		if err != nil {
			return err
		}
		_, err = entry.Write(data)
		return err
	})
	if err != nil {
		return err
	}
	if err = archive.Close(); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(buffer.Bytes())
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		os.Remove(path)
		if writeErr != nil {
			return writeErr
		}
		return closeErr
	}
	return nil
}
