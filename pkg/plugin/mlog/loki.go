package mlog

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/grafana/loki-client-go/loki"
	"github.com/prometheus/common/model"
)

// NewLokiLogger create a new loki logger
func NewLokiLogger(c LokiConfig) Logger {
	config, err := loki.NewDefaultConfig(c.GetUrl())
	if err != nil {
		panic(err)
	}
	client, err := loki.New(config)
	if err != nil {
		panic(err)
	}
	return &lokiLog{
		client: client,
	}
}

type (
	// LokiConfig 日志配置
	LokiConfig interface {
		GetUrl() string
	}

	lokiLog struct {
		client *loki.Client
	}
)

// Log 日志
func (l *lokiLog) Log(level log.Level, keyvals ...interface{}) error {
	var (
		msg    = ""
		keyLen = len(keyvals)
	)
	if len(keyvals) == 0 {
		return nil
	}
	if keyLen%2 != 0 {
		msg = fmt.Sprintf("%v", keyvals[len(keyvals)-1])
		keyLen--
	}
	labels := make(model.LabelSet)
	for i := 0; i < keyLen; i += 2 {
		if keyvals[i].(string) == "msg" {
			msg, _ = keyvals[i+1].(string)
			continue
		}
		labels[model.LabelName(keyvals[i].(string))] = model.LabelValue(fmt.Sprint(keyvals[i+1]))
	}
	labels = labels.Merge(model.LabelSet{"level": model.LabelValue(level.String())})
	return l.client.Handle(labels, time.Now(), msg)
}
