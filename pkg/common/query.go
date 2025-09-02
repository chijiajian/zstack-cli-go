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

package common

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/spf13/cobra"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/client"
	"github.com/terraform-zstack-modules/zstack-sdk-go/pkg/param"
)

var (
	zsClient    *client.ZSClient
	clientMutex sync.Mutex
)

func SetClient(client *client.ZSClient) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	zsClient = client
}

func GetClient() *client.ZSClient {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	return zsClient
}

func AddQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("q", "q", []string{}, "Query condition, can be specified multiple times")
	cmd.Flags().IntP("limit", "l", 0, "Maximum number of results to return")
	cmd.Flags().IntP("start", "s", 0, "Starting index for results")
	cmd.Flags().Bool("count", false, "Return count of matching records only")
	cmd.Flags().Bool("reply-with-count", false, "Include total count in response")
	cmd.Flags().String("sort", "", "Sort results by field (e.g. '+name' or '-createDate')")
	cmd.Flags().String("group-by", "", "Group results by field")
	cmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml, or text")
	cmd.Flags().StringSlice("fields", nil, "Fields to display (use comma without spaces: --fields name,uuid,type or multiple flags: --fields name --fields type)")
}

func BuildQueryParams(cmd *cobra.Command, args []string, nameField string) (*param.QueryParam, error) {
	queryParam := param.NewQueryParam()

	if len(args) == 1 && nameField != "" {
		resourceName := args[0]
		queryParam.AddQ(fmt.Sprintf("%s=%s", nameField, resourceName))
	}

	qConditions, _ := cmd.Flags().GetStringArray("q")
	for _, q := range qConditions {
		queryParam.AddQ(q)
	}

	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 {
		queryParam.Limit(limit)
	}

	start, _ := cmd.Flags().GetInt("start")
	if start > 0 {
		queryParam.Start(start)
	}

	count, _ := cmd.Flags().GetBool("count")
	if count {
		queryParam.Count(true)
	}

	replyWithCount, _ := cmd.Flags().GetBool("reply-with-count")
	if replyWithCount {
		queryParam.ReplyWithCount(true)
	}

	sort, _ := cmd.Flags().GetString("sort")
	if sort != "" {
		queryParam.Sort(sort)
	}

	groupBy, _ := cmd.Flags().GetString("group-by")
	if groupBy != "" {
		queryParam.GroupBy(groupBy)
	}

	fields, _ := cmd.Flags().GetStringSlice("fields")
	if len(fields) > 0 {
		queryParam.Fields(fields)
	}

	if queryParam.Values == nil {
		queryParam.Values = url.Values{}
	}

	return &queryParam, nil
}
