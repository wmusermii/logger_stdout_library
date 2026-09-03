package loggerstdoutlibrary

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func nowISO() string {
	return time.Now().Format(time.RFC3339Nano)
}

func severityNumber(text string) string {
	switch text {
	case "TRACE":
		return "1"
	case "DEBUG":
		return "2"
	case "INFO":
		return "3"
	case "WARN":
		return "4"
	case "ERROR":
		return "5"
	case "FATAL":
		return "6"
	default:
		return "0"
	}
}

func genHexID(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// GenTraceID & GenSpanID dipakai kalau caller tidak supply traceId/spanId sendiri
func GenTraceID() string { return genHexID(16) } // 32 hex char
func GenSpanID() string  { return genHexID(8) }  // 16 hex char

func durationMs(start, end time.Time) float64 {
	return float64(end.Sub(start).Microseconds()) / 1000.0
}
