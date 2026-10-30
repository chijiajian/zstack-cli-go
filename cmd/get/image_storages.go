/*
Copyright © 2025 zstack.io
*/
package get

import (
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
)

type FormattedBackupStorage struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	URL               string `json:"url" yaml:"url" header:"URL"`
	Type              string `json:"type" yaml:"type" header:"TYPE"`
	State             string `json:"state" yaml:"state" header:"STATE"`
	Status            string `json:"status" yaml:"status" header:"STATUS"`
	TotalCapacity     string `json:"totalCapacity" yaml:"totalCapacity" header:"TOTAL CAPACITY"`
	AvailableCapacity string `json:"availableCapacity" yaml:"availableCapacity" header:"AVAILABLE CAPACITY"`
}

func formatCapacity(bytes int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
		PB
	)

	switch {
	case bytes >= int64(PB):
		return fmt.Sprintf("%.2f PB", float64(bytes)/PB)
	case bytes >= int64(TB):
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= int64(GB):
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= int64(MB):
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= int64(KB):
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

var imageStoragesCmd = &cobra.Command{
	Use:     "image-storages [name]",
	Aliases: []string{"image-storage", "backup-storage", "backup-storages"},
	Short:   "Get ZStack image storages (backup storages)",
	Long: `Display one or many ZStack image storages (backup storages).

Examples:
  # List all image storages
  zstack-cli get image-storages

  # Get a specific image storage by name
  zstack-cli get image-storages my-image-storage

  # Query image storages with specific conditions
  zstack-cli get image-storages --q "name=ImageStore1"

  # Query image storages with multiple conditions
  zstack-cli get image-storages --q "name=ImageStore1" --q "state=Enabled"

  # Limit the number of results and paginate
  zstack-cli get image-storages --limit 10 --start 0

  # Output in different formats
  zstack-cli get image-storages --output json
  zstack-cli get image-storages --output yaml
  zstack-cli get image-storages --output text`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.BackupStorageInventoryView](cobraCmd, queryParam, "v1/backup-storage")
		backupStorages := result.Items

		if err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		var formattedResults []FormattedBackupStorage
		for _, storage := range backupStorages {
			formatted := FormattedBackupStorage{
				Name:              storage.Name,
				UUID:              storage.UUID,
				URL:               storage.Url,
				Type:              storage.Type,
				State:             storage.State,
				Status:            storage.Status,
				TotalCapacity:     formatCapacity(storage.TotalCapacity),
				AvailableCapacity: formatCapacity(storage.AvailableCapacity),
			}
			formattedResults = append(formattedResults, formatted)
		}

		err = common.PrintQuery(cobraCmd, formattedResults, result.Total)
		if err != nil {
			return fmt.Errorf("Error formatting output: %w", err)
		}
		return nil
	},
}

func init() {
	GetCmd.AddCommand(imageStoragesCmd)

	common.AddQueryFlags(imageStoragesCmd)
}
