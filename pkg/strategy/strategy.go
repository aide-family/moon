package strategy

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	}

	Annotation interface {
		Summary() string
		Description() string
		Get(key string) string
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
	}

	Labels      map[string]string
	Annotations map[string]string
)

func (l Annotations) Get(key string) string {
	return l[key]
}

func (l Labels) Get(key string) string {
	return l[key]
}

const (
	strategyId  = "strategy_id"
	levelId     = "level_id"
	summary     = "summary"
	description = "description"
)

func (l Annotations) Summary() string {
	if s, ok := l[summary]; ok {
		return s
	}
	return ""
}

func (l Annotations) Description() string {
	if s, ok := l[description]; ok {
		return s
	}
	return ""
}

func (l Labels) LevelId() uint {
	if id, ok := l[levelId]; ok {
		uid, _ := strconv.Atoi(strings.TrimSpace(id))
		return uint(uid)
	}
	return 0
}

func (l Labels) StrategyId() uint {
	if id, ok := l[strategyId]; ok {
		uid, _ := strconv.Atoi(strings.TrimSpace(id))
		return uint(uid)
	}
	return 0
}

func (l Annotations) String() string {
	bytes, _ := json.Marshal(l)
	return string(bytes)
}

func (l Labels) String() string {
	bytes, _ := json.Marshal(l)
	return string(bytes)
}

// ToLabels 将字符串转换为标签
func ToLabels(str string) Labels {
	labels := make(map[string]string)
	if err := json.Unmarshal([]byte(str), &labels); err != nil {
		return nil
	}
	return labels
}

// ToAnnotations 将字符串转换为注解
func ToAnnotations(str string) Annotations {
	annotations := make(map[string]string)
	if err := json.Unmarshal([]byte(str), &annotations); err != nil {
		return nil
	}
	return annotations
}
