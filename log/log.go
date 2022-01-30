package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/xwjdsh/freeproxy/config"
)

var logger *zap.Logger

func Init(cfg *config.LogConfig) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(encoder, os.Stdout, cfg.Level)
	logger = zap.New(core, zap.AddCaller())
}

func L() *zap.Logger {
	return logger
}
