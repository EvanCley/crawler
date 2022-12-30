package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
)

type Plugin = zapcore.Core // 把 zapcore.Core 作为一个自定义类型 Plugin ，zapcore.Core 定义了日志的编码格式以及输出位置等核心功能。

// 最后，我们要暴露一个通用的函数 NewLogger 来生成 logger。NOTE: 一些option选项是无法覆盖的
func NewLogger(plugin zapcore.Core, options ...zap.Option) *zap.Logger {
	return zap.New(plugin, append(DefaultOption(), options...)...)
}

func NewPlugin(writer zapcore.WriteSyncer, enabler zapcore.LevelEnabler) Plugin {
	return zapcore.NewCore(DefaultEncoder(), writer, enabler)
}

// 实现了 NewStdoutPlugin、NewStderrPlugin、NewFilePlugin 这三个函数，分别对应了输出日志到 stdout、stderr 和文件中。
// 这三个函数最终都调用了 zapcore.NewCore 函数。

func NewStdoutPlugin(enabler zapcore.LevelEnabler) Plugin {
	return NewPlugin(zapcore.Lock(zapcore.AddSync(os.Stdout)), enabler)
}

func NewStderrPlugin(enabler zapcore.LevelEnabler) Plugin {
	return NewPlugin(zapcore.Lock(zapcore.AddSync(os.Stderr)), enabler)
}

func NewFilePlugin(filePath string, enabler zapcore.LevelEnabler) (Plugin, io.Closer) {
	// filePath 表示输出文件的路径，而 enabler 代表当前环境中要打印的日志级别
	var writer = DefaultLumberjackLogger()
	writer.Filename = filePath
	return NewPlugin(zapcore.AddSync(writer), enabler), writer
}
