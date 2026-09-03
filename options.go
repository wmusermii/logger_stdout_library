package loggerstdoutlibrary

type Option func(*StdoutLogger)

func WithMinLevel(level Level) Option {
	return func(l *StdoutLogger) {
		l.minLevel = level
	}
}

func WithService(name string) Option {
	return func(l *StdoutLogger) {
		l.service = name
	}
}

func WithEnv(env string) Option {
	return func(l *StdoutLogger) {
		l.env = env
	}
}
