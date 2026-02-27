// Package sugared is a log driver for sugared logger.
package sugared

import (
	klog "github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aide-family/magicbox/log"
)

var _ log.Interface = (*sugaredLogger)(nil)
var _ log.Driver = (*initializer)(nil)

// LoggerDriver is a log driver for sugared logger.
func LoggerDriver(config Config) log.Driver {
	return &initializer{config: config}
}

type initializer struct {
	config Config
}

func (i *initializer) New() (log.Interface, error) {
	zapLevel, err := zap.ParseAtomicLevel(i.config.GetLevel().String())
	if err != nil {
		zapLevel = zap.NewAtomicLevel()
	}
	zapCfg := zap.Config{
		Level:             zapLevel,
		Development:       i.config.IsDev(),
		DisableCaller:     !i.config.GetEnableCaller(),
		DisableStacktrace: !i.config.GetEnableStack(),
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: string(i.config.GetFormat()),
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "ns",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel: func() zapcore.LevelEncoder {
				if i.config.GetEnableColor() {
					return zapcore.CapitalColorLevelEncoder
				}
				return zapcore.LowercaseLevelEncoder
			}(),
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{i.config.GetOutput()},
		ErrorOutputPaths: []string{i.config.GetOutput()},
	}
	logger, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}
	return &sugaredLogger{logger: logger.Sugar()}, nil
}

type sugaredLogger struct {
	logger *zap.SugaredLogger
}

func (s *sugaredLogger) Log(level klog.Level, keyvals ...any) error {
	isEven := len(keyvals)%2 == 0
	msg := ""
	kvs := keyvals
	if !isEven {
		msg, _ = keyvals[0].(string)
		kvs = keyvals[1:]
	}
	switch level {
	case klog.LevelDebug:
		s.logger.Debugw(msg, kvs...)
	case klog.LevelInfo:
		s.logger.Infow(msg, kvs...)
	case klog.LevelWarn:
		s.logger.Warnw(msg, kvs...)
	case klog.LevelError:
		s.logger.Errorw(msg, kvs...)
	case klog.LevelFatal:
		s.logger.Fatalw(msg, kvs...)
	default:
		s.logger.Infow(msg, kvs...)
	}
	return nil
}
