package utils

import (
	"fmt"
	"github.com/chijiajian/zstack-cli-go/pkg/clierr"
	"reflect"
	"strings"
)

type projectedField struct {
	name  string
	index []int
}

func outputFields(t reflect.Type) []projectedField {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	result := []projectedField{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		}
		if f.Anonymous && name == "" {
			for _, nested := range outputFields(f.Type) {
				nested.index = append([]int{i}, nested.index...)
				result = append(result, nested)
			}
			continue
		}
		if name == "" {
			name = f.Name
		}
		result = append(result, projectedField{name, []int{i}})
	}
	return result
}
func selectFields(t reflect.Type, fields []string) ([]projectedField, error) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	available := outputFields(t)
	selected := []projectedField{}
	for _, name := range fields {
		found := false
		for _, f := range available {
			if strings.EqualFold(name, f.name) || strings.EqualFold(name, t.FieldByIndex(f.index).Name) {
				selected = append(selected, f)
				found = true
				break
			}
		}
		if !found {
			return nil, clierr.Wrap(clierr.Usage, fmt.Errorf("unknown output field %q", name))
		}
	}
	return selected, nil
}
func projectValue(v reflect.Value, fields []projectedField) map[string]any {
	result := map[string]any{}
	for _, f := range fields {
		value := v
		for _, index := range f.index {
			if value.Kind() == reflect.Pointer {
				if value.IsNil() {
					value = reflect.Value{}
					break
				}
				value = value.Elem()
			}
			value = value.Field(index)
		}
		if value.IsValid() {
			result[f.name] = value.Interface()
		} else {
			result[f.name] = nil
		}
	}
	return result
}

// ProjectObject validates and projects public JSON fields, including SDK embeddings.
func ProjectObject(data any, fields []string) (any, error) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, fmt.Errorf("field projection expects an object")
	}
	t := v.Type()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("field projection expects an object")
	}
	selected, err := selectFields(t, fields)
	if err != nil {
		return nil, err
	}
	return projectValue(v, selected), nil
}
func ProjectRows(data any, fields []string) (any, error) {
	v := reflect.ValueOf(data)
	if !v.IsValid() || v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("field projection expects a list")
	}
	selected, err := selectFields(v.Type().Elem(), fields)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		rows = append(rows, projectValue(v.Index(i), selected))
	}
	return rows, nil
}

// ValidateOutputFields checks the schema even for empty lists and table output.
func ValidateOutputFields(data any, fields []string) error {
	typ := reflect.TypeOf(data)
	if typ == nil {
		return fmt.Errorf("field validation expects an object or list")
	}
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}
	_, err := selectFields(typ, fields)
	return err
}
