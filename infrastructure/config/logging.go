package config

import "log/slog"

type Logging struct {
	AddSource bool       `env:"LOG_ADD_SOURCE" name:"add-source" default:"false"`
	Level     slog.Level `env:"LOG_LEVEL" name:"level" default:"info" help:"Log level to apply"`
}

func (l Logging) Options() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level:     l.Level,
		AddSource: l.AddSource,
	}
}
