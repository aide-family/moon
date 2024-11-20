package main

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtime"
	cache2 "github.com/aide-family/moon/cmd/server/palace/internal/data/runtime/cache"

	"github.com/go-kratos/kratos/v2/log"
	"k8s.io/client-go/tools/cache"
)

func main() {
	scheme := runtime.NewScheme()
	err := scheme.AddKnownTypes(&ExampleStruct1{})
	if err != nil {
		log.Fatalf("add known types failed: %v", err)
	}
	err = scheme.AddKnownTypeWithName("ExampleStruct2", &ExampleStruct2{}, func(obj interface{}) (string, error) {
		return obj.(*ExampleStruct2).Name, nil
	})
	if err != nil {
		log.Fatalf("add known types failed: %v", err)
	}
	e1In := []ExampleStruct1{
		{Name: "1a"},
		{Name: "1b"},
		{Name: "1c"},
	}
	e2In := []ExampleStruct2{
		{Name: "2a"},
		{Name: "2b"},
		{Name: "2c"},
	}
	cacheEntry := cache2.NewCache(scheme)
	ctx := context.Background()
	for i := range e1In {
		err = cacheEntry.Add(ctx, &e1In[i])
		if err != nil {
			log.Fatalf("add failed: %v", err)
		}
	}
	for i := range e2In {
		err = cacheEntry.Add(ctx, &e2In[i])
		if err != nil {
			log.Fatalf("add failed: %v", err)
		}
	}

	// get
	e1out := ExampleStruct1{}
	err = cacheEntry.Get(ctx, "1a", &e1out)
	if err != nil {
		log.Errorf("get failed: %v", err)
	}
	if e1out.Name != "1a" {
		log.Fatalf("get failed: %v", err)
	}

	e2out := ExampleStruct2{}
	err = cacheEntry.Get(ctx, "2a", &e2out)
	if err != nil {
		log.Errorf("get failed: %v", err)
	}
	if e2out.Name != "2a" {
		log.Fatalf("get failed: %v", err)
	}

	// not fund
	err = cacheEntry.Get(ctx, "99", &e1out)
	if err != nil {
		log.Infof("get failed: %v", err)
	} else {
		log.Fatalf("get failed with a not exist key")
	}

	//list
	e1outList := []ExampleStruct1{}
	err = cacheEntry.List(ctx, &e1outList)
	if err != nil {
		log.Errorf("list failed: %v", err)
	}
	if len(e1outList) != 3 {
		log.Fatalf("list failed: %v", err)
	}
	for i, v := range e1outList {
		log.Infof("list: %d, %v", i, v)
	}

	// delete
	err = cacheEntry.Delete(ctx, &e1In[0])
	if err != nil {
		log.Errorf("delete failed: %v", err)
	}
	// should be not fund
	err = cacheEntry.Get(ctx, "1a", &e1out)
	if err != nil {
		log.Infof("delete success %s", err)
	} else {
		log.Fatalf("delete failed")
	}

	informer, err := cacheEntry.GetInformer(ctx, &ExampleStruct1{})
	if err != nil {
		log.Fatalf("delete failed: %v", err)
	}
	err = informer.AddIndexers(cache.Indexers{
		"name": func(obj interface{}) ([]string, error) {
			return []string{obj.(*ExampleStruct1).Name}, nil
		},
	})
	if err != nil {
		log.Fatalf("add by indexer failed: %v", err)
	}
	index, err := informer.GetIndexer().ByIndex("name", "1a")
	if err != nil {
		log.Fatalf("get 1a by index failed: %v", err)
	}
	log.Infof("get 1a by index: %v", index)

	index, err = informer.GetIndexer().ByIndex("name", "1b")
	if err != nil {
		log.Fatalf("get 1b by index failed: %v", err)
	}
	for i := range index {
		log.Infof("get 1b by index: %v", index[i])
	}
}

// ExampleStruct1 示例结构体1
type ExampleStruct1 struct {
	Name string
}

// KeyFunc 示例结构体1的键函数
func (e *ExampleStruct1) KeyFunc(obj interface{}) (string, error) {
	return obj.(*ExampleStruct1).Name, nil
}

// ExampleStruct2 示例结构体2
type ExampleStruct2 struct {
	Name string
}

// KeyFunc 示例结构体2的键函数
func (e *ExampleStruct2) KeyFunc(obj interface{}) (string, error) {
	return obj.(*ExampleStruct2).Name, nil
}
