package strategy

import (
	"path"

	"github.com/spf13/viper"
)

type Store struct {
	source []*Groups
	dir    string
	vi     *viper.Viper
}

// NewStrategyStore 初始化规则存储
func NewStrategyStore(dir string) *Store {
	vi := viper.New()
	vi.SetConfigPermissions(0644)
	return &Store{
		dir: dir,
		vi:  vi,
	}
}

// Store 存储策略
func (l *Store) Store(groups ...*Groups) error {
	l.source = groups

	for _, strategyGroups := range l.source {
		for _, strategyGroup := range strategyGroups.Groups {
			filename := strategyGroup.Name + ".yaml"
			l.vi.SetConfigFile(filename)
			l.vi.Set("groups", strategyGroup)
			// 把规则组写到配置文件中
			writePath := path.Join(l.dir, filename)
			if err := l.vi.WriteConfigAs(writePath); err != nil {
				return err
			}
		}
	}

	return nil
}
