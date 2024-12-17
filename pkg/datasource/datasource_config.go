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

var _ sql.Scanner = (*DatasourceConfig)(nil)
var _ driver.Valuer = (*DatasourceConfig)(nil)

// DatasourceConfig 数据源配置
type DatasourceConfig struct {
	datasourceConfig map[string]string
}

// Scan 实现 sql.Scanner 接口
func (d *DatasourceConfig) Scan(src any) (err error) {
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
func (d *DatasourceConfig) Value() (driver.Value, error) {
	return d.String(), nil
}

// NewDatasourceConfig 基于map创建DatasourceConfig
func NewDatasourceConfig(datasourceConfig map[string]string) *DatasourceConfig {
	return &DatasourceConfig{datasourceConfig: datasourceConfig}
}

// MarshalJSON 实现 json.Marshaler 接口
func (d *DatasourceConfig) MarshalJSON() ([]byte, error) {
	// 返回字符串形式的时间
	return []byte(d.String()), nil
}

// String 转json字符串
func (d *DatasourceConfig) String() string {
	if types.IsNil(d) || d.datasourceConfig == nil {
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
func (d *DatasourceConfig) Map() map[string]string {
	if d == nil || d.datasourceConfig == nil {
		return make(map[string]string)
	}
	return d.datasourceConfig
}

// GetRocketMQ 获取RocketMQ配置
func (d *DatasourceConfig) GetRocketMQ() *conf.RocketMQ {
	config := &conf.RocketMQ{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.String()), config)
	return config
}

// GetMQTT 获取MQTT配置
func (d *DatasourceConfig) GetMQTT() *conf.MQTT {
	config := &conf.MQTT{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.String()), config)
	return config
}

// GetKafka 获取Kafka配置
func (d *DatasourceConfig) GetKafka() *conf.Kafka {
	config := &conf.Kafka{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.String()), config)
	return config
}

// GetRabbit 获取RabbitMQ配置
func (d *DatasourceConfig) GetRabbit() *conf.RabbitMQ {
	config := &conf.RabbitMQ{}
	if d == nil || d.datasourceConfig == nil {
		return config
	}
	_ = json.Unmarshal([]byte(d.String()), config)
	return config
}
