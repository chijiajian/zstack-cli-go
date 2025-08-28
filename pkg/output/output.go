package output

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// Format 表示输出格式类型
type Format string

const (
	TableFormat Format = "table"
	JSONFormat  Format = "json"
	YAMLFormat  Format = "yaml"
	TextFormat  Format = "text"
)

// Formatter 是输出格式化器接口
type Formatter interface {
	Format(data interface{}, fields []string) error
}

// GetFormatter 根据指定的格式返回相应的格式化器
func GetFormatter(format Format) Formatter {
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
		// 默认使用表格格式
		return &TableFormatter{}
	}
}

// JSONFormatter 实现 JSON 格式输出
type JSONFormatter struct{}

func (f *JSONFormatter) Format(data interface{}, fields []string) error {
	if len(fields) > 0 {
		// 如果指定了字段，过滤数据
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

// YAMLFormatter 实现 YAML 格式输出
type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(data interface{}, fields []string) error {
	if len(fields) > 0 {
		// 如果指定了字段，过滤数据
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

// TextFormatter 实现纯文本格式输出
type TextFormatter struct{}

func (f *TextFormatter) Format(data interface{}, fields []string) error {
	if len(fields) > 0 {
		// 如果指定了字段，过滤数据
		filteredData, err := filterFields(data, fields)
		if err != nil {
			return err
		}
		data = filteredData
	}

	fmt.Printf("%v\n", data)
	return nil
}

// TableFormatter 实现表格格式输出
type TableFormatter struct{}

func (f *TableFormatter) Format(data interface{}, fields []string) error {
	// 处理不同类型的数据
	v := reflect.ValueOf(data)

	// 如果是指针，获取它指向的值
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// 根据数据类型选择不同的表格渲染方式
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return formatSlice(v.Interface(), fields)
	case reflect.Map:
		return formatMap(v.Interface(), fields)
	case reflect.Struct:
		return formatStruct(data, fields)
	default:
		// 对于简单类型，直接打印
		fmt.Printf("%v\n", data)
		return nil
	}
}

// formatSlice 格式化切片或数组为表格
func formatSlice(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)

	// 如果是空切片，显示提示信息
	if v.Len() == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	// 获取第一个元素，用于确定表头
	firstElem := v.Index(0)

	// 如果元素是结构体，使用结构体字段作为表头
	if firstElem.Kind() == reflect.Struct {
		return formatStructSlice(data, fields)
	} else if firstElem.Kind() == reflect.Map {
		return formatMapSlice(data, fields)
	}

	// 对于基本类型切片，简单列出
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Value"})

	for i := 0; i < v.Len(); i++ {
		val := fmt.Sprintf("%v", v.Index(i).Interface())
		table.Append([]string{val})
	}

	table.Render()
	return nil
}

// formatStructSlice 格式化结构体切片为表格
func formatStructSlice(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	// 获取第一个元素的类型
	elemType := reflect.TypeOf(v.Index(0).Interface())

	// 提取字段名作为表头
	var headers []string
	var fieldIndices []int

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		// 使用 json 标签作为列名，如果没有则使用字段名
		tagName := field.Tag.Get("json")
		fieldName := field.Name
		if tagName != "" && tagName != "-" {
			// 去除 json 标签中的选项部分
			tagParts := strings.Split(tagName, ",")
			fieldName = tagParts[0]
		}

		// 如果指定了字段列表，则只显示这些字段
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

	// 创建表格
	table := tablewriter.NewWriter(os.Stdout)
	table.Header(headers)

	// 添加数据行
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

// formatMapSlice 格式化 map 切片为表格
func formatMapSlice(data interface{}, fields []string) error {
	v := reflect.ValueOf(data)
	if v.Len() == 0 {
		return nil
	}

	// 收集所有可能的键作为表头
	keysSet := make(map[string]bool)
	for i := 0; i < v.Len(); i++ {
		mapItem := v.Index(i).Interface().(map[string]interface{})
		for key := range mapItem {
			// 如果指定了字段列表，则只考虑这些字段
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

	// 将键转换为有序切片
	var headers []string
	for key := range keysSet {
		headers = append(headers, key)
	}
	sort.Strings(headers) // 对表头进行排序，使输出更一致

	// 创建表格
	table := tablewriter.NewWriter(os.Stdout)
	table.Header(headers)

	// 添加数据行
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

// formatMap 格式化单个 map 为表格
func formatMap(data interface{}, fields []string) error {
	m, ok := data.(map[string]interface{})
	if !ok {
		// 尝试处理其他类型的 map
		v := reflect.ValueOf(data)
		if v.Kind() != reflect.Map {
			return fmt.Errorf("expected map, got %T", data)
		}

		// 创建表格
		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Key", "Value"})

		// 遍历 map 的键值对
		iter := v.MapRange()
		for iter.Next() {
			key := fmt.Sprintf("%v", iter.Key().Interface())

			// 如果指定了字段列表，则只显示这些字段
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

	// 处理 map[string]interface{} 类型
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Key", "Value"})

	// 为了保证输出顺序一致，对键进行排序
	var keys []string
	for k := range m {
		// 如果指定了字段列表，则只显示这些字段
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

// formatStruct 格式化结构体为表格
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

		// 使用 json 标签作为字段名，如果没有则使用字段名
		fieldName := field.Name
		tagName := field.Tag.Get("json")
		if tagName != "" && tagName != "-" {
			// 去除 json 标签中的选项部分
			tagParts := strings.Split(tagName, ",")
			fieldName = tagParts[0]
		}

		// 如果指定了字段列表，则只显示这些字段
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

// filterFields 根据字段列表过滤数据
func filterFields(data interface{}, fields []string) (interface{}, error) {
	v := reflect.ValueOf(data)

	// 如果是指针，获取它指向的值
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// 处理切片类型
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		// 创建一个新的切片来存储过滤后的数据
		resultSlice := reflect.MakeSlice(v.Type(), 0, v.Len())

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)

			// 如果元素是结构体，过滤其字段
			if elem.Kind() == reflect.Struct {
				// 创建一个新的结构体
				newElem := reflect.New(elem.Type()).Elem()

				// 复制指定的字段
				for j := 0; j < elem.NumField(); j++ {
					field := elem.Type().Field(j)
					fieldName := field.Name
					tagName := field.Tag.Get("json")
					if tagName != "" && tagName != "-" {
						// 去除 json 标签中的选项部分
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
				// 处理 map 类型的元素
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
				// 对于其他类型，直接添加
				resultSlice = reflect.Append(resultSlice, elem)
			}
		}

		return resultSlice.Interface(), nil
	} else if v.Kind() == reflect.Struct {
		// 处理单个结构体
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
		// 处理 map
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

	// 对于其他类型，返回原始数据
	return data, nil
}

// Print 是一个便捷函数，根据指定的格式输出数据
func Print(data interface{}, format Format) error {
	return PrintWithFields(data, format, nil)
}

// PrintWithFields 是一个便捷函数，根据指定的格式和字段列表输出数据
func PrintWithFields(data interface{}, format Format, fields []string) error {
	formatter := GetFormatter(format)
	return formatter.Format(data, fields)
}

// ParseFormat 将字符串转换为 Format 类型
func ParseFormat(format string) Format {
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

// PrintDryRun 打印dry-run模式下的参数
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

// PrintOperationResult 打印操作结果
func PrintOperationResult(resourceType string, data interface{}, format string) {
	switch format {
	case "json":
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(jsonData))
	case "yaml":
		yamlData, _ := yaml.Marshal(data)
		fmt.Println(string(yamlData))
	default:
		fmt.Printf("%s created successfully\n", resourceType)
		fmt.Printf("Details: %+v\n", data)
	}
}
