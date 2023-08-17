package strategystore

import (
	"path"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/viper"

	"prometheus-manager/api/strategy"
	"prometheus-manager/pkg/util/dir"
)

type Strategy struct {
	source       []*strategy.StrategyDir
	loggerHelper *log.Helper
}

// NewStrategy 初始化配置文件
func NewStrategy(source []*strategy.StrategyDir, loggerHelper *log.Helper) *Strategy {
	return &Strategy{
		source:       source,
		loggerHelper: loggerHelper,
	}
}

// StoreStrategy 存储策略
func (l *Strategy) StoreStrategy() ([]*strategy.StrategyDir, error) {
	viper.SetConfigType("yaml")

	list := make([]*strategy.StrategyDir, 0, len(l.source))

	for _, strategyDir := range l.source {
		isDir, err := dir.IsDir(strategyDir.GetDir())
		if err != nil {
			return nil, err
		}

		if !isDir {
			l.loggerHelper.Warnf("this is not dir: %s", strategyDir.GetDir())
			list = append(list, strategyDir)
			continue
		}

		strategyTempList := make([]*strategy.Strategy, 0, len(strategyDir.GetStrategies()))
		for _, strategyInfo := range strategyDir.GetStrategies() {
			if !isYamlFile(strategyInfo.GetFilename()) {
				strategyTempList = append(strategyTempList, strategyInfo)
				l.loggerHelper.Warnf("this is not yaml file: %s", strategyInfo.GetFilename())
				continue
			}
			writePath := path.Join(strategyDir.GetDir(), strategyInfo.GetFilename())

			if err != nil {
				return nil, err
			}

			viper.SetConfigFile(writePath)
			viper.SetConfigPermissions(0644)
			viper.Set("groups", strategyInfo.GetGroups())

			if err := viper.WriteConfig(); err != nil {
				return nil, err
			}
		}

		if len(strategyTempList) > 0 {
			list = append(list, &strategy.StrategyDir{
				Dir:        strategyDir.GetDir(),
				Strategies: strategyTempList,
			})
		}

	}

	return list, nil
}

// isYamlFile 判断是否是yaml文件
func isYamlFile(filename string) bool {
	return strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
}
