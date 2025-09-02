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

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
)

type FormattedImage struct {
	Name         string `json:"name" yaml:"name" header:"NAME"`
	UUID         string `json:"uuid" yaml:"uuid" header:"UUID"`
	State        string `json:"state" yaml:"state" header:"STATE"`
	Status       string `json:"status" yaml:"status" header:"STATUS"`
	Size         string `json:"size" yaml:"size" header:"SIZE"`
	ActualSize   string `json:"actualSize" yaml:"actualSize" header:"ACTUAL SIZE"`
	Format       string `json:"format" yaml:"format" header:"FORMAT"`
	MediaType    string `json:"mediaType" yaml:"mediaType" header:"MEDIA TYPE"`
	Platform     string `json:"platform" yaml:"platform" header:"PLATFORM"`
	Architecture string `json:"architecture" yaml:"architecture" header:"ARCHITECTURE"`
	Type         string `json:"type" yaml:"type" header:"TYPE"`
	GuestOsType  string `json:"guestOsType" yaml:"guestOsType" header:"GUEST OS TYPE"`
}

var imagesCmd = &cobra.Command{
	Use:   "images [name]",
	Short: "List images",
	Long:  `List all images in the ZStack cloud platform.`,
	Run: func(cobraCmd *cobra.Command, args []string) {

		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			fmt.Printf("Error building query parameters: %s\n", err)
			return
		}

		images, err := zsClient.QueryImage(*queryParam)
		if err != nil {
			fmt.Printf("Error querying images: %s\n", err)
			return
		}

		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := utils.ParseFormat(outputFormat)
		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		var formattedResults []FormattedImage
		for _, image := range images {
			formatted := FormattedImage{
				Name:         image.Name,
				UUID:         image.UUID,
				State:        image.State,
				Status:       image.Status,
				Size:         utils.FormatDiskSize(image.Size),
				ActualSize:   utils.FormatDiskSize(image.ActualSize),
				Format:       image.Format,
				MediaType:    image.MediaType,
				Platform:     image.Platform,
				Architecture: string(image.Architecture),
				Type:         image.Type,
				GuestOsType:  image.GuestOsType,
			}
			formattedResults = append(formattedResults, formatted)
		}

		err = utils.PrintWithFields(formattedResults, format, fields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
			return
		}
	},
}

func init() {
	GetCmd.AddCommand(imagesCmd)

	common.AddQueryFlags(imagesCmd)
}
