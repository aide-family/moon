package vobj

import (
	"database/sql"
	"database/sql/driver"
	"sort"
	"strings"

	"github.com/aide-family/moon/pkg/util/types"
	"golang.org/x/exp/maps"
)

const (
	// SummaryKey 告警摘要
	SummaryKey = "summary"
	// DescriptionKey 告警描述
	DescriptionKey = "description"
)

// NewAnnotations 返回一个新的 Annotations 对象
func NewAnnotations(annotations map[string]string) *Annotations {
	if annotations == nil {
		annotations = make(map[string]string)
	}
	return (*Annotations)(&annotations)
}

var _ sql.Scanner = (*Annotations)(nil)
var _ driver.Valuer = (*Annotations)(nil)

// Annotations 告警文案
type Annotations map[string]string

// Map 转换为 map
func (l *Annotations) Map() map[string]string {
	if l == nil {
		return nil
	}
	return *l
}

// Scan 实现 sql.Scanner 接口
func (l *Annotations) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		return types.Unmarshal(s, l)
	case string:
		return types.Unmarshal([]byte(s), l)
	default:
		return ErrUnsupportedType
	}
}

// MarshalJSON 实现 json.Marshaler 接口
func (l *Annotations) MarshalJSON() ([]byte, error) {
	// 返回字符串形式的时间
	return []byte(l.String()), nil
}

// String 字符串
func (l *Annotations) String() string {
	if types.IsNil(l) || len(*l) == 0 {
		return "{}"
	}
	bs := strings.Builder{}
	bs.WriteString(`{`)
	labelKeys := maps.Keys(*l)
	sort.Strings(labelKeys)
	list := make([]string, 0, len(labelKeys)*5)
	list = append(list, "{")
	for _, k := range labelKeys {
		list = append(list, `"`, k, `":"`, (*l)[k], `"`, ",")
	}
	list = append(list[:len(list)-1], "}")
	return types.TextJoin(list...)
}

// Get 获取 key 对应的 value
func (l *Annotations) Get(key string) string {
	if types.IsNil(l) {
		return ""
	}
	return (*l)[key]
}

// Set 设置 key 对应的 value
func (l *Annotations) Set(key, value string) {
	if types.IsNil(l) {
		return
	}
	(*l)[key] = value
}

// GetSummary 获取摘要
func (l *Annotations) GetSummary() string {
	return l.Get(SummaryKey)
}

// GetDescription 获取描述
func (l *Annotations) GetDescription() string {
	return l.Get(DescriptionKey)
}

// Value 实现 driver.Valuer 接口
func (l *Annotations) Value() (driver.Value, error) {
	return l.String(), nil
}
