package logging

import (
	"log/slog"
	"os"
)

func ConfigureLogger(logLevel string) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(logLevel))
	if err != nil {
		slog.Error("error parsing log level. Setting log level to INFO.", "err", err)
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}))
	slog.SetDefault(logger)

	slog.Info("configured logger", "level", level.String())
}
