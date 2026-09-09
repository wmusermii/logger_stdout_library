package loggerstdoutlibrary

import "sync"

var (
	globalMu     sync.RWMutex
	globalWriter Writer = &StdoutWriter{} // default: tetap stdout kalau Init() tidak dipanggil
)

// Init membaca konfigurasi dari environment variable (LOG_OUTPUT, LOG_DIR)
// dan menyiapkan writer yang sesuai. WAJIB dipanggil sekali di awal main(),
// sebelum BeginMain pertama kali dipakai.
//
// Kalau Init() tidak dipanggil sama sekali, logger tetap jalan seperti
// sebelumnya (output ke stdout) — jadi tidak ada breaking change untuk
// service yang belum mau pakai fitur file logging.
func Init() error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}
	return InitWithConfig(cfg)
}

// InitWithConfig sama seperti Init, tapi config di-supply manual (berguna
// untuk unit test, atau kalau service tidak mau bergantung pada env var).
func InitWithConfig(cfg Config) error {
	var w Writer

	switch cfg.Output {
	case OutputFile:
		fw, err := NewFileWriter(cfg.LogDir)
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

// Close menutup writer aktif (relevan untuk FileWriter, supaya file
// ter-flush dengan benar). Panggil via defer di main().
func Close() error {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if fw, ok := globalWriter.(*FileWriter); ok {
		return fw.Close()
	}
	return nil
}
