package strategy

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

const (
	strategyId  = "strategy_id"
	levelId     = "level_id"
	summary     = "summary"
	description = "description"
	instance    = "instance"
)

var (
	_ fmt.Stringer = (*Labels)(nil)
	_ Label        = (*Labels)(nil)
	_ fmt.Stringer = (*Annotations)(nil)
	_ Annotation   = (*Annotations)(nil)
)

type (
	Label interface {
		StrategyId() uint
		LevelId() uint
		Get(key string) string
		GetInstance() string
		Map() map[string]string
		String() string
		sql.Scanner
		driver.Valuer
	}

	Annotation interface {
		Summary() string
		Description() string
		Get(key string) string
		Map() map[string]string
		String() string
		sql.Scanner
		driver.Valuer
	}

	Groups struct {
		Groups []*Group `json:"groups"`
	}

	Group struct {
		Name  string  `json:"name"`
		Rules []*Rule `json:"rules"`
	}

	Rule struct {
		Alert       string      `json:"alert"`
		Expr        string      `json:"expr"`
		For         string      `json:"for"`
		Labels      Labels      `json:"labels"`
		Annotations Annotations `json:"annotations"`
		// 数据源
		endpoint string
		lock     sync.RWMutex
	}

	Labels      map[string]string
	Annotations map[string]string
)

// SetEndpoint 设置数据源
func (r *Rule) SetEndpoint(endpoint string) {
	if r == nil {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.endpoint = endpoint
}

// Endpoint 获取数据源
func (r *Rule) Endpoint() string {
	if r == nil {
		return ""
	}
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.endpoint
}

func (l *Annotations) Map() map[string]string {
	if l == nil {
		return nil
	}

	return *l
}

func (l *Labels) Map() map[string]string {
	if l == nil {
		return nil
	}
	return *l
}

func (l *Annotations) Scan(src any) error {
	if l == nil || src == nil {
		return nil
	}
	if srcBytes, ok := src.([]byte); ok {
		return json.Unmarshal(srcBytes, l)
	}
	return nil
}

func (l *Annotations) Value() (driver.Value, error) {
	if l == nil {
		return "{}", nil
	}
	return json.Marshal(l)
}

func (l *Labels) Scan(src any) error {
	if l == nil || src == nil {
		return nil
	}
	if srcBytes, ok := src.([]byte); ok {
		return json.Unmarshal(srcBytes, l)
	}
	return nil
}

func (l *Labels) Value() (driver.Value, error) {
	if l == nil {
		return "{}", nil
	}
	return json.Marshal(l)
}

func (l *Labels) GetInstance() string {
	if l == nil {
		return ""
	}
	return (*l)[instance]
}

func (l *Annotations) Get(key string) string {
	if l == nil {
		return ""
	}
	return (*l)[key]
}

func (l *Labels) Get(key string) string {
	if l == nil {
		return ""
	}
	return (*l)[key]
}

func (l *Annotations) Summary() string {
	if l == nil {
		return ""
	}
	return (*l)[summary]
}

func (l *Annotations) Description() string {
	if l == nil {
		return ""
	}
	return (*l)[description]
}

func (l *Labels) LevelId() uint {
	if l == nil {
		return 0
	}
	if id, ok := (*l)[levelId]; ok {
		uid, _ := strconv.Atoi(strings.TrimSpace(id))
		return uint(uid)
	}
	return 0
}

func (l *Labels) StrategyId() uint {
	if l == nil {
		return 0
	}
	if id, ok := (*l)[strategyId]; ok {
		uid, _ := strconv.Atoi(strings.TrimSpace(id))
		return uint(uid)
	}
	return 0
}

func (l *Annotations) String() string {
	if l == nil {
		return ""
	}
	bytes, _ := json.Marshal(l)
	return string(bytes)
}

func (l *Labels) String() string {
	if l == nil {
		return ""
	}
	bytes, _ := json.Marshal(l)
	return string(bytes)
}

// MapToLabels 将map转换为标签
func MapToLabels(m map[string]string) *Labels {
	labels := Labels(m)
	return &labels
}

// ToLabels 将字符串转换为标签
func ToLabels(str string) *Labels {
	labels := make(Labels)
	_ = json.Unmarshal([]byte(str), &labels)
	return &labels
}

// MapToAnnotations 将map转换为注解
func MapToAnnotations(m map[string]string) *Annotations {
	annotations := Annotations(m)
	return &annotations
}

// ToAnnotations 将字符串转换为注解
func ToAnnotations(str string) *Annotations {
	annotations := make(Annotations)
	_ = json.Unmarshal([]byte(str), &annotations)
	return &annotations
}
