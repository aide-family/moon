package nutsdbcacher

import (
	"context"
	"errors"
	"time"

	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/nutsdb/nutsdb"
)

type NutsDbConfig interface {
	GetPath() string
	GetBucket() string
}

func NewNutsDbCacher(cfg NutsDbConfig) (conn.Cache, error) {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(cfg.GetPath()), // 数据库会自动创建这个目录文件
	)
	if err != nil {
		return nil, err
	}
	return &nutsDbCacher{
		cli:    db,
		bucket: cfg.GetBucket(),
		path:   cfg.GetPath(),
	}, nil
}

type nutsDbCacher struct {
	cli    *nutsdb.DB
	bucket string
	path   string
}

func (l *nutsDbCacher) Get(ctx context.Context, key string) (string, error) {
	if err := l.newBucket(l.bucket); err != nil {
		return "", err
	}
	var bytes []byte
	err := l.cli.View(func(tx *nutsdb.Tx) error {
		mGet, err := tx.Get(l.bucket, []byte(key))
		if err != nil {
			return err
		}
		bytes = mGet
		return nil
	})
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (l *nutsDbCacher) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if err := l.newBucket(l.bucket); err != nil {
		return err
	}

	return l.cli.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(l.bucket, []byte(key), []byte(value), uint32(expiration.Seconds()))
	})
}

func (l *nutsDbCacher) Delete(ctx context.Context, key string) error {
	if err := l.newBucket(l.bucket); err != nil {
		return err
	}
	return l.cli.Update(func(tx *nutsdb.Tx) error {
		if err := tx.Delete(l.bucket, []byte(key)); err != nil {
			if errors.Is(err, nutsdb.ErrKeyNotFound) {
				return nil
			}
			return err
		}
		return nil
	})
}

func (l *nutsDbCacher) Close() error {
	return l.cli.Close()
}

func (l *nutsDbCacher) Exist(ctx context.Context, key string) bool {
	res, err := l.Get(ctx, key)
	if err != nil {
		return false
	}
	return res != ""
}

func (l *nutsDbCacher) newBucket(prefix string) error {
	bucket := prefix
	tx, err := l.cli.Begin(true)
	if err != nil {
		return err
	}
	if !tx.ExistBucket(nutsdb.DataStructureBTree, bucket) {
		if err = tx.NewBucket(nutsdb.DataStructureBTree, bucket); err != nil {
			return err
		}
	}
	return tx.Commit()
}
