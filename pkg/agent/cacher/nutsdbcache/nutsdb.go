package nutsdbcache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aide-family/moon/pkg/agent"
	"github.com/nutsdb/nutsdb"
)

// NewNutsDbCache creates a new nutsdb cache.
func NewNutsDbCache(db *nutsdb.DB, bucket string) agent.Cache {
	return &nutsDbCache{
		db:     db,
		bucket: bucket,
	}
}

type nutsDbCache struct {
	db     *nutsdb.DB
	bucket string
	ctx    context.Context
}

func (n *nutsDbCache) Close() error {
	return n.db.Close()
}

func (n *nutsDbCache) WithContext(ctx context.Context) agent.Cache {
	n.ctx = context.WithoutCancel(ctx)
	return n
}

func (n *nutsDbCache) SetNX(key string, value any, expiration time.Duration) bool {
	if err := n.newBucket(n.bucket); err != nil {
		return false
	}
	return !n.Exists(key) && n.Set(key, value, expiration) == nil
}

func (n *nutsDbCache) Exists(key string) bool {
	if err := n.newBucket(n.bucket); err != nil {
		return false
	}
	err := n.db.View(func(tx *nutsdb.Tx) error {
		_, err := tx.Get(n.bucket, []byte(key))
		return err
	})

	return err == nil
}

func (n *nutsDbCache) Get(key string, value any) error {
	if err := n.newBucket(n.bucket); err != nil {
		return err
	}
	var bytes []byte
	err := n.db.View(func(tx *nutsdb.Tx) error {
		mGet, err := tx.Get(n.bucket, []byte(key))
		if err != nil {
			return err
		}
		bytes = mGet
		return nil
	})
	if err != nil {
		if errors.Is(err, nutsdb.ErrKeyNotFound) {
			return agent.NoCache
		}
		return err
	}
	return json.Unmarshal(bytes, value)
}

func (n *nutsDbCache) Set(key string, value any, expiration time.Duration) error {
	if err := n.newBucket(n.bucket); err != nil {
		return err
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return n.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(n.bucket, []byte(key), bytes, uint32(expiration.Seconds()))
	})
}

func (n *nutsDbCache) Delete(key string) error {
	if err := n.newBucket(n.bucket); err != nil {
		return err
	}
	return n.db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.Delete(n.bucket, []byte(key)); err != nil {
			if errors.Is(err, nutsdb.ErrKeyNotFound) {
				return nil
			}
			return err
		}
		return nil
	})
}

func (n *nutsDbCache) newBucket(prefix string) error {
	bucket := prefix
	tx, err := n.db.Begin(true)
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
