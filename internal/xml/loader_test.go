package xml

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, data []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "pp-test-*.xml")
	require.NoError(t, err)
	_, err = f.Write(data)
	require.NoError(t, err)
	f.Close()
	return f.Name()
}

func TestLoad_PlainXML(t *testing.T) {
	path := writeTemp(t, []byte(`<?xml version="1.0"?><client/>`))
	r, err := Load(path)
	require.NoError(t, err)
	r.Close()
}

func TestLoad_ZIPFormat(t *testing.T) {
	path := writeTemp(t, []byte{0x50, 0x4B, 0x03, 0x04, 0x00})
	_, err := Load(path)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedFormat))
	assert.Contains(t, err.Error(), "ZIP")
}

func TestLoad_AESFormat(t *testing.T) {
	path := writeTemp(t, []byte("PORTFOLIOsomedata"))
	_, err := Load(path)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedFormat))
	assert.Contains(t, err.Error(), "AES")
}

func TestLoad_UnknownFormat(t *testing.T) {
	path := writeTemp(t, []byte{0xFF, 0xFE, 0x00})
	_, err := Load(path)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedFormat))
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nonexistent.xml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}
