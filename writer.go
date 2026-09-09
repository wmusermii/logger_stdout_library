package loggerstdoutlibrary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Writer interface {
	Write(b []byte) error
}

type StdoutWriter struct{}

func (s *StdoutWriter) Write(b []byte) error {
	_, err := os.Stdout.Write(append(b, '\n'))
	return err
}

type fileHandle struct {
	file *os.File
	date string
}

// FileWriter sekarang menulis SEMUA log ke satu file yang namanya
// ditentukan oleh LOG_SERVICE_NAME di .env, terlepas dari serviceName
// yang dipakai di tiap BeginMain. Tetap rotasi harian.
type FileWriter struct {
	mu          sync.Mutex
	dir         string
	serviceName string
	current     *fileHandle
}

func NewFileWriter(dir, serviceName string) (*FileWriter, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("logger_stdout_library: gagal membuat direktori log %q: %w", dir, err)
	}
	return &FileWriter{
		dir:         dir,
		serviceName: serviceName,
	}, nil
}

func (f *FileWriter) Write(b []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	if f.current == nil || f.current.date != today {
		if f.current != nil && f.current.file != nil {
			f.current.file.Close()
		}

		filename := fmt.Sprintf("%s-%s.log", sanitizeFileName(f.serviceName), today)
		fullPath := filepath.Join(f.dir, filename)

		file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("logger_stdout_library: gagal membuka file log %q: %w", fullPath, err)
		}

		f.current = &fileHandle{file: file, date: today}
	}

	_, err := f.current.file.Write(append(b, '\n'))
	return err
}

func (f *FileWriter) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.current != nil && f.current.file != nil {
		return f.current.file.Close()
	}
	return nil
}

func sanitizeFileName(s string) string {
	r := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ":", "-")
	return r.Replace(s)
}
