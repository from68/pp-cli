package format

import (
	"encoding/json"
	"io"
)

// WriteJSON writes the given value as a JSON array to w.
func WriteJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
