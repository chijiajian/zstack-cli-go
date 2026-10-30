package inspect

import (
	"encoding/json"
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNormalizeNestedNumbers(t *testing.T) {
	data := map[string]any{"values": []any{json.Number("9007199254740993"), json.Number("18446744073709551615"), map[string]any{"size": json.Number("1.5")}}}
	normalizeNumbers(data)
	encoded, err := yaml.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err = yaml.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	values := decoded["values"].([]any)
	if fmt.Sprint(values[0]) != "9007199254740993" || values[1] != uint64(18446744073709551615) {
		t.Fatalf("numbers changed: %s", encoded)
	}
}
