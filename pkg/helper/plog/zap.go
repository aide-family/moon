package plog

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	OutputJsonType = "json"
)

type (
	Config interface {
		GetFilename() string
		GetLevel() string
		GetMaxSize() int32
		GetMaxBackups() int32
		GetMaxAge() int32
		GetCompress() bool
		GetLocalTime() bool
		GetEncoder() string
	}
)

func getHook(c Config) *lumberjack.Logger {
	filename := "./server.log"
	maxSize := 100
	maxBackups := 10
	maxAge := 7
	if c.GetMaxBackups() != 0 {
		maxBackups = int(c.GetMaxBackups())
	}
	if c.GetMaxAge() != 0 {
		maxAge = int(c.GetMaxAge())
	}
	if c.GetMaxSize() != 0 {
		maxSize = int(c.GetMaxSize())
	}
	if c.GetFilename() != "" {
		filename = c.GetFilename()
	}
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		LocalTime:  c.GetLocalTime(),
		Compress:   c.GetCompress(),
	}
}

func NewZapLog(c Config) *zap.Logger {
	hook := getHook(c)
	encoderConfig := zapcore.EncoderConfig{
		LevelKey:       "level",                        // 日志 level 的 key
		TimeKey:        "time",                         // 日志时间 的 key
		LineEnding:     zapcore.DefaultLineEnding,      // 回车符换行
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // level大写: Info, Debug,Warn等
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // 时间格式: "2006-01-02T15:04:05.000Z0700"
		EncodeDuration: zapcore.SecondsDurationEncoder, // 时间戳用float64型,更加准确, 另一种是NanosDurationEncoder int64
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 产生日志文件的路径格式: 包名/文件名:行号
	}

	var syncers []zapcore.WriteSyncer // io writer
	syncers = append(syncers, zapcore.AddSync(hook))

	zapLevel := getLevel(c.GetLevel())

	atomicLevel := zap.NewAtomicLevel() // 设置日志 level
	atomicLevel.SetLevel(zapLevel)      // 打印 debug, info, warn,error, panic,fetal 全部级别日志

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if c.GetEncoder() == OutputJsonType {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(syncers...), atomicLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zapLevel),
	}

	zapOpts := []zap.Option{
		zap.AddCaller(),   //日志打印输出 文件名, 行号, 函数名
		zap.Development(), // 可输出 d-panic, panic 级别的日志
		zap.AddCallerSkip(3),
	}

	return zap.New(zapcore.NewTee(cores...), zapOpts...)
}

func getLevel(level string) zapcore.Level {
	switch level {
	case "debug", "DEBUG":
		return zap.DebugLevel
	case "info", "INFO":
		return zap.InfoLevel
	case "warn", "WARN":
		return zap.WarnLevel
	case "error", "ERROR":
		return zap.ErrorLevel
	case "panic", "PANIC":
		return zap.PanicLevel
	case "fatal", "FATAL":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
