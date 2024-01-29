package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/nutsdb/nutsdb"
)

var _ GlobalCache = (*nutsDbCache)(nil)

type GlobalCache interface {
	HDel(ctx context.Context, prefix string, keys ...string) error
	HSet(ctx context.Context, prefix string, values ...[]byte) error
	HGet(ctx context.Context, prefix string, keys string) ([]byte, error)
	HGetAll(ctx context.Context, prefix string) (map[string][]byte, error)
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) bool
	Exists(ctx context.Context, keys ...string) int64
	Close() error
}

const defaultStoreDir = "./cache"
const defaultBucket = "default"

type (
	nutsDbCache struct {
		db     *nutsdb.DB
		bucket string
	}

	NutsDbOption func(*nutsDbCache)
)

func (l *nutsDbCache) Exists(_ context.Context, keys ...string) int64 {
	if err := l.newBucket(l.bucket); err != nil {
		return 0
	}
	kBytes := make([][]byte, 0, len(keys))
	for _, key := range keys {
		kBytes = append(kBytes, []byte(key))
	}
	var count int64
	if err := l.db.View(func(tx *nutsdb.Tx) error {
		mGet, err := tx.MGet(l.bucket, kBytes...)
		if err != nil {
			return err
		}
		count = int64(len(mGet))
		return nil
	}); err != nil {
		return 0
	}
	return count
}

func (l *nutsDbCache) SetNX(_ context.Context, key string, value []byte, ttl time.Duration) bool {
	if err := l.newBucket(l.bucket); err != nil {
		return false
	}
	// 1. 获取值是否存在
	if err := l.db.Update(func(tx *nutsdb.Tx) error {
		if _, err := tx.Get(l.bucket, []byte(key)); err != nil {
			if errors.Is(err, nutsdb.ErrKeyNotFound) {
				return tx.Put(l.bucket, []byte(key), value, uint32(ttl.Seconds()))
			}
			return err
		}
		return errors.New("key already exists")
	}); err != nil {
		return false
	}
	return true
}

func (l *nutsDbCache) HGet(_ context.Context, prefix string, keys string) ([]byte, error) {
	if err := l.newBucket(l.bucket); err != nil {
		return nil, err
	}
	var res []byte
	if err := l.db.View(func(tx *nutsdb.Tx) error {
		if v, err := tx.MGet(l.bucket, []byte(prefix+keys)); err != nil {
			if errors.Is(err, nutsdb.ErrKeyNotFound) {
				return nil
			}
			return err
		} else {
			if len(v) == 0 {
				return nutsdb.ErrKeyNotFound
			}
			res = v[0]
			return nil
		}
	}); err != nil {
		return nil, err
	}
	return res, nil
}

func (l *nutsDbCache) Del(_ context.Context, keys ...string) error {
	if err := l.newBucket(l.bucket); err != nil {
		return err
	}
	return l.db.Update(func(tx *nutsdb.Tx) error {
		for _, key := range keys {
			if err := tx.Delete(l.bucket, []byte(key)); err != nil {
				if errors.Is(err, nutsdb.ErrKeyNotFound) {
					return nil
				}
				return err
			}
		}
		return nil
	})
}

func (l *nutsDbCache) Close() error {
	return l.db.Close()
}

func (l *nutsDbCache) HDel(_ context.Context, prefix string, keys ...string) error {
	if err := l.newBucket(l.bucket); err != nil {
		return err
	}
	return l.db.Update(func(tx *nutsdb.Tx) error {
		if len(keys) == 0 {
			return nil
		}
		for _, key := range keys {
			if err := tx.Delete(l.bucket, []byte(prefix+key)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (l *nutsDbCache) HSet(_ context.Context, prefix string, values ...[]byte) error {
	if err := l.newBucket(l.bucket); err != nil {
		return err
	}

	args := make([][]byte, 0, len(values))
	for index, value := range values {
		if index%2 == 1 {
			args = append(args, value)
		} else {
			key := prefix + string(value)
			args = append(args, []byte(key))
		}
	}
	return l.db.Update(func(tx *nutsdb.Tx) error {
		return tx.MSet(l.bucket, nutsdb.Persistent, args...)
	})
}

func (l *nutsDbCache) HGetAll(_ context.Context, prefix string) (map[string][]byte, error) {
	if err := l.newBucket(l.bucket); err != nil {
		return nil, err
	}
	res := make(map[string][]byte)
	err := l.db.View(func(tx *nutsdb.Tx) error {
		keys, values, err := tx.GetAll(l.bucket)
		if err != nil {
			return err
		}
		if len(values) == 0 || len(keys) == 0 || len(keys) != len(values) {
			return nutsdb.ErrKeyNotFound
		}
		for i := 0; i < len(keys); i++ {
			key := string(keys[i])
			// 判断是否是前缀
			if !strings.HasPrefix(key, prefix) {
				continue
			}
			key = strings.TrimPrefix(key, prefix)
			res[key] = values[i]
		}
		return nil
	})
	return res, err
}

func (l *nutsDbCache) Get(_ context.Context, key string) ([]byte, error) {
	if err := l.newBucket(l.bucket); err != nil {
		return nil, err
	}
	var value []byte
	err := l.db.View(func(tx *nutsdb.Tx) error {
		mGet, err := tx.Get(l.bucket, []byte(key))
		if err != nil {
			return err
		}
		value = mGet
		return nil
	})
	return value, err
}

func (l *nutsDbCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	if err := l.newBucket(l.bucket); err != nil {
		return err
	}
	return l.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(l.bucket, []byte(key), value, uint32(ttl.Seconds()))
	})
}

func (l *nutsDbCache) newBucket(prefix string) error {
	bucket := prefix
	tx, err := l.db.Begin(true)
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

func NewNutsDbCache(opts ...NutsDbOption) (GlobalCache, error) {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(defaultStoreDir), // 数据库会自动创建这个目录文件
	)
	if err != nil {
		return nil, err
	}
	cache := nutsDbCache{db: db, bucket: defaultBucket}
	for _, opt := range opts {
		opt(&cache)
	}

	return &cache, nil
}

func WithDB(db *nutsdb.DB) NutsDbOption {
	return func(cache *nutsDbCache) {
		cache.db = db
	}
}

func WithBucket(bucket string) NutsDbOption {
	return func(cache *nutsDbCache) {
		cache.bucket = bucket
	}
}
