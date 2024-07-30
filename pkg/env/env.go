package env

import (
	"os"
	"sync"

	"github.com/aide-family/moon/pkg/util/types"
)

var (
	name     string
	version  string
	metadata map[string]string
	id, _    = os.Hostname()
	env      string

	nameOnce, versionOnce, metadataOnce, envOnce sync.Once
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

// Metadata 获取服务元数据
func Metadata() map[string]string {
	return metadata
}

// Env 获取服务环境
func Env() string {
	return env
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

// SetMetadata 设置服务元数据
func SetMetadata(m map[string]string) {
	if types.IsNil(m) || len(m) == 0 {
		return
	}
	metadataOnce.Do(func() {
		metadata = m
	})
}

// SetEnv 设置服务环境变量
func SetEnv(e string) {
	if e == "" {
		return
	}
	envOnce.Do(func() {
		env = e
	})
}

// Type 运行环境类型
type Type string

const (
	// Local 本地环境
	Local Type = "local"
	// Dev  开发环境
	Dev Type = "dev"
	// Test 测试环境
	Test Type = "test"
	// Prod 生产环境
	Prod Type = "prod"
)

// IsLocal 是否是本地环境
func IsLocal() bool {
	return env == string(Local)
}

// IsDev 是否是开发环境
func IsDev() bool {
	return env == string(Dev)
}

// IsTest 是否是测试环境
func IsTest() bool {
	return env == string(Test)
}

// IsProd 是否是生产环境
func IsProd() bool {
	return env == string(Prod)
}

// IsEnv 是否是指定环境
func IsEnv(e string) bool {
	return env == e
}
