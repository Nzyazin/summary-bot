package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap *zap.SugaredLogger
}

func NewLogger(level string, filePath string) (*Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()

	if filePath != "" {
		logDir := filepath.Dir(filePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	var core zapcore.Core
	if filePath == "" {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		core = zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel)
	} else {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zapLevel)
		core = zapcore.NewTee(consoleCore, fileCore)
	}
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := logger.Sugar()

	return &Logger{zap: sugar}, nil
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.zap.Debugf(format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.zap.Infof(format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.zap.Warnf(format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.zap.Errorf(format, args...)
}
