package format

import (
	"fmt"
	"io"
)

// OutputFormat specifies the desired output format.
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatCSV   OutputFormat = "csv"
	FormatTSV   OutputFormat = "tsv"
)

// ParseFormat parses the output format string.
func ParseFormat(s string) (OutputFormat, error) {
	switch s {
	case "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	case "tsv":
		return FormatTSV, nil
	default:
		return "", fmt.Errorf("unknown output format %q; valid values: table, json, csv, tsv", s)
	}
}

// Write writes tabular data using the specified output format.
func Write(w io.Writer, format OutputFormat, headers []string, rows [][]string, jsonVal interface{}) error {
	switch format {
	case FormatJSON:
		return WriteJSON(w, jsonVal)
	case FormatCSV:
		return WriteCSV(w, headers, rows)
	case FormatTSV:
		return WriteTSV(w, headers, rows)
	default:
		return WriteTable(w, headers, rows)
	}
}
