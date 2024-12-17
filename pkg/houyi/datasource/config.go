package datasource

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"sort"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"golang.org/x/exp/maps"
)

var _ sql.Scanner = (*Config)(nil)
var _ driver.Valuer = (*Config)(nil)

// Config 数据源配置
type Config struct {
	datasourceConfig map[string]string
}

// Scan 实现 sql.Scanner 接口
func (d *Config) Scan(src any) (err error) {
	switch s := src.(type) {
	case []byte:
		err = types.Unmarshal(s, &d.datasourceConfig)
	case string:
		err = types.Unmarshal([]byte(s), &d.datasourceConfig)
	default:
		err = vobj.ErrUnsupportedType
	}
	return err
}

// Value 实现 driver.Valuer 接口
func (d *Config) Value() (driver.Value, error) {
	return d.String(), nil
}

// NewDatasourceConfig 基于map创建DatasourceConfig
func NewDatasourceConfig(datasourceConfig map[string]string) *Config {
	return &Config{datasourceConfig: datasourceConfig}
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

	confKeys := maps.Keys(d.datasourceConfig)
	sort.Strings(confKeys)
	list := make([]string, 0, len(confKeys)*5)
	list = append(list, "{")
	for _, k := range confKeys {
		list = append(list, `"`, k, `":"`, d.datasourceConfig[k], `"`, ",")
	}
	list = append(list[:len(list)-1], "}")
	return types.TextJoin(list...)
}

// Map 转map
func (d *Config) Map() map[string]string {
	if d == nil || d.datasourceConfig == nil {
		return make(map[string]string)
	}
	return d.datasourceConfig
}

// GetRocketMQ 获取RocketMQ配置
func (d *Config) GetRocketMQ() *conf.RocketMQ {
	config := &conf.RocketMQ{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.String()), config)
	return config
}

// GetMQTT 获取MQTT配置
func (d *Config) GetMQTT() *conf.MQTT {
	config := &conf.MQTT{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.String()), config)
	return config
}

// GetKafka 获取Kafka配置
func (d *Config) GetKafka() *conf.Kafka {
	config := &conf.Kafka{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.String()), config)
	return config
}
