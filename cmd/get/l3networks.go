package get

import (
	"fmt"

	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

// 定义格式化的三层网络结构体
type FormattedL3Network struct {
	Name          string `json:"name" yaml:"name" header:"NAME"`
	UUID          string `json:"uuid" yaml:"uuid" header:"UUID"`
	Type          string `json:"type" yaml:"type" header:"TYPE"`
	State         string `json:"state" yaml:"state" header:"STATE"`
	L2NetworkUuid string `json:"l2NetworkUuid" yaml:"l2NetworkUuid" header:"L2 NETWORK UUID"`
	ZoneUuid      string `json:"zoneUuid" yaml:"zoneUuid" header:"ZONE UUID"`
	IpVersion     int    `json:"ipVersion" yaml:"ipVersion" header:"IP VERSION"`
	DnsDomain     string `json:"dnsDomain" yaml:"dnsDomain" header:"DNS DOMAIN"`
	Dns           string `json:"dns" yaml:"dns" header:"DNS"`
	System        bool   `json:"system" yaml:"system" header:"SYSTEM"`
	Category      string `json:"category" yaml:"category" header:"CATEGORY"`
	EnableIPAM    bool   `json:"enableIPAM" yaml:"enableIPAM" header:"ENABLE IPAM"`
}

// l3NetworksCmd 表示 l3-networks 命令
var l3NetworksCmd = &cobra.Command{
	Use:   "l3-networks [name]",
	Short: "List L3 networks",
	Long:  `List all Layer 3 networks in the ZStack cloud platform.`,
	Run: func(cobraCmd *cobra.Command, args []string) {
		// 1. 创建客户端
		zsClient := client.GetClient()
		if zsClient == nil {
			fmt.Println("Error: Not logged in. Please run 'zstack-cli login' first.")
			return
		}

		// 2. 创建查询参数
		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			fmt.Printf("Error building query parameters: %s\n", err)
			return
		}

		// 3. 调用 API
		l3Networks, err := zsClient.QueryL3Network(*queryParam)
		if err != nil {
			fmt.Printf("Error querying L3 networks: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedL3Network
		for _, network := range l3Networks {
			formatted := FormattedL3Network{
				Name:          network.Name,
				UUID:          network.UUID,
				Type:          network.Type,
				State:         network.State,
				L2NetworkUuid: network.L2NetworkUuid,
				ZoneUuid:      network.ZoneUuid,
				IpVersion:     network.IpVersion,
				DnsDomain:     network.DnsDomain,
				Dns:           strings.Join(network.Dns, ","),
				System:        network.System,
				Category:      network.Category,
				EnableIPAM:    network.EnableIPAM,
			}
			formattedResults = append(formattedResults, formatted)
		}

		// 使用 output 包进行输出
		err = output.PrintWithFields(formattedResults, format, fields)
		if err != nil {
			fmt.Printf("Error formatting output: %s\n", err)
			return
		}
	},
}

func init() {
	GetCmd.AddCommand(l3NetworksCmd)

	common.AddQueryFlags(l3NetworksCmd)
	//l3NetworksCmd.Flags().StringSlice("fields", nil, "Fields to display (comma-separated without spaces, e.g. name,uuid,type)")
}
