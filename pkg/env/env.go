package env

import (
	"os"
	"sync"

	"github.com/aide-cloud/moon/pkg/types"
)

var (
	name     string
	version  string
	metadata map[string]string
	id, _    = os.Hostname()

	nameOnce, versionOnce, metadataOnce sync.Once
)

func init() {
	name = "moon"
	version = "0.0.1"
	metadata = make(map[string]string)
}

// Name 获取服务名称
func Name() string {
	return name
}

// Version 获取服务版本
func Version() string {
	return version
}

// ID 获取服务ID
func ID() string {
	return id
}

func Metadata() map[string]string {
	return metadata
}

// SetName 设置服务名称
func SetName(n string) {
	if n == "" {
		return
	}
	nameOnce.Do(func() {
		name = n
	})
}

// SetVersion 设置服务版本
func SetVersion(v string) {
	if v == "" {
		return
	}
	versionOnce.Do(func() {
		version = v
	})
}

func SetMetadata(m map[string]string) {
	if types.IsNil(m) || len(m) == 0 {
		return
	}
	metadataOnce.Do(func() {
		metadata = m
	})
}
