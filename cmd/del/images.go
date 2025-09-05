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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
)

var DeleteImagesCmd = &cobra.Command{
	Use:   "image [name]",
	Short: "Delete image(s) by name",
	Long: `Delete one or more images by name. ,
Examples:
  # Delete a single image
  zstack-cli delete image my-image

  # Delete multiple images (same name matched)
  zstack-cli delete image my-image`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Error: image name is required")
			return
		}
		if err := deleteImage(cmd, args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
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
		fmt.Println("no Ready images found with name or UUID: %s", nameOrUUID)
		return nil
	}

	fmt.Println("The following images will be deleted:")
	for _, img := range images {
		fmt.Printf("- %s (%s)\n", img.Name, img.UUID)
	}

	fmt.Print("Are you sure you want to delete these images? (yes/No): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input != "yes" && input != "y" {
		fmt.Printf("Aborted. No images were deleted.\n")
		return nil
	}

	for _, img := range images {
		if dryRunFlag {
			fmt.Printf("[Dry-run] Would delete image: %s (%s)\n", img.Name, img.UUID)
			continue
		}

		if err := zsClient.DeleteImage(img.UUID, param.DeleteModePermissive); err != nil {
			fmt.Printf("Failed to delete image %s (%s): %s\n", img.Name, img.UUID, err)
		}
		fmt.Printf("Deleted image: %s (%s)\n", img.Name, img.UUID)

	}
	return nil
}
