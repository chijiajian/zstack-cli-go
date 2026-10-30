package utils

import (
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
)

// tablewriter may discard renderer write errors. Record the first failure at the
// writer boundary so every table formatter reports the same output contract.
type outputWriter struct {
	writer io.Writer
	err    error
}

func (w *outputWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.writer.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	w.err = err
	return n, err
}
func newOutputTable() (*tablewriter.Table, *outputWriter) {
	output := &outputWriter{writer: os.Stdout}
	return tablewriter.NewWriter(output), output
}
func renderOutputTable(table *tablewriter.Table, output *outputWriter) error {
	if err := table.Render(); err != nil {
		return err
	}
	return output.err
}
