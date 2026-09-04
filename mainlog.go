package loggerstdoutlibrary

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// MainLog adalah instance yang hidup selama satu flow service utama berjalan
type MainLog struct {
	mu        sync.Mutex
	entry     MainEntry
	startedAt time.Time
	finished  bool
}

// BeginMainParams adalah parameter untuk memulai satu flow trace
type BeginMainParams struct {
	TraceID          string // opsional, auto-generate jika kosong
	SpanID           string // opsional, auto-generate jika kosong
	SeverityText     string // default: "INFO"
	EventName        string
	EventCategory    string
	Body             interface{}
	ServiceName      string
	ServiceVersion   string
	OperationName    string
	OperationVersion string
	LocalIP          string
	Channel          string
	Layer            string
	ClientIP         string
}

// BeginMain memulai satu flow log. Panggil ini di awal service/handler.
func BeginMain(p BeginMainParams) *MainLog {
	if p.SeverityText == "" {
		p.SeverityText = "INFO"
	}
	traceID := p.TraceID
	if traceID == "" {
		traceID = GenTraceID()
	}
	spanID := p.SpanID
	if spanID == "" {
		spanID = GenSpanID()
	}

	now := time.Now()
	m := &MainLog{
		startedAt: now,
		entry: MainEntry{
			TraceID:        traceID,
			SpanID:         spanID,
			SeverityText:   p.SeverityText,
			SeverityNumber: severityNumber(p.SeverityText),
			Timestamp:      nowISO(),
			Event: Event{
				EventName:     p.EventName,
				EventCategory: p.EventCategory,
			},
			Body: p.Body,
			Resource: Resource{
				ServiceName:      p.ServiceName,
				ServiceVersion:   p.ServiceVersion,
				OperationName:    p.OperationName,
				OperationVersion: p.OperationVersion,
				LocalIP:          p.LocalIP,
			},
			Attributes: Attributes{
				Channel:  p.Channel,
				Layer:    p.Layer,
				ClientIP: p.ClientIP,
			},
			StartTime:   now.Format(time.RFC3339Nano),
			NativeCalls: []*NativeCall{},
		},
	}
	return m
}

// UpdateMainParams — hanya field yang mau diubah, sisanya nil/kosong akan diabaikan
type UpdateMainParams struct {
	Body          interface{}
	SeverityText  string
	EventName     string
	EventCategory string
}

// UpdateMain dipanggil di tengah flow ketika ada nilai baru yang perlu ditambahkan
func (m *MainLog) UpdateMain(p UpdateMainParams) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p.Body != nil {
		m.entry.Body = p.Body
	}
	if p.SeverityText != "" {
		m.entry.SeverityText = p.SeverityText
		m.entry.SeverityNumber = severityNumber(p.SeverityText)
	}
	if p.EventName != "" {
		m.entry.Event.EventName = p.EventName
	}
	if p.EventCategory != "" {
		m.entry.Event.EventCategory = p.EventCategory
	}
}

// FinishMain dipanggil setelah flow utama selesai, lalu langsung print ke stdout
func (m *MainLog) FinishMain(responseCode, responseMessage string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.finished {
		return // guard supaya tidak dobel print
	}
	end := time.Now()

	m.entry.ResponseCode = responseCode
	m.entry.ResponseMessage = responseMessage
	m.entry.EndTime = end.Format(time.RFC3339Nano)
	m.entry.Duration = formatDuration(durationMs(m.startedAt, end))
	m.finished = true

	m.write()
}

func (m *MainLog) write() {
	b, err := json.Marshal(m.entry)
	if err != nil {
		os.Stdout.WriteString("logger_stdout_library: marshal error: " + err.Error() + "\n")
		return
	}
	os.Stdout.Write(append(b, '\n'))
}

func (m *MainLog) SafeFinishOnPanic() {
	if r := recover(); r != nil {
		m.FinishMain("500", fmt.Sprintf("panic recovered: %v", r))
		panic(r) // re-throw, jangan ditelan diam-diam
	}
}

func formatDuration(ms float64) string {
	return strconv.FormatFloat(ms, 'f', 3, 64) // contoh: "123.456"
}
