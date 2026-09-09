package loggerstdoutlibrary

import "sync"

var (
	globalMu     sync.RWMutex
	globalWriter Writer = &StdoutWriter{} // default: stdout kalau Init() belum dipanggil
)

// Init membaca konfigurasi dari environment variable (LOG_OUTPUT, LOG_DIR,
// LOG_SERVICE_NAME) dan menyiapkan writer yang sesuai. WAJIB dipanggil
// sekali di awal main(), sebelum BeginMain pertama kali dipakai.
func Init() error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}
	return InitWithConfig(cfg)
}

// InitWithConfig sama seperti Init, tapi config di-supply manual.
func InitWithConfig(cfg Config) error {
	var w Writer

	switch cfg.Output {
	case OutputFile:
		fw, err := NewFileWriter(cfg.LogDir, cfg.ServiceName)
		if err != nil {
			return err
		}
		w = fw
	default:
		w = &StdoutWriter{}
	}

	globalMu.Lock()
	globalWriter = w
	globalMu.Unlock()
	return nil
}

func getWriter() Writer {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalWriter
}

// Close menutup writer aktif (relevan untuk FileWriter).
func Close() error {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if fw, ok := globalWriter.(*FileWriter); ok {
		return fw.Close()
	}
	return nil
}
