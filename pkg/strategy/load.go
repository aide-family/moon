package strategy

import (
	"bytes"
	"path/filepath"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/spf13/viper"
)

type Load struct {
	source config.Source
	kvs    []*config.KeyValue
	vi     *viper.Viper
}

// NewStrategyLoad 初始化规则加载器
func NewStrategyLoad(source config.Source) *Load {
	vi := viper.New()
	return &Load{
		source: source,
		vi:     vi,
	}
}

func (l *Load) load() error {
	load, err := l.source.Load()
	if err != nil {
		return err
	}
	l.kvs = load
	return nil
}

func (l *Load) Load() ([]*Group, error) {
	if err := l.load(); err != nil {
		return nil, err
	}

	list := make([]*Group, 0)
	for _, kv := range l.kvs {
		var groups Groups
		filename := kv.Key
		ext, ok := isYaml(filename)
		if !ok {
			continue
		}
		l.vi.SetConfigType(ext)
		if err := l.vi.ReadConfig(bytes.NewBuffer(kv.Value)); err != nil {
			return nil, err
		}
		if err := l.vi.Unmarshal(&groups); err != nil {
			return nil, err
		}

		list = append(list, groups.Groups...)
	}

	return list, nil
}

func isYaml(filename string) (string, bool) {
	ext := filepath.Ext(filename)
	if ext == ".yaml" || ext == ".yml" {
		return ext[1:], true
	}
	return "", false
}
