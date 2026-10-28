package bootstrap

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3/config"

	"github.com/aide-family/moon/pkg/merr"
)

// Load 读取 Bootstrap YAML 并展开环境变量。各服务只保留自己的 conf 类型。
func Load[T any](path string, logger *slog.Logger) (*T, error) {
	dotenvPath, err := loadDotEnv(path)
	if err != nil {
		logger.Error("dotenv load failed", slog.String("error", err.Error()))
		return nil, merr.Server("config load failed")
	}
	if dotenvPath != "" {
		logger.Info("dotenv file loaded", slog.String("path", dotenvPath))
	}

	c := config.New(config.WithSource(NewSource(path)))
	defer c.Close()
	if err := c.Load(); err != nil {
		logger.Error("config load failed", slog.String("error", err.Error()))
		return nil, merr.Server("config load failed")
	}
	var out T
	if err := c.Scan(&out); err != nil {
		logger.Error("config scan failed", slog.String("error", err.Error()))
		return nil, merr.Server("config load failed")
	}
	return &out, nil
}
