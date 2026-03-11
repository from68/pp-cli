package format

import (
	"encoding/csv"
	"io"

	"github.com/olekukonko/tablewriter"
)

// WriteTable writes rows with headers as an ASCII table.
func WriteTable(w io.Writer, headers []string, rows [][]string) error {
	t := tablewriter.NewTable(w)
	t.Header(headers)
	for _, row := range rows {
		if err := t.Append(row); err != nil {
			return err
		}
	}
	return t.Render()
}

// WriteCSV writes rows with headers as comma-separated values.
func WriteCSV(w io.Writer, headers []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(headers); err != nil {
		return err
	}
	if err := cw.WriteAll(rows); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

// WriteTSV writes rows with headers as tab-separated values.
func WriteTSV(w io.Writer, headers []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	cw.Comma = '\t'
	if err := cw.Write(headers); err != nil {
		return err
	}
	if err := cw.WriteAll(rows); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}
