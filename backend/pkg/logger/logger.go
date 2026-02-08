package logger

import (
	"os"

	"videodownload/internal/model"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// Init initializes the logger
func Init(cfg *model.LoggingConfig) error {
	// Create log directory if not exists
	if err := os.MkdirAll("./log", 0755); err != nil {
		return err
	}

	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(cfg.Level)); err != nil {
		logLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{cfg.FilePath, "stdout"},
		ErrorOutputPaths: []string{cfg.FilePath, "stderr"},
	}

	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// Sync flushes the logger
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}
