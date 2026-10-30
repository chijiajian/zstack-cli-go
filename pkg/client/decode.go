package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/zstackio/zstack-sdk-go-v2/pkg/util/timeutils"
)

// decodeResponse keeps standard JSON validation and exact numeric decoding, while
// accepting the SDK's timestamp formats only for fields declared as time.Time.
func decodeResponse(data []byte, result any) error {
	var value any
	if err := decodeJSONDocument(bytes.NewReader(data), &value); err != nil {
		return err
	}

	normalized, err := normalizeResponseTimes(value, reflect.TypeOf(result))
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	return decoder.Decode(result)
}
func normalizeResponseTimes(value any, typ reflect.Type) (any, error) {
	if typ == nil || value == nil {
		return value, nil
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == reflect.TypeOf(time.Time{}) {
		text, ok := value.(string)
		if !ok {
			return value, nil
		}
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			parsed, err = timeutils.ParseTimeStr(text)
		}
		if err != nil {
			return nil, err
		}
		return parsed.Format(time.RFC3339Nano), nil
	}
	switch typ.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return value, nil
		}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "-" {
				continue
			}
			if field.Anonymous && name == "" {
				if _, err := normalizeResponseTimes(object, field.Type); err != nil {
					return nil, err
				}
				continue
			}
			if name == "" {
				name = field.Name
			}
			// encoding/json matches field names case-insensitively.
			for key, entry := range object {
				if !strings.EqualFold(key, name) {
					continue
				}
				normalized, err := normalizeResponseTimes(entry, field.Type)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", key, err)
				}
				object[key] = normalized
			}

		}
	case reflect.Map:
		object, ok := value.(map[string]any)
		if !ok || typ.Key().Kind() != reflect.String {
			return value, nil
		}
		for key, entry := range object {
			normalized, err := normalizeResponseTimes(entry, typ.Elem())
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			object[key] = normalized
		}
	case reflect.Slice, reflect.Array:
		items, ok := value.([]any)
		if !ok {
			return value, nil
		}
		for i, item := range items {
			normalized, err := normalizeResponseTimes(item, typ.Elem())
			if err != nil {
				return nil, err
			}
			items[i] = normalized
		}
	}
	return value, nil
}

// decodeJSONDocument rejects extra JSON values and trailing garbage, while allowing
// whitespace. All response consumers share this framing contract.
func decodeJSONDocument(reader io.Reader, result any) error {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(result); err != nil {
		return err
	}
	var extra json.RawMessage
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return fmt.Errorf("trailing JSON content: %w", err)
	}
	return fmt.Errorf("multiple JSON values in response")
}
