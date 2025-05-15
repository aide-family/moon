package kv_test

import (
	"testing"

	"github.com/aide-family/moon/pkg/util/kv"
)

func Test_New(t *testing.T) {
	m := kv.New(map[string]string{"a": "a"})
	a, ok := m.Get("a")
	if !ok {
		t.Fatal("test failed")
	}
	if a != "a" {
		t.Fatal("test failed")
	}
	m.Set("b", "b")
	b, ok := m.Get("b")
	if !ok {
		t.Fatal("test failed")
	}
	if b != "b" {
		t.Fatal("test failed")
	}
	m.Del("b")
	_, ok = m.Get("b")
	if ok {
		t.Fatal("test failed")
	}
	marshalBinary, err := m.MarshalBinary()
	if err != nil {
		t.Fatal("test failed")
	}
	if len(marshalBinary) == 0 {
		t.Fatal("test failed")
	}
	kvStr := string(marshalBinary)
	if kvStr != `{"a":"a"}` {
		t.Fatal("test failed")
	}
	if err := m.UnmarshalBinary([]byte(`{"a":"a1","b":"b1"}`)); err != nil {
		t.Fatal("test failed")
	}
	if m.Len() != 2 {
		t.Fatal("test failed")
	}
	a1, ok := m.Get("a")
	if !ok {
		t.Fatal("test failed")
	}
	if a1 != "a1" {
		t.Fatal("test failed")
	}
	b1, ok := m.Get("b")
	if !ok {
		t.Fatal("test failed")
	}
	if b1 != "b1" {
		t.Fatal("test failed")
	}
	t.Log(m)
}
