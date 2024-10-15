package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/coocood/freecache"
)

func Test_defaultCache_Exist(t *testing.T) {
	ctx := context.Background()
	size := 0
	cli := cache.NewFreeCache(freecache.NewCache(types.Ternary(size > 0, size, 10*1024*1024)))
	defer cli.Close()
	key := fmt.Sprintf("oauth:%d:%s", 1, types.MD5("xx"))
	t.Log(cli.Exist(ctx, key))
	t.Log(cli.Set(ctx, key, "xxx", 0))
	time.Sleep(5 * time.Second)
	t.Log(cli.Exist(ctx, key))
	t.Log(cli.Get(ctx, key))
	t.Log(cli.Delete(ctx, key))
	t.Log(cli.Exist(ctx, key))
}
