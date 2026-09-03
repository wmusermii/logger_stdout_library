package loggerstdoutlibrary

import (
	"encoding/json"
	"os"
	"time"
)

// StdoutLogger adalah implementasi Logger yang menulis ke os.Stdout
type StdoutLogger struct {
	minLevel Level
	service  string
	env      string
}

// NewStdoutLogger membuat instance baru StdoutLogger
func NewStdoutLogger(opts ...Option) *StdoutLogger {
	l := &StdoutLogger{
		minLevel: InfoLevel, // default
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

type logEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service,omitempty"`
	Env       string                 `json:"env,omitempty"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

func (l *StdoutLogger) log(level Level, msg string, fields ...Field) {
	if level < l.minLevel {
		return
	}

	fieldMap := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		fieldMap[f.Key] = f.Value
	}

	entry := logEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Service:   l.service,
		Env:       l.env,
		Message:   msg,
		Fields:    fieldMap,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		os.Stdout.WriteString(msg + "\n")
		return
	}
	os.Stdout.Write(append(b, '\n'))
}

func (l *StdoutLogger) Debug(msg string, fields ...Field) { l.log(DebugLevel, msg, fields...) }
func (l *StdoutLogger) Info(msg string, fields ...Field)  { l.log(InfoLevel, msg, fields...) }
func (l *StdoutLogger) Warn(msg string, fields ...Field)  { l.log(WarnLevel, msg, fields...) }
func (l *StdoutLogger) Error(msg string, fields ...Field) { l.log(ErrorLevel, msg, fields...) }
