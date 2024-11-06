package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/coocood/freecache"
	"github.com/go-kratos/kratos/v2/errors"
)

var _ ICacher = (*defaultCache)(nil)

// NewFreeCache 创建一个默认的缓存
func NewFreeCache(cli *freecache.Cache) ICacher {
	return &defaultCache{
		cli:  cli,
		keys: make(map[string]struct{}, 1024),
	}
}

type (
	defaultCache struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
)

func (d *defaultCache) Keys(_ context.Context, prefix string) ([]string, error) {
	keys := make([]string, 0, 1024)
	for k := range d.keys {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (d *defaultCache) DelKeys(ctx context.Context, prefix string) error {
	keys, err := d.Keys(ctx, prefix)
	if err != nil {
		return err
	}
	for _, k := range keys {
		if err = d.Delete(ctx, k); err != nil {
			return err
		}
	}
	return nil
}

func (d *defaultCache) GetInt64(ctx context.Context, key string) (int64, error) {
	return d.getInt64(ctx, key)
}

func (d *defaultCache) SetInt64(ctx context.Context, key string, value int64, expiration time.Duration) error {
	return d.cli.Set([]byte(key), []byte(strconv.FormatInt(value, 10)), int(expiration.Seconds()))
}

func (d *defaultCache) GetFloat64(ctx context.Context, key string) (float64, error) {
	return d.getFloat64(ctx, key)
}

func (d *defaultCache) SetFloat64(_ context.Context, key string, value float64, expiration time.Duration) error {
	return d.cli.Set([]byte(key), []byte(strconv.FormatFloat(value, 'f', -1, 64)), int(expiration.Seconds()))
}

func (d *defaultCache) GetObject(_ context.Context, key string, obj IObjectSchema) error {
	data, err := d.cli.Get([]byte(key))
	if err != nil && !errors.Is(err, freecache.ErrNotFound) {
		return err
	}
	return obj.UnmarshalBinary(data)
}

func (d *defaultCache) SetObject(_ context.Context, key string, obj IObjectSchema, expiration time.Duration) error {
	data, err := obj.MarshalBinary()
	if err != nil {
		return err
	}
	return d.cli.Set([]byte(key), data, int(expiration.Seconds()))
}

func (d *defaultCache) GetBool(_ context.Context, key string) (bool, error) {
	data, err := d.cli.Get([]byte(key))
	if err != nil && !errors.Is(err, freecache.ErrNotFound) {
		return false, err
	}

	return string(data) == "true", nil
}

func (d *defaultCache) SetBool(_ context.Context, key string, value bool, expiration time.Duration) error {
	return d.cli.Set([]byte(key), []byte(strconv.FormatBool(value)), int(expiration.Seconds()))
}

func (d *defaultCache) addNum(ctx context.Context, key string, num int64, expiration time.Duration) (int64, error) {
	parseInt, err := d.getInt64(ctx, key)
	if err != nil && !errors.Is(err, freecache.ErrNotFound) {
		return 0, err
	}

	val := parseInt + num
	if err = d.cli.Set([]byte(key), []byte(strconv.FormatInt(val, 10)), int(expiration.Seconds())); err != nil {
		return 0, err
	}
	return val, nil
}

func (d *defaultCache) getInt64(_ context.Context, key string) (int64, error) {
	dataBytes, err := d.cli.Get([]byte(key))
	if err != nil {
		return 0, err
	}
	var parseInt int64
	if len(dataBytes) > 0 {
		parseInt, err = strconv.ParseInt(string(dataBytes), 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return parseInt, nil
}

func (d *defaultCache) getFloat64(_ context.Context, key string) (float64, error) {
	dataBytes, err := d.cli.Get([]byte(key))
	if err != nil {
		return 0, err
	}
	var parseFloat float64
	if len(dataBytes) > 0 {
		parseFloat, err = strconv.ParseFloat(string(dataBytes), 10)
		if err != nil {
			return 0, err
		}
	}
	return parseFloat, nil
}

func (d *defaultCache) getString(_ context.Context, key string) (string, error) {
	dataBytes, err := d.cli.Get([]byte(key))
	if err != nil {
		return "", err
	}
	return string(dataBytes), nil
}

func (d *defaultCache) Inc(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return d.addNum(ctx, key, 1, expiration)
}

func (d *defaultCache) Dec(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return d.addNum(ctx, key, -1, expiration)
}

func (d *defaultCache) IncMax(ctx context.Context, key string, max int64, expiration time.Duration) (bool, error) {
	num, err := d.getInt64(ctx, key)
	if err != nil && !errors.Is(err, freecache.ErrNotFound) {
		return false, err
	}
	if num >= max {
		return false, nil
	}
	_, err = d.Inc(ctx, key, expiration)
	return true, err
}

func (d *defaultCache) DecMin(ctx context.Context, key string, min int64, expiration time.Duration) (bool, error) {
	num, err := d.getInt64(ctx, key)
	if err != nil && !errors.Is(err, freecache.ErrNotFound) {
		return false, err
	}
	if num <= min {
		return false, nil
	}
	_, err = d.Dec(ctx, key, expiration)
	return true, err
}

func (d *defaultCache) Close() error {
	if d.cli == nil {
		return nil
	}
	d.cli.Clear()
	d.cli = nil
	return nil
}

func (d *defaultCache) Delete(_ context.Context, key string) error {
	d.cli.Del([]byte(key))
	return nil
}

func (d *defaultCache) Exist(ctx context.Context, key string) (bool, error) {
	_, err := d.getString(ctx, key)
	if err != nil {
		if errors.Is(err, freecache.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *defaultCache) Get(ctx context.Context, key string) (string, error) {
	return d.getString(ctx, key)
}

func (d *defaultCache) Set(_ context.Context, key string, value string, expiration time.Duration) error {
	return d.cli.Set([]byte(key), []byte(value), int(expiration.Seconds()))
}

func (d *defaultCache) SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	return d.IncMax(ctx, key, 1, expiration)
}
