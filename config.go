package loggerstdoutlibrary

import (
	"fmt"
	"os"
	"strings"
)

type OutputType string

const (
	OutputStdout OutputType = "stdout"
	OutputFile   OutputType = "file"
)

// Config adalah konfigurasi output logger.
type Config struct {
	Output OutputType
	LogDir string // wajib diisi kalau Output == file
}

// LoadConfigFromEnv membaca konfigurasi dari environment variable:
//
//	LOG_OUTPUT = "stdout" | "file"   (default: "stdout" jika kosong)
//	LOG_DIR    = path direktori tujuan log (wajib kalau LOG_OUTPUT=file,
//	             boleh berupa path ke shared/mounted network drive)
func LoadConfigFromEnv() (Config, error) {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_OUTPUT")))
	if raw == "" {
		raw = string(OutputStdout)
	}

	var output OutputType
	switch raw {
	case string(OutputStdout):
		output = OutputStdout
	case string(OutputFile):
		output = OutputFile
	default:
		return Config{}, fmt.Errorf("LOG_OUTPUT tidak valid: %q (harus 'stdout' atau 'file')", raw)
	}

	cfg := Config{Output: output}

	if output == OutputFile {
		dir := strings.TrimSpace(os.Getenv("LOG_DIR"))
		if dir == "" {
			return Config{}, fmt.Errorf("LOG_DIR wajib diisi ketika LOG_OUTPUT=file")
		}
		cfg.LogDir = dir
	}

	return cfg, nil
}
