package strategyload

import (
	"bytes"
	"context"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api/strategy"
)

type Strategy struct {
	source config.Source
	kvs    []*config.KeyValue
}

// NewStrategy 初始化配置文件
func NewStrategy(source config.Source) *Strategy {
	return &Strategy{
		source: source,
	}
}

func (l *Strategy) load() error {
	load, err := l.source.Load()
	if err != nil {
		return err
	}
	l.kvs = load
	return nil
}

func (l *Strategy) Scan(ctx context.Context, v *[]*strategy.Strategy) error {
	_, span := otel.Tracer("strategyload/load").Start(ctx, "Strategy.Scan")
	defer span.End()
	if err := l.load(); err != nil {
		return err
	}

	viper.SetConfigType("yaml")

	for _, kv := range l.kvs {
		var tmp strategy.Strategy
		if err := viper.ReadConfig(bytes.NewBuffer(kv.Value)); err != nil {
			return err
		}

		if err := viper.Unmarshal(&tmp); err != nil {
			return err
		}

		tmp.Filename = kv.Key

		*v = append(*v, &tmp)
	}

	return nil
}
