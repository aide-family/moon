package cache

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/aide-family/moon/pkg/runtime"
	"github.com/aide-family/moon/pkg/runtime/client"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/client-go/tools/cache"
)

type Options struct {
	Scheme *runtime.Scheme

	Resync *time.Duration
}

var _ client.Reader = &Reader{}

type Reader struct {
	indexer cache.Indexer
	kind    string
}

func (c *Reader) Get(_ context.Context, key string, out client.Object) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("object %s not found, kind: %s", key, c.kind)
	}

	if _, isObj := obj.(runtime.Object); !isObj {
		return fmt.Errorf("cache contained %T, which is not an Object", obj)
	}
	// 进行深度复制，以避免改变缓存
	obj = obj.(runtime.Object).DeepCopyObject()

	outVal := reflect.ValueOf(out)
	objVal := reflect.ValueOf(obj)
	// 判断输出的对象类型是否相同
	if !objVal.Type().AssignableTo(outVal.Type()) {
		return fmt.Errorf("cache had type %s, but %s was asked for", objVal.Type(), outVal.Type())
	}
	reflect.Indirect(outVal).Set(reflect.Indirect(objVal))
	out.GetObjectKind().SetKind(c.kind)
	return nil
}

func (c *Reader) List(_ context.Context, out []client.Object) error {
	objs := c.indexer.List()

	v, err := conversion.EnforcePtr(out)
	if err != nil {
		return err
	}

	runtimeObjs := make([]runtime.Object, 0, len(objs))
	for _, item := range objs {
		obj, isObj := item.(runtime.Object)
		if !isObj {
			return fmt.Errorf("cache contained %T, which is not an Object", obj)
		}
		outObj := obj.DeepCopyObject()
		outObj.GetObjectKind().SetKind(c.kind)
		runtimeObjs = append(runtimeObjs, outObj)
	}

	v.Elem().Set(reflect.ValueOf(runtimeObjs))
	return nil
}
