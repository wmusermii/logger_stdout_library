package loggerstdoutlibrary

import "time"

// NativeCallHandle adalah handle ke satu native call yang sedang berjalan,
// dikembalikan dari BeginNative supaya bisa di-update/finish belakangan.
type NativeCallHandle struct {
	parent    *MainLog
	call      *NativeCall
	startedAt time.Time
}

// BeginNativeParams parameter untuk memulai pemanggilan ke service/sistem lain
type BeginNativeParams struct {
	Type string // contoh: "HTTP", "GRPC", "DB", "KAFKA"
	URL  string
}

// BeginNative dipanggil ketika flow utama mulai memanggil service lain.
// Native call langsung ditambahkan ke dalam nativeCalls milik MainLog.
func (m *MainLog) BeginNative(p BeginNativeParams) *NativeCallHandle {
	now := time.Now()
	nc := &NativeCall{
		Type:      p.Type,
		URL:       p.URL,
		StartTime: now.Format(time.RFC3339Nano),
	}

	m.mu.Lock()
	m.entry.NativeCalls = append(m.entry.NativeCalls, nc)
	m.mu.Unlock()

	return &NativeCallHandle{
		parent:    m,
		call:      nc,
		startedAt: now,
	}
}

// UpdateNative dipanggil ketika ada perubahan data di tengah pemanggilan sub-service
// (misal URL berubah karena redirect, atau menambah info tambahan)
func (h *NativeCallHandle) UpdateNative(url string) {
	h.parent.mu.Lock()
	defer h.parent.mu.Unlock()
	if url != "" {
		h.call.URL = url
	}
}

// FinishNative dipanggil setelah pemanggilan sub-service selesai (success/error)
func (h *NativeCallHandle) FinishNative(responseCode, responseMessage string) {
	end := time.Now()

	h.parent.mu.Lock()
	defer h.parent.mu.Unlock()

	h.call.EndTime = end.Format(time.RFC3339Nano)
	h.call.Duration = durationMs(h.startedAt, end)
	h.call.ResponseCode = responseCode
	h.call.ResponseMessage = responseMessage
}
