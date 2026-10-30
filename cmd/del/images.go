// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// cmd/del/images.go
package del

import (
	"errors"
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

var DeleteImagesCmd = &cobra.Command{
	Use:     "image [name]",
	Aliases: []string{"images"},
	Short:   "Delete image(s) by name",
	Long: `Delete one or more images by name. ,
Examples:
  # Delete a single image
  zstack-cli delete image my-image

  # Delete multiple images (same name matched)
  zstack-cli delete image my-image`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error { return deleteImage(cmd, args[0]) },
}

func init() {
	DeleteImagesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	DeleteCmd.AddCommand(DeleteImagesCmd)
}

func deleteImage(cmd *cobra.Command, nameOrUUID string) error {
	zsClient := client.GetClient()
	if zsClient == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}

	images, err := client.GetReadyImagesByNameOrUUID(zsClient, nameOrUUID)
	if err != nil {
		return fmt.Errorf("failed to find image: %s", err)
	}

	if len(images) == 0 {
		fmt.Printf("no Ready images found with name or UUID: %s\n", nameOrUUID)
		return nil
	}

	fmt.Println("The following images will be deleted:")
	for _, img := range images {
		fmt.Printf("- %s (%s)\n", img.Name, img.UUID)
	}

	if dryRunFlag {
		fmt.Println("Dry-run: no resources were changed.")
		return nil
	}

	confirmed, err := common.Confirm(cmd, "Are you sure you want to delete these images? (yes/No): ")
	if err != nil {
		return err
	}
	if !confirmed {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Aborted. No images were deleted.")
		return err
	}

	var failures []error
	for _, img := range images {
		if dryRunFlag {
			fmt.Printf("[Dry-run] Would delete image: %s (%s)\n", img.Name, img.UUID)
			continue
		}

		if err := zsClient.DeleteImage(img.UUID, param.DeleteModePermissive); err != nil {
			failures = append(failures, fmt.Errorf("Failed to delete image %s (%s): %w", img.Name, img.UUID, err))
			continue
		}
		fmt.Printf("Deleted image: %s (%s)\n", img.Name, img.UUID)

	}
	return errors.Join(failures...)
}
