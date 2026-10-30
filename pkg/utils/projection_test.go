package utils

import "testing"

type ProjectionBase struct {
	UUID string `json:"uuid"`
}
type projectionRecord struct {
	*ProjectionBase
	Name    string `json:"name"`
	Ignored string `json:"-"`
}

func TestProjectionEmbeddedFieldsAndNilPointers(t *testing.T) {
	for _, record := range []*projectionRecord{{ProjectionBase: &ProjectionBase{UUID: "id"}}, {}} {
		result, err := ProjectRows([]*projectionRecord{record}, []string{"UUID"})
		if err != nil {
			t.Fatal(err)
		}
		rows := result.([]map[string]any)
		if len(rows[0]) != 1 {
			t.Fatal(rows)
		}
	}
	if _, err := ProjectRows([]projectionRecord{}, []string{"Ignored"}); err == nil {
		t.Fatal("hidden field accepted")
	}
	if _, err := ProjectObject(nil, []string{"uuid"}); err == nil {
		t.Fatal("nil object accepted")
	}
	if _, err := ProjectRows([]projectionRecord{}, []string{"typo"}); err == nil {
		t.Fatal("unknown field accepted for empty output")
	}
}
