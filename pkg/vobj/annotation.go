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

// NewAnnotations returns a new Annotations object.
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

// Map converts the Annotations object to a map.
func (l *Annotations) Map() map[string]string {
	if l == nil {
		return nil
	}
	return *l
}

// Scan implements the sql.Scanner interface.
func (l *Annotations) Scan(src any) error {
	switch src.(type) {
	case []byte:
		return types.Unmarshal(src.([]byte), l)
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
	if types.IsNil(l) {
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

func (l *Annotations) Get(key string) string {
	if types.IsNil(l) {
		return ""
	}
	return (*l)[key]
}

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

// Value implements the driver.Valuer interface.
func (l *Annotations) Value() (driver.Value, error) {
	return l.String(), nil
}
