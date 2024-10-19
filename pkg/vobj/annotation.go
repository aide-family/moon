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
	SummaryKey     = "summary"
	DescriptionKey = "description"
)

var _ sql.Scanner = (*Annotations)(nil)
var _ driver.Valuer = (*Annotations)(nil)

// Annotations 告警文案
type Annotations map[string]string

// Scan implements the sql.Scanner interface.
func (l *Annotations) Scan(src any) error {
	switch src.(type) {
	case []byte:
		return types.Unmarshal(src.([]byte), &l)
	case string:
		return types.Unmarshal([]byte(src.(string)), l)
	default:
		return ErrUnsupportedType
	}
}

// MarshalJSON 实现 json.Marshaler 接口
func (l *Annotations) MarshalJSON() ([]byte, error) {
	// 返回字符串形式的时间
	return []byte(l.String()), nil
}

func (l *Annotations) String() string {
	if types.IsNil(l) || l == nil {
		return "{}"
	}
	bs := strings.Builder{}
	bs.WriteString(`{`)
	labelKeys := maps.Keys(*l)
	sort.Strings(labelKeys)
	for _, k := range labelKeys {
		bs.WriteString(`"` + k + `":"` + map[string]string(*l)[k] + `",`)
	}
	str := strings.TrimRight(bs.String(), ",")
	return str + "}"
}

func (l Annotations) Get(key string) string {
	if types.IsNil(l) || l == nil {
		return ""
	}
	return l[key]
}

func (l Annotations) Set(key, value string) {
	if types.IsNil(l) {
		l = make(map[string]string)
	}
	l[key] = value
}

// GetSummary 获取摘要
func (l *Annotations) GetSummary() string {
	return l.Get(SummaryKey)
}

// GetDescription 获取描述
func (l *Annotations) GetDescription() string {
	return l.Get(DescriptionKey)
}

// Value implements the driver.Valuer interface.
func (l Annotations) Value() (driver.Value, error) {
	return types.Marshal(l)
}
