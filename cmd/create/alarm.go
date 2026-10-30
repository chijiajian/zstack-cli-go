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

package create

import (
	"fmt"
	"strings"

	zstackclient "github.com/chijiajian/zstack-cli-go/pkg/client"
	"github.com/chijiajian/zstack-cli-go/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

var alarmCmd = &cobra.Command{
	Use:   "alarm <name>",
	Short: "Create an alarm",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alarmParam, err := buildCreateAlarmParam(cmd, args[0])
		if err != nil {
			return err
		}

		if dryRunFlag {
			printCreateDryRun(alarmParam)
			return nil
		}

		cli := zstackclient.GetClient()
		if cli == nil {
			return fmt.Errorf("not logged in. Please run 'zstack-cli login' first")
		}

		result, err := cli.CreateAlarm(alarmParam)
		if err != nil {
			return err
		}
		return utils.PrintOperationResult("alarm", result, outputFlag)
	},
}

func init() {
	CreateCmd.AddCommand(alarmCmd)
	alarmCmd.Flags().StringP("description", "d", "", "Detailed description of the alarm")
	alarmCmd.Flags().String("namespace", "", "Metric namespace")
	alarmCmd.Flags().String("metric", "", "Metric name")
	alarmCmd.Flags().String("operator", "", "Comparison operator")
	alarmCmd.Flags().Float64("threshold", 0, "Alarm threshold")
	alarmCmd.Flags().Int("period", 0, "Metric period in seconds")
	alarmCmd.Flags().Int("repeat-interval", 0, "Repeat interval in seconds")
	alarmCmd.Flags().Int("repeat-count", 0, "Repeat count")
	alarmCmd.Flags().String("type", "", "Alarm type")
	alarmCmd.Flags().Bool("enable-recovery", false, "Enable recovery notification")
	alarmCmd.Flags().String("emergency-level", "", "Emergency level")
	alarmCmd.Flags().StringSlice("label", []string{}, "Alarm label in key=value form")
	alarmCmd.Flags().StringSlice("action", []string{}, "Action in uuid[:type] form")
	alarmCmd.Flags().String("resource-uuid", "", "Resource UUID")
	alarmCmd.Flags().StringSlice("tag", []string{}, "Tag UUID list")
	alarmCmd.Flags().StringSlice("system-tag", []string{}, "System tag list")
	alarmCmd.Flags().StringSlice("user-tag", []string{}, "User tag list")
}

func buildCreateAlarmParam(cmd *cobra.Command, name string) (param.CreateAlarmParam, error) {
	namespace, _ := cmd.Flags().GetString("namespace")
	metric, _ := cmd.Flags().GetString("metric")
	operator, _ := cmd.Flags().GetString("operator")
	if namespace == "" {
		return param.CreateAlarmParam{}, fmt.Errorf("required flag --namespace not set")
	}
	if metric == "" {
		return param.CreateAlarmParam{}, fmt.Errorf("required flag --metric not set")
	}
	if operator == "" {
		return param.CreateAlarmParam{}, fmt.Errorf("required flag --operator not set")
	}
	if !cmd.Flags().Changed("threshold") {
		return param.CreateAlarmParam{}, fmt.Errorf("required flag --threshold not set")
	}
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	description, _ := cmd.Flags().GetString("description")
	period, _ := cmd.Flags().GetInt("period")
	repeatInterval, _ := cmd.Flags().GetInt("repeat-interval")
	repeatCount, _ := cmd.Flags().GetInt("repeat-count")
	alarmType, _ := cmd.Flags().GetString("type")
	emergencyLevel, _ := cmd.Flags().GetString("emergency-level")
	resourceUUID, _ := cmd.Flags().GetString("resource-uuid")
	tagUUIDs, _ := cmd.Flags().GetStringSlice("tag")
	systemTags, _ := cmd.Flags().GetStringSlice("system-tag")
	userTags, _ := cmd.Flags().GetStringSlice("user-tag")
	labelValues, _ := cmd.Flags().GetStringSlice("label")
	actionValues, _ := cmd.Flags().GetStringSlice("action")

	labels, err := parseAlarmLabels(labelValues)
	if err != nil {
		return param.CreateAlarmParam{}, err
	}
	actions, err := parseAlarmActions(actionValues)
	if err != nil {
		return param.CreateAlarmParam{}, err
	}

	return param.CreateAlarmParam{
		BaseParam: param.BaseParam{SystemTags: systemTags, UserTags: userTags},
		Params: param.CreateAlarmParamDetail{
			Name:               name,
			Description:        stringPtrIfNotEmpty(description),
			ComparisonOperator: operator,
			Period:             intPtrIfNotZero(period),
			Namespace:          namespace,
			MetricName:         metric,
			Threshold:          threshold,
			RepeatInterval:     intPtrIfNotZero(repeatInterval),
			Labels:             labels,
			Actions:            actions,
			RepeatCount:        intPtrIfNotZero(repeatCount),
			Type:               stringPtrIfNotEmpty(alarmType),
			EnableRecovery:     changedBoolPtr(cmd, "enable-recovery"),
			EmergencyLevel:     stringPtrIfNotEmpty(emergencyLevel),
			ResourceUuid:       stringPtrIfNotEmpty(resourceUUID),
			TagUuids:           tagUUIDs,
		},
	}, nil
}

func parseAlarmLabels(values []string) ([]param.LabelParam, error) {
	labels := make([]param.LabelParam, 0, len(values))
	for _, value := range values {
		key, labelValue, ok := strings.Cut(value, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid --label %q; expected key=value", value)
		}
		labels = append(labels, param.LabelParam{
			Key:   stringPtrIfNotEmpty(key),
			Value: stringPtrIfNotEmpty(labelValue),
			Op:    "=",
		})
	}
	return labels, nil
}

func parseAlarmActions(values []string) ([]param.CreateAlarm_ActionParamParam, error) {
	actions := make([]param.CreateAlarm_ActionParamParam, 0, len(values))
	for _, value := range values {
		actionUUID, actionType, _ := strings.Cut(value, ":")
		if actionUUID == "" {
			return nil, fmt.Errorf("invalid --action %q; expected uuid[:type]", value)
		}
		actions = append(actions, param.CreateAlarm_ActionParamParam{
			ActionUuid: stringPtrIfNotEmpty(actionUUID),
			ActionType: stringPtrIfNotEmpty(actionType),
		})
	}
	return actions, nil
}
