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

package get

import (
	"fmt"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type FormattedImage struct {
	Name            string `json:"name" yaml:"name" header:"NAME"`
	UUID            string `json:"uuid" yaml:"uuid" header:"UUID"`
	State           string `json:"state" yaml:"state" header:"STATE"`
	Status          string `json:"status" yaml:"status" header:"STATUS"`
	Size            string `json:"size" yaml:"size" header:"SIZE"`
	SizeBytes       *int64 `json:"sizeBytes" yaml:"sizeBytes" header:"SIZE BYTES" table:"opt-in"`
	ActualSize      string `json:"actualSize" yaml:"actualSize" header:"ACTUAL SIZE"`
	ActualSizeBytes *int64 `json:"actualSizeBytes" yaml:"actualSizeBytes" header:"ACTUAL SIZE BYTES" table:"opt-in"`
	Format          string `json:"format" yaml:"format" header:"FORMAT"`
	MediaType       string `json:"mediaType" yaml:"mediaType" header:"MEDIA TYPE"`
	Platform        string `json:"platform" yaml:"platform" header:"PLATFORM"`
	Architecture    string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	Type            string `json:"type" yaml:"type" header:"TYPE"`
	GuestOsType     string `json:"guestOsType" yaml:"guestOsType" header:"GUEST OS TYPE"`
}

type imageQueryInventory struct {
	view.ImageInventoryView
	SizeBytes       *int64 `json:"size"`
	ActualSizeBytes *int64 `json:"actualSize"`
}

var imagesCmd = &cobra.Command{
	Use:   "images [name]",
	Short: "List images",
	Long: `List all images in the ZStack cloud platform.
Structured output includes sizeBytes, actualSizeBytes, or null if not reported.`,
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[imageQueryInventory](cobraCmd, queryParam, "v1/images")
		images := result.Items
		if err != nil {
			return fmt.Errorf("Error querying images: %w", err)
		}

		var formattedResults []FormattedImage
		for _, image := range images {
			formatted := FormattedImage{
				Name:            image.Name,
				UUID:            image.UUID,
				State:           image.State,
				Status:          image.Status,
				Size:            utils.FormatDiskSize(inventoryNumber(image.SizeBytes)),
				SizeBytes:       image.SizeBytes,
				ActualSize:      utils.FormatDiskSize(inventoryNumber(image.ActualSizeBytes)),
				ActualSizeBytes: image.ActualSizeBytes,
				Format:          image.Format,
				MediaType:       image.MediaType,
				Platform:        image.Platform,
				Architecture:    string(image.Architecture),
				Type:            image.Type,
				GuestOsType:     image.GuestOsType,
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
	GetCmd.AddCommand(imagesCmd)

	common.AddQueryFlags(imagesCmd)
}
