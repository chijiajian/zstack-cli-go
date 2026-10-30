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
	"github.com/chijiajian/zstack-cli-go/pkg/query"
	"github.com/spf13/cobra"
)

type FormattedAlarm struct {
	Name               string  `json:"name" yaml:"name" header:"NAME"`
	UUID               string  `json:"uuid" yaml:"uuid" header:"UUID"`
	Description        string  `json:"description" yaml:"description" header:"DESCRIPTION"`
	Namespace          string  `json:"namespace" yaml:"namespace" header:"NAMESPACE"`
	MetricName         string  `json:"metricName" yaml:"metricName" header:"METRIC"`
	ComparisonOperator string  `json:"comparisonOperator" yaml:"comparisonOperator" header:"OPERATOR"`
	Threshold          float64 `json:"threshold" yaml:"threshold" header:"THRESHOLD"`
	Status             string  `json:"status" yaml:"status" header:"STATUS"`
	State              string  `json:"state" yaml:"state" header:"STATE"`
	EmergencyLevel     string  `json:"emergencyLevel" yaml:"emergencyLevel" header:"LEVEL"`
}

type FormattedEventRecord struct {
	ID             int64  `json:"id" yaml:"id" header:"ID"`
	UUID           string `json:"uuid" yaml:"uuid" header:"UUID"`
	CreateTime     int64  `json:"createTime" yaml:"createTime" header:"CREATE TIME"`
	Namespace      string `json:"namespace" yaml:"namespace" header:"NAMESPACE"`
	EmergencyLevel string `json:"emergencyLevel" yaml:"emergencyLevel" header:"LEVEL"`
	ResourceID     string `json:"resourceId" yaml:"resourceId" header:"RESOURCE ID"`
	DataUUID       string `json:"dataUuid" yaml:"dataUuid" header:"DATA UUID"`
	ReadStatus     bool   `json:"readStatus" yaml:"readStatus" header:"READ"`
	Labels         string `json:"labels" yaml:"labels" header:"LABELS"`
}

type FormattedAudit struct {
	ID           int64  `json:"id" yaml:"id" header:"ID"`
	UUID         string `json:"uuid" yaml:"uuid" header:"UUID"`
	CreateTime   int64  `json:"createTime" yaml:"createTime" header:"CREATE TIME"`
	APIName      string `json:"apiName" yaml:"apiName" header:"API"`
	Operator     string `json:"operator" yaml:"operator" header:"OPERATOR"`
	ClientIP     string `json:"clientIp" yaml:"clientIp" header:"CLIENT IP"`
	Duration     int64  `json:"duration" yaml:"duration" header:"DURATION"`
	ResourceUUID string `json:"resourceUuid" yaml:"resourceUuid" header:"RESOURCE UUID"`
	ResourceType string `json:"resourceType" yaml:"resourceType" header:"RESOURCE TYPE"`
	ResourceName string `json:"resourceName" yaml:"resourceName" header:"RESOURCE NAME"`
	Success      bool   `json:"success" yaml:"success" header:"SUCCESS"`
}

var alarmsCmd = &cobra.Command{
	Use:   "alarms [name]",
	Short: "List alarms",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}
		result, err := common.Fetch[view.AlarmInventoryView](cobraCmd, queryParam, "v1/zwatch/alarms")
		alarms := result.Items
		if err != nil {
			return fmt.Errorf("Error querying alarms: %w", err)
		}
		formatted := make([]FormattedAlarm, 0, len(alarms))
		for _, alarm := range alarms {
			formatted = append(formatted, FormattedAlarm{
				Name:               alarm.Name,
				UUID:               alarm.UUID,
				Description:        alarm.Description,
				Namespace:          alarm.Namespace,
				MetricName:         alarm.MetricName,
				ComparisonOperator: alarm.ComparisonOperator,
				Threshold:          alarm.Threshold,
				Status:             alarm.Status,
				State:              alarm.State,
				EmergencyLevel:     alarm.EmergencyLevel,
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

var eventsCmd = &cobra.Command{
	Use:   "events [namespace]",
	Short: "List event records",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		queryParam, err := common.BuildQueryParams(cobraCmd, args, "namespace")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}
		result, err := common.Fetch[view.EventRecordsInventoryView](cobraCmd, queryParam, "v1/zwatch/event-records", query.WithDefaultSort("+id"))
		events := result.Items
		if err != nil {
			return fmt.Errorf("Error querying events: %w", err)
		}
		formatted := make([]FormattedEventRecord, 0, len(events))
		for _, event := range events {
			formatted = append(formatted, FormattedEventRecord{
				ID:             event.Id,
				UUID:           event.UUID,
				CreateTime:     event.CreateTime,
				Namespace:      event.Namespace,
				EmergencyLevel: event.EmergencyLevel,
				ResourceID:     event.ResourceId,
				DataUUID:       event.DataUuid,
				ReadStatus:     event.ReadStatus,
				Labels:         event.Labels,
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

var auditsCmd = &cobra.Command{
	Use:   "audits [api-name]",
	Short: "List audit records",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		queryParam, err := common.BuildQueryParams(cobraCmd, args, "apiName")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}
		result, err := common.Fetch[view.AuditsInventoryView](cobraCmd, queryParam, "v1/zwatch/audit-records", query.WithDefaultSort("+id"))
		audits := result.Items
		if err != nil {
			return fmt.Errorf("Error querying audits: %w", err)
		}
		formatted := make([]FormattedAudit, 0, len(audits))
		for _, audit := range audits {
			formatted = append(formatted, FormattedAudit{
				ID:           audit.Id,
				UUID:         audit.UUID,
				CreateTime:   audit.CreateTime,
				APIName:      audit.ApiName,
				Operator:     audit.Operator,
				ClientIP:     audit.ClientIp,
				Duration:     audit.Duration,
				ResourceUUID: audit.ResourceUuid,
				ResourceType: audit.ResourceType,
				ResourceName: audit.ResourceName,
				Success:      audit.Success,
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

func init() {
	GetCmd.AddCommand(alarmsCmd, eventsCmd, auditsCmd)
	common.AddQueryFlags(alarmsCmd)
	common.AddQueryFlags(eventsCmd)
	common.AddQueryFlags(auditsCmd)
}
