package cache_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/plugin/cache"
)

var _ cache.Object = (*User)(nil)

type User struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

func (u *User) UniqueKey() string {
	return fmt.Sprintf("%d", u.ID)
}

func (u *User) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

func Test_Cache(t *testing.T) {
	conf := &config.Cache{
		Driver: config.Cache_MEMORY,
	}
	c, err := cache.NewCache(conf)
	if err != nil {
		t.Errorf("new cache error: %v", err)
		return
	}
	defer c.Close()

	users := []*User{
		{
			ID:   1,
			Name: "moon",
		},
		{
			ID:   2,
			Name: "monitor",
		},
	}
	usersMap := make(map[string]any)
	for _, user := range users {
		usersMap[user.UniqueKey()] = user
	}
	ctx := context.Background()
	key := "user"
	if err := c.Client().HSet(ctx, key, usersMap).Err(); err != nil {
		t.Errorf("set cache error: %v", err)
		return
	}

	var user0 User
	if err := c.Client().HGet(ctx, key, "1").Scan(&user0); err != nil {
		t.Errorf("get cache error: %v", err)
		return
	}
	if user0.Name != "moon" {
		t.Errorf("get cache error: %v", err)
		return
	}
	var user1 User
	if err := c.Client().HGet(ctx, key, "2").Scan(&user1); err != nil {
		t.Errorf("get cache error: %v", err)
		return
	}
	if user1.Name != "monitor" {
		t.Errorf("get cache error: %v", err)
		return
	}

	var userAll []*User
	res, err := c.Client().HGetAll(ctx, key).Result()
	if err != nil {
		t.Errorf("get cache error: %v", err)
		return
	}
	for _, v := range res {
		user := new(User)
		if err := user.UnmarshalBinary([]byte(v)); err != nil {
			t.Errorf("get cache error: %v", err)
			return
		}
		userAll = append(userAll, user)
	}
	if len(userAll) != 2 {
		t.Errorf("get cache error: %v", err)
		return
	}
	bs, _ := json.Marshal(userAll)
	t.Logf("userAll: %s", string(bs))
}
