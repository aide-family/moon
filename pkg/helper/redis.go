package helper

import (
	"strconv"
	"strings"
)

type RedisKey string

const (
	// PromGroupDeleteKey 删除规则, 用于记录删除的ID列表数据
	PromGroupDeleteKey RedisKey = "prom:group:delete"
	// PromGroupChangeKey 更新规则, 用于记录更新的ID列表数据
	PromGroupChangeKey RedisKey = "prom:group:change"
)

func (r RedisKey) String() string {
	return string(r)
}

func (r RedisKey) Key(args ...string) RedisKey {
	return RedisKey(strings.Join(append([]string{r.String()}, args...), ":"))
}

func (r RedisKey) KeyInt(args ...int32) RedisKey {
	var s []string
	for _, v := range args {
		s = append(s, strconv.FormatInt(int64(v), 10))
	}
	return r.Key(s...)
}
