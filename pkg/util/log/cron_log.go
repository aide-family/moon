package log

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
)

type cronLog struct {
	log log.Logger
}

func (c *cronLog) Info(msg string, keysAndValues ...interface{}) {
	_ = c.log.Log(log.LevelInfo, append([]any{"msg", msg}, keysAndValues...))
}

func (c *cronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	_ = c.log.Log(log.LevelError, append([]any{"err", err, "msg", msg}, keysAndValues...))
}

// NewCronLog cron日志
func NewCronLog(logger log.Logger) cron.Logger {
	return &cronLog{
		log: logger,
	}
}
