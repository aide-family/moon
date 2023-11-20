package strategy

import (
	"path"

	"github.com/spf13/viper"
)

type Store struct {
	source []*Groups
	dir    string
}

// NewStrategyStore 初始化规则存储
func NewStrategyStore(dir string) *Store {
	return &Store{
		dir: dir,
	}
}

// Store 存储策略
func (l *Store) Store(groups ...*Groups) error {
	l.source = groups
	vi := viper.New()
	vi.SetConfigPermissions(0644)

	for _, strategyGroups := range l.source {
		for _, strategyGroup := range strategyGroups.Groups {
			filename := strategyGroup.Name + ".yaml"
			vi.SetConfigFile(filename)
			vi.Set("groups", strategyGroup)
			// 把规则组写到配置文件中
			writePath := path.Join(l.dir, filename)
			if err := vi.WriteConfigAs(writePath); err != nil {
				return err
			}
		}
	}

	return nil
}
