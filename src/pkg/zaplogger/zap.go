package zaplogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CreateDefaultZapLogger creates a logger with default zap configuration.
func CreateDefaultZapLogger(level zapcore.Level) (*zap.Logger, error) {
	lcfg := DefaultZapLoggerConfig
	lcfg.Level = zap.NewAtomicLevelAt(level)
	c, err := lcfg.Build()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// // 自定义时间输出格式.
// func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
// 	enc.AppendString("[" + t.Format("2006-01-02T15:04:05.000") + "]")
// }

// // 自定义日志级别显示.
// func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
// 	enc.AppendString("[" + level.CapitalString() + "]")
// }

// // 自定义文件：行号输出项.
// func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
// 	enc.AppendString("[" + caller.TrimmedPath() + "]")
// }

// DefaultZapLoggerConfig defines default zap logger configuration.
var DefaultZapLoggerConfig = zap.Config{
	Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
	Development: true,
	Sampling: &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	},
	Encoding: "console", // DefaultLogFormat,
	// copied from "zap.NewProductionEncoderConfig" with some updates
	EncoderConfig: zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "", // "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000"),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	},
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
}
