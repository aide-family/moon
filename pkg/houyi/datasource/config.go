package datasource

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/util/types"
)

var (
	_ sql.Scanner   = (*Config)(nil)
	_ driver.Valuer = (*Config)(nil)
)

// Config 数据源配置
type Config struct {
	datasourceConfig map[string]any
	conf             string
}

type (
	// MetricConfig 指标数据源配置
	MetricConfig struct {
		ClientCert string                `json:"clientCert"`
		ClientKey  string                `json:"clientKey"`
		Headers    []*MetricConfigHeader `json:"headers"`
		Params     []*MetricConfigParams `json:"params"`
		Password   string                `json:"password"`
		SelfCACert string                `json:"selfCACert"`
		ServerName string                `json:"serverName"`
		SkipVerify bool                  `json:"skipVerify"`
		Username   string                `json:"username"`
	}

	// MetricConfigHeader 指标数据源配置-请求头参数
	MetricConfigHeader struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	// MetricConfigParams 指标数据源配置-请求参数
	MetricConfigParams struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
)

// Scan 实现 sql.Scanner 接口
func (d *Config) Scan(src any) (err error) {
	switch s := src.(type) {
	case []byte:
		d.conf = string(s)
		err = types.Unmarshal(s, &d.datasourceConfig)
	case string:
		d.conf = s
		err = types.Unmarshal([]byte(s), &d.datasourceConfig)
	default:
		err = label.ErrUnsupportedType
	}
	return err
}

// Value 实现 driver.Valuer 接口
func (d *Config) Value() (driver.Value, error) {
	return d.String(), nil
}

// NewDatasourceConfig 基于map创建DatasourceConfig
func NewDatasourceConfig(datasourceConfig map[string]any) *Config {
	return &Config{datasourceConfig: datasourceConfig}
}

// NewDatasourceConfigByString 基于string创建DatasourceConfig
func NewDatasourceConfigByString(datasourceConfig string) *Config {
	m := make(map[string]any)
	_ = json.Unmarshal([]byte(datasourceConfig), &m)
	return &Config{datasourceConfig: m}
}

// MarshalJSON 实现 json.Marshaler 接口
func (d *Config) MarshalJSON() ([]byte, error) {
	// 返回字符串形式的时间
	return []byte(d.String()), nil
}

// String 转json字符串
func (d *Config) String() string {
	if types.IsNil(d) || len(d.datasourceConfig) == 0 {
		return "{}"
	}

	bs, _ := types.Marshal(d.datasourceConfig)
	return string(bs)
}

// Map 转map
func (d *Config) Map() map[string]any {
	if d == nil || d.datasourceConfig == nil {
		return make(map[string]any)
	}
	return d.datasourceConfig
}

// GetRocketMQ 获取RocketMQ配置
func (d *Config) GetRocketMQ() *conf.RocketMQ {
	config := &conf.RocketMQ{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.conf), config)
	return config
}

// GetMQTT 获取MQTT配置
func (d *Config) GetMQTT() *conf.MQTT {
	config := &conf.MQTT{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.conf), config)
	return config
}

// GetKafka 获取Kafka配置
func (d *Config) GetKafka() *conf.Kafka {
	config := &conf.Kafka{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.conf), config)
	return config
}

// GetMetric 获取指标数据源配置
func (d *Config) GetMetric() *MetricConfig {
	config := &MetricConfig{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.conf), config)
	return config
}
