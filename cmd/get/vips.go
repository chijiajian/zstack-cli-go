package get

import (
	"fmt"
	"strings"

	"github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/chijiajian/zstack-cli-go/pkg/output"
	"github.com/spf13/cobra"
)

// 定义格式化的虚拟IP结构体
type FormattedVip struct {
	Name               string `json:"name" yaml:"name" header:"NAME"`
	UUID               string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description        string `json:"description" yaml:"description" header:"DESCRIPTION"`
	L3NetworkUUID      string `json:"l3NetworkUuid" yaml:"l3NetworkUuid" header:"L3 NETWORK UUID"`
	Ip                 string `json:"ip" yaml:"ip" header:"IP"`
	State              string `json:"state" yaml:"state" header:"STATE"`
	Gateway            string `json:"gateway" yaml:"gateway" header:"GATEWAY"`
	Netmask            string `json:"netmask" yaml:"netmask" header:"NETMASK"`
	PrefixLen          string `json:"prefixLen" yaml:"prefixLen" header:"PREFIX LENGTH"`
	ServiceProvider    string `json:"serviceProvider" yaml:"serviceProvider" header:"SERVICE PROVIDER"`
	PeerL3NetworkUuids string `json:"peerL3NetworkUuids" yaml:"peerL3NetworkUuids" header:"PEER L3 NETWORK UUIDS"`
	UseFor             string `json:"useFor" yaml:"useFor" header:"USE FOR"`
	System             bool   `json:"system" yaml:"system" header:"SYSTEM"`
	ServicesTypes      string `json:"servicesTypes" yaml:"servicesTypes" header:"SERVICES TYPES"`
}

// vipsCmd 表示 vips 命令
var vipsCmd = &cobra.Command{
	Use:   "vips [name]",
	Short: "List virtual IPs",
	Long:  `List all virtual IPs in the ZStack cloud platform.`,
	Args:  cobra.MaximumNArgs(1),
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
		vips, err := zsClient.QueryVip(*queryParam)
		if err != nil {
			fmt.Printf("Error querying virtual IPs: %s\n", err)
			return
		}

		// 4. 格式化输出
		outputFormat, _ := cobraCmd.Flags().GetString("output")
		format := output.ParseFormat(outputFormat)

		fields, _ := cobraCmd.Flags().GetStringSlice("fields")

		// 准备格式化的数据
		var formattedResults []FormattedVip
		for _, vip := range vips {
			// 收集服务类型
			serviceTypes := make([]string, 0, len(vip.ServicesRefs))
			for _, ref := range vip.ServicesRefs {
				serviceTypes = append(serviceTypes, ref.ServiceType)
			}

			formatted := FormattedVip{
				Name:        vip.Name,
				UUID:        vip.UUID,
				Description: vip.Description,

				L3NetworkUUID:      vip.L3NetworkUUID,
				Ip:                 vip.Ip,
				State:              vip.State,
				Gateway:            vip.Gateway,
				Netmask:            vip.Netmask,
				PrefixLen:          vip.PrefixLen,
				ServiceProvider:    vip.ServiceProvider,
				PeerL3NetworkUuids: vip.PeerL3NetworkUuids,
				UseFor:             vip.UseFor,
				System:             vip.System,
				ServicesTypes:      strings.Join(serviceTypes, ","),
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
	GetCmd.AddCommand(vipsCmd)
	common.AddQueryFlags(vipsCmd)
}
