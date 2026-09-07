package loggerstdoutlibrary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Writer adalah abstraksi tujuan tulis log. Memungkinkan nanti nambah
// jenis output lain (misal: syslog, elastic) tanpa ubah mainlog.go.
type Writer interface {
	Write(b []byte, serviceName string) error
}

// ---------- StdoutWriter ----------

type StdoutWriter struct{}

func (s *StdoutWriter) Write(b []byte, _ string) error {
	_, err := os.Stdout.Write(append(b, '\n'))
	return err
}

// ---------- FileWriter ----------

type fileHandle struct {
	file *os.File
	date string // format "2006-01-02", untuk deteksi kapan harus rotate
}

// FileWriter menulis tiap service ke file terpisah, dan otomatis
// membuka file baru setiap hari (rotasi harian).
// Aman dipakai untuk banyak service sekaligus dalam satu direktori
// (misal shared network drive) karena file di-key oleh serviceName.
type FileWriter struct {
	mu    sync.Mutex
	dir   string
	files map[string]*fileHandle
}

// NewFileWriter menyiapkan direktori tujuan (dibuat otomatis kalau belum ada).
func NewFileWriter(dir string) (*FileWriter, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("logger_stdout_library: gagal membuat direktori log %q: %w", dir, err)
	}
	return &FileWriter{
		dir:   dir,
		files: make(map[string]*fileHandle),
	}, nil
}

func (f *FileWriter) Write(b []byte, serviceName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if serviceName == "" {
		serviceName = "unknown-service"
	}
	today := time.Now().Format("2006-01-02")

	fh, exists := f.files[serviceName]
	if !exists || fh.date != today {
		if exists && fh.file != nil {
			fh.file.Close() // tutup file hari sebelumnya
		}

		filename := fmt.Sprintf("%s-%s.log", sanitizeFileName(serviceName), today)
		fullPath := filepath.Join(f.dir, filename)

		file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("logger_stdout_library: gagal membuka file log %q: %w", fullPath, err)
		}

		fh = &fileHandle{file: file, date: today}
		f.files[serviceName] = fh
	}

	_, err := fh.file.Write(append(b, '\n'))
	return err
}

// Close menutup semua file handle yang sedang terbuka.
// Panggil ini saat service akan shutdown (misal via defer di main()).
func (f *FileWriter) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var lastErr error
	for _, fh := range f.files {
		if err := fh.file.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func sanitizeFileName(s string) string {
	r := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-")
	return r.Replace(s)
}
