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

package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type OutputFormat string

const (
	TableFormat OutputFormat = "table"
	JSONFormat  OutputFormat = "json"
	YAMLFormat  OutputFormat = "yaml"
	TextFormat  OutputFormat = "text"
)

type Formatter interface {
	Format(data interface{}, fields []string) error
}

func GetFormatter(format OutputFormat) Formatter {
	switch format {
	case TableFormat:
		return &TableFormatter{}
	case JSONFormat:
		return &JSONFormatter{}
	case YAMLFormat:
		return &YAMLFormatter{}
	case TextFormat:
		return &TextFormatter{}
	default:
		return &TableFormatter{}
	}
}

type JSONFormatter struct{}

func (f *JSONFormatter) Format(data interface{}, fields []string) error {
	if len(fields) > 0 {

		filteredData, err := filterFields(data, fields)
		if err != nil {
			return err
		}
		data = filteredData
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(data interface{}, fields []string) error {
	if len(fields) > 0 {
		filteredData, err := filterFields(data, fields)
		if err != nil {
			return err
		}
		data = filteredData
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(yamlData))
	return nil
}

type TextFormatter struct{}

func (f *TextFormatter) Format(data interface{}, fields []string) error {
	if len(fields) > 0 {
		filteredData, err := filterFields(data, fields)
		if err != nil {
			return err
		}
		data = filteredData
	}

	fmt.Printf("%v\n", data)
	return nil
}

type TableFormatter struct{}

func (f *TableFormatter) Format(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return formatSlice(v.Interface(), fields)
	case reflect.Map:
		return formatMap(v.Interface(), fields)
	case reflect.Struct:
		return formatStruct(data, fields)
	default:
		fmt.Printf("%v\n", data)
		return nil
	}
}

func formatSlice(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)

	if v.Len() == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	firstElem := v.Index(0)

	if firstElem.Kind() == reflect.Struct {
		return formatStructSlice(data, fields)
	} else if firstElem.Kind() == reflect.Map {
		return formatMapSlice(data, fields)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Value"})

	for i := 0; i < v.Len(); i++ {
		val := fmt.Sprintf("%v", v.Index(i).Interface())
		table.Append([]string{val})
	}

	table.Render()
	return nil
}

func formatStructSlice(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	elemType := reflect.TypeOf(v.Index(0).Interface())

	var headers []string
	var fieldIndices []int

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		tagName := field.Tag.Get("json")
		fieldName := field.Name
		if tagName != "" && tagName != "-" {

			tagParts := strings.Split(tagName, ",")
			fieldName = tagParts[0]
		}

		if len(fields) > 0 {
			include := false
			for _, f := range fields {
				if strings.EqualFold(f, fieldName) || strings.EqualFold(f, field.Name) {
					include = true
					break
				}
			}
			if !include {
				continue
			}
		}

		headers = append(headers, fieldName)
		fieldIndices = append(fieldIndices, i)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(headers)

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		var row []string

		for _, idx := range fieldIndices {
			fieldValue := item.Field(idx)
			row = append(row, fmt.Sprintf("%v", fieldValue.Interface()))
		}

		table.Append(row)
	}

	table.Render()
	return nil
}

func formatMapSlice(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	keysSet := make(map[string]bool)
	for i := 0; i < v.Len(); i++ {
		mapItem := v.Index(i).Interface().(map[string]interface{})
		for key := range mapItem {

			if len(fields) > 0 {
				include := false
				for _, f := range fields {
					if strings.EqualFold(f, key) {
						include = true
						break
					}
				}
				if !include {
					continue
				}
			}
			keysSet[key] = true
		}
	}

	var headers []string
	for key := range keysSet {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(headers)

	for i := 0; i < v.Len(); i++ {
		mapItem := v.Index(i).Interface().(map[string]interface{})
		var row []string

		for _, header := range headers {
			val, exists := mapItem[header]
			if exists {
				row = append(row, fmt.Sprintf("%v", val))
			} else {
				row = append(row, "")
			}
		}

		table.Append(row)
	}

	table.Render()
	return nil
}

func formatMap(data interface{}, fields []string) error {
	m, ok := data.(map[string]interface{})
	if !ok {

		v := reflect.ValueOf(data)
		if v.Kind() != reflect.Map {
			return fmt.Errorf("expected map, got %T", data)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Key", "Value"})

		iter := v.MapRange()
		for iter.Next() {
			key := fmt.Sprintf("%v", iter.Key().Interface())

			if len(fields) > 0 {
				include := false
				for _, f := range fields {
					if strings.EqualFold(f, key) {
						include = true
						break
					}
				}
				if !include {
					continue
				}
			}

			value := fmt.Sprintf("%v", iter.Value().Interface())
			table.Append([]string{key, value})
		}

		table.Render()
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Key", "Value"})

	var keys []string
	for k := range m {

		if len(fields) > 0 {
			include := false
			for _, f := range fields {
				if strings.EqualFold(f, k) {
					include = true
					break
				}
			}
			if !include {
				continue
			}
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		table.Append([]string{k, fmt.Sprintf("%v", v)})
	}

	table.Render()
	return nil
}

func formatStruct(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %T", data)
	}

	t := v.Type()

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Field", "Value"})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		fieldName := field.Name
		tagName := field.Tag.Get("json")
		if tagName != "" && tagName != "-" {

			tagParts := strings.Split(tagName, ",")
			fieldName = tagParts[0]
		}

		if len(fields) > 0 {
			include := false
			for _, f := range fields {
				if strings.EqualFold(f, fieldName) || strings.EqualFold(f, field.Name) {
					include = true
					break
				}
			}
			if !include {
				continue
			}
		}

		fieldValue := v.Field(i)
		table.Append([]string{fieldName, fmt.Sprintf("%v", fieldValue.Interface())})
	}

	table.Render()
	return nil
}

func filterFields(data interface{}, fields []string) (interface{}, error) {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {

		resultSlice := reflect.MakeSlice(v.Type(), 0, v.Len())

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)

			if elem.Kind() == reflect.Struct {

				newElem := reflect.New(elem.Type()).Elem()

				for j := 0; j < elem.NumField(); j++ {
					field := elem.Type().Field(j)
					fieldName := field.Name
					tagName := field.Tag.Get("json")
					if tagName != "" && tagName != "-" {
						tagParts := strings.Split(tagName, ",")
						fieldName = tagParts[0]
					}

					include := false
					for _, f := range fields {
						if strings.EqualFold(f, fieldName) || strings.EqualFold(f, field.Name) {
							include = true
							break
						}
					}

					if include {
						newElem.Field(j).Set(elem.Field(j))
					}
				}

				resultSlice = reflect.Append(resultSlice, newElem)
			} else if elem.Kind() == reflect.Map {

				newMap := reflect.MakeMap(elem.Type())

				iter := elem.MapRange()
				for iter.Next() {
					key := iter.Key()
					keyStr := fmt.Sprintf("%v", key.Interface())

					include := false
					for _, f := range fields {
						if strings.EqualFold(f, keyStr) {
							include = true
							break
						}
					}

					if include {
						newMap.SetMapIndex(key, iter.Value())
					}
				}

				resultSlice = reflect.Append(resultSlice, newMap)
			} else {

				resultSlice = reflect.Append(resultSlice, elem)
			}
		}

		return resultSlice.Interface(), nil
	} else if v.Kind() == reflect.Struct {

		newStruct := reflect.New(v.Type()).Elem()

		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			fieldName := field.Name
			tagName := field.Tag.Get("json")
			if tagName != "" && tagName != "-" {
				tagParts := strings.Split(tagName, ",")
				fieldName = tagParts[0]
			}

			include := false
			for _, f := range fields {
				if strings.EqualFold(f, fieldName) || strings.EqualFold(f, field.Name) {
					include = true
					break
				}
			}

			if include {
				newStruct.Field(i).Set(v.Field(i))
			}
		}

		return newStruct.Interface(), nil
	} else if v.Kind() == reflect.Map {

		newMap := reflect.MakeMap(v.Type())

		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			keyStr := fmt.Sprintf("%v", key.Interface())

			include := false
			for _, f := range fields {
				if strings.EqualFold(f, keyStr) {
					include = true
					break
				}
			}

			if include {
				newMap.SetMapIndex(key, iter.Value())
			}
		}

		return newMap.Interface(), nil
	}

	return data, nil
}

func Print(data interface{}, format OutputFormat) error {
	return PrintWithFields(data, format, nil)
}

func PrintWithFields(data interface{}, format OutputFormat, fields []string) error {
	formatter := GetFormatter(format)
	return formatter.Format(data, fields)
}

func ParseFormat(format string) OutputFormat {
	switch strings.ToLower(format) {
	case "json":
		return JSONFormat
	case "yaml":
		return YAMLFormat
	case "text":
		return TextFormat
	case "table":
		return TableFormat
	default:
		return TableFormat
	}
}

func PrintDryRun(data interface{}, format string) {
	switch format {
	case "json":
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
	case "yaml":
		yamlData, _ := yaml.Marshal(data)
		fmt.Println(string(yamlData))
	default:
		fmt.Printf("Would create with parameters: %+v\n", data)
	}
}

type ResourceTableDefinition struct {
	Headers []string
	Fields  []FieldDefinition
}

type FieldDefinition struct {
	Path      []string
	Formatter func(interface{}) string
}

var resourceTableDefinitions = map[string]ResourceTableDefinition{
	"image": {
		Headers: []string{"NAME", "UUID", "STATUS", "SIZE", "MEDIA-TYPE", "FORMAT", "CREATED"},
		Fields: []FieldDefinition{
			{Path: []string{"Name"}, Formatter: stringFormatter},
			{Path: []string{"UUID"}, Formatter: stringFormatter},
			{Path: []string{"Status"}, Formatter: stringFormatter},
			{Path: []string{"Size"}, Formatter: sizeFormatter},
			{Path: []string{"MediaType"}, Formatter: stringFormatter},
			{Path: []string{"Format"}, Formatter: stringFormatter},
			{Path: []string{"CreateDate"}, Formatter: timeFormatter},
		},
	},
	"instance": {
		Headers: []string{"NAME", "UUID", "STATUS", "HOST", "CPU", "MEMORY", "IMAGE", "CREATED"},
		Fields: []FieldDefinition{
			{Path: []string{"Name"}, Formatter: stringFormatter},
			{Path: []string{"UUID"}, Formatter: stringFormatter},
			{Path: []string{"State"}, Formatter: stringFormatter},
			{Path: []string{"HostUUID"}, Formatter: stringFormatter},
			{Path: []string{"CPUNum"}, Formatter: intFormatter},
			{Path: []string{"MemorySize"}, Formatter: sizeFormatter},
			{Path: []string{"ImageUUID"}, Formatter: stringFormatter},
			{Path: []string{"CreateDate"}, Formatter: timeFormatter},
		},
	},
	"instanceoffering": {
		Headers: []string{"NAME", "UUID", "CPU", "MEMORY", "TYPE", "ALLOCATOR_STRATEGY", "STATE"},
		Fields: []FieldDefinition{
			{Path: []string{"Name"}, Formatter: stringFormatter},
			{Path: []string{"UUID"}, Formatter: stringFormatter},
			{Path: []string{"CpuNum"}, Formatter: intFormatter},
			{Path: []string{"MemorySize"}, Formatter: sizeFormatter},
			{Path: []string{"Type"}, Formatter: stringFormatter},
			{Path: []string{"AllocatorStrategy"}, Formatter: stringFormatter},
			{Path: []string{"State"}, Formatter: stringFormatter},
		},
	},
	"diskoffering": {
		Headers: []string{"NAME", "UUID", "DISK_SIZE", "TYPE", "ALLOCATOR_STRATEGY", "STATE", "CREATED"},
		Fields: []FieldDefinition{
			{Path: []string{"Name"}, Formatter: stringFormatter},
			{Path: []string{"UUID"}, Formatter: stringFormatter},
			{Path: []string{"DiskSize"}, Formatter: sizeFormatter},
			{Path: []string{"Type"}, Formatter: stringFormatter},
			{Path: []string{"AllocatorStrategy"}, Formatter: stringFormatter},
			{Path: []string{"State"}, Formatter: stringFormatter},
			{Path: []string{"CreateDate"}, Formatter: timeFormatter},
		},
	},
}

var resourceTypeAliases = map[string]string{
	"virtualmachine":    "instance",
	"l3network":         "network",
	"network":           "network",
	"instanceoffering":  "instanceoffering",
	"instance":          "instance",
	"vm":                "instance",
	"instance-offering": "instanceoffering",
}

func PrintOperationResult(resourceType string, result interface{}, format string) {
	switch format {
	case "json":
		printJSON(result)
	case "yaml":
		printYAML(result)
	case "wide":
		printWideFormat(resourceType, result)
	case "name":
		printNameOnly(resourceType, result)
	default:
		printSimpleFormat(resourceType, result)
	}
}

func printSimpleFormat(resourceType string, result interface{}) {
	name := extractName(result)
	if name != "" {
		fmt.Printf("%s/%s created\n", strings.ToLower(resourceType), name)
	} else {
		fmt.Printf("%s created successfully\n", resourceType)
	}
}

func printNameOnly(resourceType string, result interface{}) {
	name := extractName(result)
	if name != "" {
		fmt.Printf("%s/%s\n", strings.ToLower(resourceType), name)
	} else {
		fmt.Printf("%s\n", resourceType)
	}
}

func extractName(result interface{}) string {
	if result == nil {
		return ""
	}

	return getFieldValueAsString(result, "Name")
}

func printWideFormat(resourceType string, result interface{}) {
	if result == nil {
		fmt.Printf("No result data to display\n")
		return
	}

	normalizedType := normalizeResourceType(resourceType)

	tableDef, found := resourceTableDefinitions[normalizedType]
	if !found {

		fmt.Printf("No table definition for resource type '%s', using JSON format:\n", resourceType)
		printJSON(result)
		return
	}

	printResourceTable(result, tableDef)
}

func normalizeResourceType(resourceType string) string {
	lowerType := strings.ToLower(resourceType)
	if alias, found := resourceTypeAliases[lowerType]; found {
		return alias
	}
	return lowerType
}

func printResourceTable(result interface{}, tableDef ResourceTableDefinition) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, strings.Join(tableDef.Headers, "\t"))

	var values []string
	for _, fieldDef := range tableDef.Fields {
		value := getFieldValue(result, fieldDef.Path)
		formattedValue := fieldDef.Formatter(value)
		values = append(values, formattedValue)
	}
	fmt.Fprintln(w, strings.Join(values, "\t"))

	w.Flush()
}

func getFieldValue(obj interface{}, path []string) interface{} {
	if obj == nil || len(path) == 0 {
		return nil
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	fieldName := path[0]
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}

	if len(path) == 1 {
		return field.Interface()
	}

	return getFieldValue(field.Interface(), path[1:])
}

func getFieldValueAsString(obj interface{}, fieldName string) string {
	value := getFieldValue(obj, []string{fieldName})
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func stringFormatter(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func intFormatter(value interface{}) string {
	if value == nil {
		return "0"
	}

	switch v := value.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%d", int(v))
	default:
		return fmt.Sprintf("%v", value)
	}
}

func sizeFormatter(value interface{}) string {
	if value == nil {
		return "0"
	}

	var size int64
	switch v := value.(type) {
	case int:
		size = int64(v)
	case int32:
		size = int64(v)
	case int64:
		size = v
	case float64:
		size = int64(v)
	default:
		return fmt.Sprintf("%v", value)
	}

	return formatSize(size)
}

func timeFormatter(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case time.Time:
		return formatTime(v)
	case string:

		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			return formatTime(t)
		}
		return v
	default:
		return fmt.Sprintf("%v", value)
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func printJSON(result interface{}) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling to JSON: %s\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

func printYAML(result interface{}) {
	yamlData, err := yaml.Marshal(result)
	if err != nil {
		fmt.Printf("Error marshalling to YAML: %s\n", err)
		return
	}
	fmt.Println(string(yamlData))
}
