package logger

import (
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(filename string, debug bool) error {
	logFile, err := os.OpenFile(filepath.Clean(filename)+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	encoderConfig := zap.NewProductionEncoderConfig()

	encoderConfig.TimeKey = "time"
	encoderConfig.ConsoleSeparator = " | "
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	stdoutEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), level),
		zapcore.NewCore(stdoutEncoder, zapcore.AddSync(os.Stdout), level),
	)

	zap.ReplaceGlobals(zap.New(core))
	return err
}

func Sync() {
	err := zap.L().Sync()
	if err != nil {
		log.Fatal("Failed to sync logger data")
	}
}
