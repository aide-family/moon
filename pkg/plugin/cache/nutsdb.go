package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/nutsdb/nutsdb"
)

var _ ICacher = (*nutsDBCacher)(nil)

// NewNutsDbCacher creates a new nutsdb cache client.
func NewNutsDbCacher(client *nutsdb.DB, bucket string) ICacher {
	return &nutsDBCacher{
		client: client,
		bucket: bucket,
	}
}

type (
	nutsDBCacher struct {
		client *nutsdb.DB
		bucket string
	}
)

func (n *nutsDBCacher) Close() error {
	return n.client.Close()
}

func (n *nutsDBCacher) Delete(ctx context.Context, key string) error {
	return n.client.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(n.bucket, []byte(key))
	})
}

func (n *nutsDBCacher) Exist(ctx context.Context, key string) (bool, error) {
	_, err := n.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (n *nutsDBCacher) Get(_ context.Context, key string) (string, error) {
	var val []byte
	err := n.client.View(func(tx *nutsdb.Tx) error {
		data, err := tx.Get(n.bucket, []byte(key))
		if err != nil {
			return err
		}
		val = data
		return nil
	})
	return string(val), err
}

func (n *nutsDBCacher) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return n.client.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(n.bucket, []byte(key), []byte(value), uint32(expiration.Seconds()))
	})
}

func (n *nutsDBCacher) Inc(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	getInt64, _ := n.GetInt64(ctx, key)
	val := getInt64 + 1
	if err := n.SetInt64(ctx, key, val, expiration); err != nil {
		return 0, err
	}
	return val, nil
}

func (n *nutsDBCacher) Dec(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	getInt64, _ := n.GetInt64(ctx, key)
	val := getInt64 - 1
	if err := n.SetInt64(ctx, key, val, expiration); err != nil {
		return 0, err
	}
	return val, nil
}

func (n *nutsDBCacher) IncMax(ctx context.Context, key string, max int64, expiration time.Duration) (bool, error) {
	getInt64, _ := n.GetInt64(ctx, key)
	if getInt64 >= max {
		return false, nil
	}
	val := getInt64 + 1
	if err := n.SetInt64(ctx, key, val, expiration); err != nil {
		return false, err
	}
	return val <= max, nil
}

func (n *nutsDBCacher) DecMin(ctx context.Context, key string, min int64, expiration time.Duration) (bool, error) {
	getInt64, _ := n.GetInt64(ctx, key)
	if getInt64 <= min {
		return false, nil
	}
	val := getInt64 - 1
	if err := n.SetInt64(ctx, key, val, expiration); err != nil {
		return false, err
	}
	return val >= min, nil
}

func (n *nutsDBCacher) GetInt64(ctx context.Context, key string) (int64, error) {
	get, err := n.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(get, 10, 64)
}

func (n *nutsDBCacher) SetInt64(ctx context.Context, key string, value int64, expiration time.Duration) error {
	return n.Set(ctx, key, strconv.FormatInt(value, 10), expiration)
}

func (n *nutsDBCacher) GetFloat64(ctx context.Context, key string) (float64, error) {
	get, err := n.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(get, 64)
}

func (n *nutsDBCacher) SetFloat64(ctx context.Context, key string, value float64, expiration time.Duration) error {
	return n.Set(ctx, key, strconv.FormatFloat(value, 'f', -1, 64), expiration)
}

func (n *nutsDBCacher) GetObject(ctx context.Context, key string, obj IObjectSchema) error {
	bs, err := n.Get(ctx, key)
	if err != nil {
		return err
	}
	return obj.UnmarshalBinary([]byte(bs))
}

func (n *nutsDBCacher) SetObject(ctx context.Context, key string, obj IObjectSchema, expiration time.Duration) error {
	bs, err := obj.MarshalBinary()
	if err != nil {
		return err
	}
	return n.Set(ctx, key, string(bs), expiration)
}

func (n *nutsDBCacher) GetBool(ctx context.Context, key string) (bool, error) {
	get, err := n.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(get)
}

func (n *nutsDBCacher) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	return n.Set(ctx, key, strconv.FormatBool(value), expiration)
}

func (n *nutsDBCacher) SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	exist, err := n.Exist(ctx, key)
	if err != nil {
		return false, err
	}
	if exist {
		return false, nil
	}
	return true, n.Set(ctx, key, value, expiration)
}
