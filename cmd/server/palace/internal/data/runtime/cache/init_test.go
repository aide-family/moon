package cache

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtime"
	"k8s.io/client-go/tools/cache"
)

type TestStruct struct {
	Field1 string
	Field2 int
}

func (t *TestStruct) KeyFunc(obj interface{}) (string, error) {
	return obj.(*TestStruct).Field1, nil
}

var ts = &TestStruct{}

func NewMockIndexer() cache.Indexer {
	return cache.NewIndexer(ts.KeyFunc, cache.Indexers{})
}

var testScheme = runtime.NewScheme()

func init() {
	testScheme.AddKnownTypes(&TestStruct{})
}
