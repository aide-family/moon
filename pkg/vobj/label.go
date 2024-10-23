package vobj

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/aide-family/moon/pkg/util/types"

	"golang.org/x/exp/maps"
)

const (
	// StrategyID 策略id
	StrategyID = "__moon__strategy_id__"
	// LevelID 策略级别id
	LevelID = "__moon__level_id__"
	// TeamID 团队id
	TeamID = "__moon__team_id__"
	// DatasourceID 数据源id
	DatasourceID = "__moon__datasource_id__"
	// DatasourceURL 数据源url
	DatasourceURL = "__moon__datasource_url__"
	// Domain 域名
	Domain = "__moon__domain__"
	// DomainSubject 域名主题
	DomainSubject = "__moon__domain_subject__"
	// DomainExpiresOn 域名过期时间
	DomainExpiresOn = "__moon__domain_expires_on__"
	// DomainPort 端口
	DomainPort = "__moon__domain_port__"
)

var _ sql.Scanner = (*Labels)(nil)
var _ driver.Valuer = (*Labels)(nil)

// ErrUnsupportedType 不支持的类型错误
var ErrUnsupportedType = errors.New("unsupported type")

// Labels 标签
type Labels struct {
	label map[string]string
}

// NewLabels 基于map创建Labels
func NewLabels(labels map[string]string) *Labels {
	return &Labels{label: labels}
}

// MarshalJSON 实现 json.Marshaler 接口
func (l *Labels) MarshalJSON() ([]byte, error) {
	// 返回字符串形式的时间
	return []byte(l.String()), nil
}

// String 转json字符串
func (l *Labels) String() string {
	if types.IsNil(l) || l.label == nil {
		return "{}"
	}
	bs := strings.Builder{}
	bs.WriteString(`{`)
	labelKeys := maps.Keys(l.label)
	sort.Strings(labelKeys)
	for _, k := range labelKeys {
		bs.WriteString(`"` + k + `":"` + l.label[k] + `",`)
	}
	str := strings.TrimRight(bs.String(), ",")
	return str + "}"
}

// Map 转map
func (l *Labels) Map() map[string]string {
	if l == nil || l.label == nil {
		return make(map[string]string)
	}

	return l.label
}

// Get 获取value
func (l *Labels) Get(key string) string {
	if l == nil || l.label == nil {
		return ""
	}
	return l.label[key]
}

// Has 判断是否存在
func (l *Labels) Has(key string) bool {
	if l == nil || l.label == nil {
		return false
	}
	_, ok := l.label[key]
	return ok
}

// Match 判断value是否满足正则字符串
func (l *Labels) Match(key, reg string) bool {
	if l == nil || l.label == nil || !l.Has(key) {
		return false
	}
	matched, err := regexp.MatchString(reg, l.Get(key))
	if err != nil {
		return false
	}
	return matched
}

// Append 追加
func (l *Labels) Append(key, val string) *Labels {
	l.label[key] = val
	return l
}

// AppendMap 追加map
func (l *Labels) AppendMap(m map[string]string) *Labels {
	for k, v := range m {
		l.label[k] = v
	}
	return l
}

// Index 索引
func (l *Labels) Index() string {
	str := strings.Builder{}
	str.WriteString("{")
	keys := maps.Keys(l.label)
	// 排序
	sort.Strings(keys)
	for _, key := range keys {
		k := key
		v := l.label[key]
		str.WriteString(`"` + k + `"`)
		str.WriteString(":")
		str.WriteString(`"` + v + `"`)
		str.WriteString(",")
	}
	return strings.TrimRight(str.String(), ",") + "}"
}

// Value 实现 driver.Valuer 接口
func (l Labels) Value() (driver.Value, error) {
	return types.Marshal(l.label)
}

// Scan 实现 sql.Scanner 接口
func (l *Labels) Scan(src any) (err error) {
	switch src.(type) {
	case []byte:
		err = types.Unmarshal(src.([]byte), &l.label)
	case string:
		err = types.Unmarshal([]byte(src.(string)), &l.label)
	default:
		err = ErrUnsupportedType
	}
	return err
}

// GetDatasourceID 获取数据源id
func (l *Labels) GetDatasourceID() uint32 {
	id, _ := strconv.ParseUint(l.Get(DatasourceID), 10, 32)
	return uint32(id)
}

// GetDatasourceURL 获取数据源url
func (l *Labels) GetDatasourceURL() string {
	return l.Get(DatasourceURL)
}

// GetTeamID 获取团队id
func (l *Labels) GetTeamID() uint32 {
	id, _ := strconv.ParseUint(l.Get(TeamID), 10, 32)
	return uint32(id)
}

// GetStrategyID 获取策略id
func (l *Labels) GetStrategyID() uint32 {
	id, _ := strconv.ParseUint(l.Get(StrategyID), 10, 32)
	return uint32(id)
}

// GetLevelID 获取策略级别id
func (l *Labels) GetLevelID() uint32 {
	id, _ := strconv.ParseUint(l.Get(LevelID), 10, 32)
	return uint32(id)
}
