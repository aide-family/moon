package mlog

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aide-family/moon/pkg/util/types"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/go-kratos/kratos/v2/log"
)

// AliyunLogConfig 阿里云日志配置
type AliyunLogConfig interface {
	GetAccessKey() string
	GetAccessSecret() string
	GetEndpoint() string
	GetSecurityToken() string
	GetExpireTime() string
	GetProject() string
	GetStore() string
}

// TODO 阿里云日志插件未测试

// NewAliYunLog new an aliyun logger with options.
func NewAliYunLog(c AliyunLogConfig) Logger {
	if c == nil {
		panic("aliyun log config is nil")
	}

	opts := c

	providerAdapter := sls.NewUpdateFuncProviderAdapter(func() (accessKeyID, accessKeySecret, securityToken string, expireTime time.Time, err error) {
		t, err := time.Parse(time.DateTime, opts.GetExpireTime())
		return opts.GetAccessKey(), opts.GetAccessSecret(), opts.GetSecurityToken(), t, err
	})
	config := &producer.ProducerConfig{
		CredentialsProvider: providerAdapter,
		Endpoint:            opts.GetEndpoint(),
	}
	producerInst := producer.InitProducer(config)

	return &aliyunLog{
		opts:     opts,
		producer: producerInst,
	}
}

type (
	aliyunLog struct {
		producer *producer.Producer
		opts     AliyunLogConfig
	}
)

// Log 日志
func (a *aliyunLog) Log(level log.Level, keyvals ...interface{}) error {
	contents := make([]*sls.LogContent, 0, len(keyvals)/2+1)

	contents = append(contents, &sls.LogContent{
		Key:   types.Of(level.Key()),
		Value: types.Of(level.String()),
	})
	for i := 0; i < len(keyvals); i += 2 {
		contents = append(contents, &sls.LogContent{
			Key:   types.Of(toString(keyvals[i])),
			Value: types.Of(toString(keyvals[i+1])),
		})
	}

	logInst := &sls.Log{
		Time:     types.Of(uint32(time.Now().Unix())),
		Contents: contents,
	}
	return a.producer.SendLog(a.opts.GetProject(), a.opts.GetStore(), "", "", logInst)
}

// toString convert any type to string
func toString(v interface{}) string {
	var key string
	if v == nil {
		return key
	}
	switch v := v.(type) {
	case float64:
		key = strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		key = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		key = strconv.Itoa(v)
	case uint:
		key = strconv.FormatUint(uint64(v), 10)
	case int8:
		key = strconv.Itoa(int(v))
	case uint8:
		key = strconv.FormatUint(uint64(v), 10)
	case int16:
		key = strconv.Itoa(int(v))
	case uint16:
		key = strconv.FormatUint(uint64(v), 10)
	case int32:
		key = strconv.Itoa(int(v))
	case uint32:
		key = strconv.FormatUint(uint64(v), 10)
	case int64:
		key = strconv.FormatInt(v, 10)
	case uint64:
		key = strconv.FormatUint(v, 10)
	case string:
		key = v
	case bool:
		key = strconv.FormatBool(v)
	case []byte:
		key = string(v)
	case fmt.Stringer:
		key = v.String()
	default:
		newValue, _ := types.Marshal(v)
		key = string(newValue)
	}
	return key
}
