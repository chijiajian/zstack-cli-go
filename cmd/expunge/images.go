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

// cmd/expunge/images.go
package expunge

import (
	"errors"
	"fmt"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
)

var ExpungeImagesCmd = &cobra.Command{
	Use:     "image [name]",
	Aliases: []string{"images"},
	Short:   "Expunge image(s) by name",
	Long: `Expunge one or more images by name. ,
Examples:
  # Expunge a single image
  zstack-cli expunge image my-image

  # Expunge multiple images (same name matched)
  zstack-cli expunge image my-image`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error { return expungeImage(cmd, args[0]) },
}

func init() {
	ExpungeImagesCmd.Flags().BoolP("yes", "y", false, "Automatic yes to confirmation prompts")
	ExpungeCmd.AddCommand(ExpungeImagesCmd)
}

func expungeImage(cmd *cobra.Command, nameOrUUID string) error {
	zsClient := client.GetClient()
	if zsClient == nil {
		return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
	}

	images, err := client.GetDeletedImagesByNameOrUUID(zsClient, nameOrUUID)
	if err != nil {
		return fmt.Errorf("failed to find image: %s", err)
	}

	if len(images) == 0 {
		fmt.Printf("no Deleted images found with name or UUID: %s\n", nameOrUUID)
		return nil
	}

	fmt.Println("The following images will be expunged:")
	for _, img := range images {
		fmt.Printf("- %s (%s)\n", img.Name, img.UUID)
	}

	if dryRunFlag {
		fmt.Println("Dry-run: no resources were changed.")
		return nil
	}

	confirmed, err := common.Confirm(cmd, "Are you sure you want to expunge these images? (yes/no): ")
	if err != nil {
		return err
	}
	if !confirmed {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "Aborted. No images were expunged.")
		return err
	}

	var failures []error
	for _, img := range images {
		if dryRunFlag {
			fmt.Printf("[Dry-run] Would expunge image: %s (%s)\n", img.Name, img.UUID)
			continue
		}

		if err := zsClient.ExpungeImage(img.UUID); err != nil {
			failures = append(failures, fmt.Errorf("Failed to expunge image %s (%s): %w", img.Name, img.UUID, err))
			continue
		}
		fmt.Printf("Expunged image: %s (%s)\n", img.Name, img.UUID)

	}
	return errors.Join(failures...)
}
