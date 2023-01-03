package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// DefaultEncoderConfig 默认的一些配置
func DefaultEncoderConfig() zapcore.EncoderConfig {
	var encoderConfig = zap.NewProductionEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return encoderConfig
}

// 统一用 json
func DefaultEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(DefaultEncoderConfig())
}

// DefaultOption 默认的选项会打印调用时的文件与行号，并且只有当日志等级在 DPanic 等级之上时，才输出函数的堆栈信息。
func DefaultOption() []zap.Option {
	var stackTraceLevel zap.LevelEnablerFunc = func(level zapcore.Level) bool {
		return level >= zap.DPanicLevel
	}
	return []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(stackTraceLevel),
	}
}

// 我们借助 lumberjack.v2 来完成日志的切割。
// 1.不会自动清理backup
// 2.每200mb压缩一次，不按时间rotate
func DefaultLumberjackLogger() *lumberjack.Logger {
	return &lumberjack.Logger{
		MaxSize:   200,
		LocalTime: true,
		Compress:  true,
	}
}
