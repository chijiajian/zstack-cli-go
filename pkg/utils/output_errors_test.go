package utils

import (
	"os"
	"testing"
)

func TestFormattersPropagateOutputFailure(t *testing.T) {
	closed, err := os.CreateTemp(t.TempDir(), "closed-output")
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	original := os.Stdout
	os.Stdout = closed
	defer func() { os.Stdout = original }()
	for _, format := range []OutputFormat{JSONFormat, YAMLFormat, TextFormat, TableFormat} {
		t.Run(string(format), func(t *testing.T) {
			if err := PrintDryRun(map[string]any{"uuid": "id1"}, string(format)); err == nil {
				t.Fatal("dry-run output failure reported success")
			}
			if err := Print([]map[string]any{{"uuid": "id1"}}, format); err == nil {
				t.Fatal("output write failure reported success")
			}
		})
	}
}

func TestDryRunRejectsJSONEncodingFailure(t *testing.T) {
	if err := PrintDryRun(make(chan int), "json"); err == nil {
		t.Fatal("encoding failure reported success")
	}
}

func TestOperationResultsReportWriteAndEncodingFailures(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	original := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = original }()
	result := struct{ Name string }{"demo"}
	for _, format := range []string{"", "text", "name", "wide", "json", "yaml"} {
		if err := PrintOperationResult("instance", result, format); err == nil {
			t.Fatalf("format %q swallowed failure", format)
		}
	}
	if err := PrintOperationResult("other", result, "wide"); err == nil {
		t.Fatal("fallback swallowed output failure")
	}
	if err := PrintOperationResult("instance", make(chan int), "json"); err == nil {
		t.Fatal("encoding failure swallowed")
	}
}
