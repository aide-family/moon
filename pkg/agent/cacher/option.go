package cacher

import (
	"github.com/aide-family/moon/pkg"
	"github.com/nutsdb/nutsdb"
	"github.com/redis/go-redis/v9"
)

func WithRedis(client *redis.Client) Option {
	return func(c *cacheBuild) {
		if pkg.IsNil(c.redisCli) {
			return
		}
		c.redisCli = client
	}
}

func WithNutsDb(db *nutsdb.DB) Option {
	return func(c *cacheBuild) {
		if pkg.IsNil(c.nutsDB) {
			return
		}
		c.nutsDB = db
	}
}

func WithDefaultNutsDB(cachePath string) Option {
	return func(c *cacheBuild) {
		filepath := "./cache"
		if len(cachePath) > 0 {
			filepath = cachePath
		}
		if !pkg.IsNil(c.nutsDB) {
			return
		}
		db, err := nutsdb.Open(
			nutsdb.DefaultOptions,
			nutsdb.WithDir(filepath), // 数据库会自动创建这个目录文件
		)
		if err != nil {
			panic(err)
		}
		c.nutsDB = db
	}
}
