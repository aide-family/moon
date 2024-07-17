package vobj

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/aide-family/moon/pkg/util/types"
	"golang.org/x/exp/maps"
)

const (
	StrategyID = "strategy_id"
)

var _ sql.Scanner = (*Labels)(nil)
var _ driver.Valuer = (*Labels)(nil)

var ErrUnsupportedType = errors.New("unsupported type")

type Labels struct {
	label map[string]string
}

type LabelsJSON map[string]string

func NewLabels(labels map[string]string) *Labels {
	return &Labels{label: labels}
}

func (l *Labels) String() string {
	if types.IsNil(l) || l.label == nil {
		return "{}"
	}
	bs, _ := json.Marshal(l.label)
	return string(bs)
}

func (l LabelsJSON) String() string {
	if types.IsNil(l) {
		return "{}"
	}
	bs, _ := json.Marshal(l)
	return string(bs)
}

func (l *Labels) Map() map[string]string {
	if l == nil || l.label == nil {
		return make(map[string]string)
	}

	return l.label
}

func (l *Labels) Append(key, val string) *Labels {
	l.label[key] = val
	return l
}

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

func (l Labels) Value() (driver.Value, error) {
	return json.Marshal(l.label)
}

func (l *Labels) Scan(src any) error {
	switch src.(type) {
	case []byte:
		return json.Unmarshal(src.([]byte), &l.label)
	case string:
		return json.Unmarshal([]byte(src.(string)), &l.label)
	default:
		return ErrUnsupportedType
	}
}
