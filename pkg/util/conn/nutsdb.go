package conn

import (
	"github.com/nutsdb/nutsdb"
)

// NutsDbConfig nuts db 配置
type NutsDbConfig interface {
	GetPath() string
	GetBucket() string
}

// NewNutsDB 创建一个 nuts db
func NewNutsDB(cfg NutsDbConfig) (*nutsdb.DB, error) {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(cfg.GetPath()), // 数据库会自动创建这个目录文件
	)
	if err != nil {
		return nil, err
	}
	bucket := cfg.GetBucket()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Commit()
	}()
	if !tx.ExistBucket(nutsdb.DataStructureBTree, bucket) {
		if err = tx.NewBucket(nutsdb.DataStructureBTree, bucket); err != nil {
			return nil, err
		}
	}
	return db, nil
}
