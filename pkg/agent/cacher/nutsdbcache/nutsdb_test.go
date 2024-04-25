package nutsdbcache

import (
	"testing"

	"github.com/nutsdb/nutsdb"
)

var defaultStoreDir = "./cache"

type cacheVal struct {
	Value string `json:"value"`
}

var val = &cacheVal{
	Value: "default value",
}

func TestNutsDbCache(t *testing.T) {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(defaultStoreDir), // 数据库会自动创建这个目录文件
	)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	cache := NewNutsDbCache(db, "default")
	if err = cache.Set("key", val, 0); err != nil {
		t.Fatal(err)
	}
	var tempVal cacheVal
	if err = cache.Get("key", &tempVal); err != nil {
		t.Fatal(err)
	}
	if tempVal.Value != val.Value {
		t.Fatal("value not equal")
	}

	if cache.SetNX("key", val, 0) {
		t.Fatal(err)
	}
	if !cache.Exists("key") {
		t.Fatal(err)
	}
	if err = cache.Delete("key"); err != nil {
		t.Fatal(err)
	}
	if cache.Exists("key") {
		t.Fatal(err)
	}
}
