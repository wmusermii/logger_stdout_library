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

type Config struct {
	Output      OutputType
	LogDir      string
	ServiceName string // baru: nama file log, dikontrol dari .env
}

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

		serviceName := strings.TrimSpace(os.Getenv("LOG_SERVICE_NAME"))
		if serviceName == "" {
			return Config{}, fmt.Errorf("LOG_SERVICE_NAME wajib diisi ketika LOG_OUTPUT=file")
		}
		cfg.ServiceName = serviceName
	}

	return cfg, nil
}
