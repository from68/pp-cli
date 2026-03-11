package xml

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrUnsupportedFormat is returned when the file format is not supported.
var ErrUnsupportedFormat = errors.New("unsupported file format")

// Load opens the portfolio file and returns a reader to the XML content.
// It detects the format by magic bytes and returns ErrUnsupportedFormat for
// ZIP and AES formats (planned for a later phase).
func Load(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("cannot open file: %w", err)
	}

	// Read enough bytes to detect the format.
	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		f.Close()
		return nil, fmt.Errorf("cannot read file header: %w", err)
	}
	header = header[:n]

	// Seek back to the start so the decoder sees the full content.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, fmt.Errorf("cannot seek file: %w", err)
	}

	if err := detectFormat(header); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}

// detectFormat returns an error for unsupported formats based on magic bytes.
func detectFormat(header []byte) error {
	if len(header) == 0 {
		return fmt.Errorf("%w: empty file", ErrUnsupportedFormat)
	}

	// ZIP: PK\x03\x04
	if len(header) >= 4 && header[0] == 'P' && header[1] == 'K' && header[2] == 0x03 && header[3] == 0x04 {
		return fmt.Errorf("%w: ZIP-compressed files are not yet supported (planned for Phase 4)", ErrUnsupportedFormat)
	}

	// AES: 9-byte prefix "PORTFOLIO"
	if len(header) >= 9 && string(header[:9]) == "PORTFOLIO" {
		return fmt.Errorf("%w: AES-encrypted files are not yet supported (planned for Phase 4)", ErrUnsupportedFormat)
	}

	// Plain XML: first byte is '<'
	if header[0] == '<' {
		return nil
	}

	return fmt.Errorf("%w: unrecognised file format (first byte: 0x%02x)", ErrUnsupportedFormat, header[0])
}
