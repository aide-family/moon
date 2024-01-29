package cache

import (
	"context"
	"testing"
	"time"
)

func Test_nutsDbCache_HDel(t *testing.T) {
	c, err := NewNutsDbCache()
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()
	if err = c.HSet(context.Background(), "test", []byte("1"), []byte("1")); err != nil {
		t.Error(err)
		return
	}
	all, err := c.HGetAll(context.Background(), "test")
	if err != nil {
		t.Error(err)
		return
	}
	for key, value := range all {
		t.Log(key, string(value))
	}
	if err = c.HDel(context.Background(), "test", "1"); err != nil {
		t.Error(err)
		return
	}
	all, err = c.HGetAll(context.Background(), "test")
	if err != nil {
		t.Error(err)
		return
	}
	for key, value := range all {
		t.Log(key, string(value))
	}
}

func TestSetAdnGet(t *testing.T) {
	c, err := NewNutsDbCache()
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()
	if err = c.Set(context.Background(), "test", []byte("1"), 0); err != nil {
		t.Error(err)
		return
	}
	if value, err := c.Get(context.Background(), "test"); err != nil {
		t.Error(err)
	} else {
		t.Log(string(value))
	}
}

func TestSetNx(t *testing.T) {
	c, err := NewNutsDbCache()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(c.SetNX(context.Background(), "test", []byte("1"), 10*time.Second))
	time.Sleep(1 * time.Second)
	t.Log(c.SetNX(context.Background(), "test", []byte("1"), 10*time.Second))
	time.Sleep(9 * time.Second)
	t.Log(c.SetNX(context.Background(), "test", []byte("1"), 10*time.Second))
}
