package loggerstdoutlibrary

import "testing"

func TestFullFlow(t *testing.T) {
	m := BeginMain(BeginMainParams{
		EventName:     "TestFlow",
		EventCategory: "test",
		ServiceName:   "test-service",
	})

	nc := m.BeginNative(BeginNativeParams{Type: "HTTP", URL: "https://example.com"})
	nc.FinishNative("200", "OK")

	m.FinishMain("200", "success")

	if len(m.entry.NativeCalls) != 1 {
		t.Fatalf("expected 1 native call, got %d", len(m.entry.NativeCalls))
	}
	if m.entry.NativeCalls[0].ResponseCode != "200" {
		t.Fatalf("unexpected response code")
	}
}
