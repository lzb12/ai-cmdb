package logger

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"cmdb-server/internal/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(cfg config.LogConfig) (*zap.Logger, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.Filename), 0o755); err != nil {
		return nil, err
	}

	level := parseLevel(cfg.Level)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	})

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
	cores := []zapcore.Core{
		zapcore.NewCore(jsonEncoder, fileWriter, level),
	}

	if cfg.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))
	}

	core := zapcore.NewTee(cores...)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)), nil
}

func parseLevel(value string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
