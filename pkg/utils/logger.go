package utils

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global Zap logger instance
var Logger *zap.Logger

func init() {
	// Configure Zap logger
	// You can customize this further (e.g., log to a file, different encoding)
	config := zap.Config{
		Encoding:    "json", // or "console"
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		OutputPaths: []string{"stdout"},
		// ErrorOutputPaths: []string{"stderr"}, // You can specify error output paths
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		// Development: true, // Enables more verbose logging, like caller, stacktraces on Warn and above
	}

	var err error
	Logger, err = config.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	// Flushes buffer, if any
	// defer Logger.Sync() // This would be for a main function where logger is used. For a package init, it's not needed like this.

	Logger.Info("Zap logger initialized successfully")
}
