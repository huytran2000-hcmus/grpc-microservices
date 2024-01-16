package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(lvl zapcore.Level) (*zap.Logger, error) {
	cfg := defaultZapConfig()
	cfg.Level.SetLevel(lvl)

	return cfg.Build()
}

func defaultZapConfig() zap.Config {
	cfg := zap.NewProductionConfig()

	cfg.Encoding = "json"

	cfg.EncoderConfig.NameKey = "logger"
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.StacktraceKey = "stacktrace"
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	return cfg
}
