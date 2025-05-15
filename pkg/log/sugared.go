package log

import (
	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/moon-monitor/moon/pkg/config"
)

var _ log.Logger = (*sugaredLogger)(nil)

type sugaredLogger struct {
	logger *zap.SugaredLogger
}

func (s *sugaredLogger) Log(level log.Level, keyvals ...any) error {
	isEven := len(keyvals)%2 == 0
	msg := ""
	kvs := keyvals
	if !isEven {
		msg, _ = keyvals[0].(string)
		kvs = keyvals[1:]
	}
	switch level {
	case log.LevelDebug:
		s.logger.Debugw(msg, kvs...)
	case log.LevelInfo:
		s.logger.Infow(msg, kvs...)
	case log.LevelWarn:
		s.logger.Warnw(msg, kvs...)
	case log.LevelError:
		s.logger.Errorw(msg, kvs...)
	case log.LevelFatal:
		s.logger.Fatalw(msg, kvs...)
	default:
		s.logger.Infow(msg, kvs...)
	}
	return nil
}

func NewSugaredLogger(isDev bool, level config.Log_Level, cfg *config.Log_SugaredLogConfig) (log.Logger, error) {
	zapLevel, err := zap.ParseAtomicLevel(level.String())
	if err != nil {
		zapLevel = zap.NewAtomicLevel()
	}
	zapCfg := zap.Config{
		Level:             zapLevel,
		Development:       isDev,
		DisableCaller:     cfg.GetDisableCaller(),
		DisableStacktrace: cfg.GetDisableStacktrace(),
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: cfg.GetFormat(),
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
				if cfg.GetEnableColor() {
					return zapcore.CapitalColorLevelEncoder
				}
				return zapcore.LowercaseLevelEncoder
			}(),
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{cfg.GetOutput()},
		ErrorOutputPaths: []string{cfg.GetOutput()},
	}
	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}
	return &sugaredLogger{logger: logger.Sugar()}, nil
}
