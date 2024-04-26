package strategy

import (
	"strconv"
	"sync"
	"time"

	"github.com/aide-family/moon/pkg/agent"
)

type (
	Ruler interface {
		agent.Eval
		GetDatasource() agent.Datasource
	}

	EvalRule struct {
		// ID 用作唯一标识
		ID          string            `json:"-"`
		Alert       string            `json:"alert"`
		Expr        string            `json:"expr"`
		For         string            `json:"for"`
		Labels      agent.Labels      `json:"labels"`
		Annotations agent.Annotations `json:"annotations"`

		datasource agent.Datasource `json:"-"`
		eventsAt   time.Time        `json:"-"`
		mut        sync.RWMutex     `json:"-"`
	}

	EvalGroup struct {
		Name  string      `json:"name"`
		Rules []*EvalRule `json:"rules"`
	}
)

const (
	LabelsKeyDatasourceID = "datasourceId"
	LabelsKeyRuleID       = "__alert_id__"
	LabelsKeyLevelID      = "__level_id__"
	LabelsKeyGroupID      = "groupID"
)

// GetID get rule id
func (e *EvalRule) GetID() string {
	if e == nil {
		return ""
	}
	return e.ID
}

// GetAlert get alert name
func (e *EvalRule) GetAlert() string {
	if e == nil {
		return ""
	}
	return e.Alert
}

// GetExpr get expr
func (e *EvalRule) GetExpr() string {
	if e == nil {
		return ""
	}
	return e.Expr
}

func (e *EvalRule) GetFor() string {
	if e == nil {
		return ""
	}
	return e.For
}

func (e *EvalRule) GetLabels() agent.Labels {
	if e == nil {
		return nil
	}
	return e.Labels
}

func (e *EvalRule) GetAnnotations() agent.Annotations {
	if e == nil {
		return nil
	}
	return e.Annotations
}

func (e *EvalRule) GetDatasource() agent.Datasource {
	if e == nil {
		return nil
	}
	return e.datasource
}

// SetDatasource set datasource
func (e *EvalRule) SetDatasource(datasource agent.Datasource) {
	if e == nil {
		return
	}
	e.mut.Lock()
	defer e.mut.Unlock()
	e.datasource = datasource
}

// GetName get group name
func (g *EvalGroup) GetName() string {
	if g == nil {
		return ""
	}
	return g.Name
}

func (g *EvalGroup) GetRules() []*EvalRule {
	if g == nil {
		return nil
	}
	return g.Rules
}

// SetEventsAt set events at
func (e *EvalRule) SetEventsAt(eventsAt time.Time) {
	if e == nil {
		return
	}
	e.mut.Lock()
	defer e.mut.Unlock()
	e.eventsAt = eventsAt
}

// ForEventsAt 断言eventsAt是否在for时间范围内
func (e *EvalRule) ForEventsAt(startsAt time.Time) bool {
	if e == nil {
		return false
	}
	return startsAt.Add(BuildDuration(e.For)).Before(time.Now())
}

// BuildDuration 字符串转为api时间
func BuildDuration(duration string) time.Duration {
	durationLen := len(duration)
	if duration == "" || durationLen < 2 {
		return 0
	}
	value, _ := strconv.Atoi(duration[:durationLen-1])
	// 获取字符串最后一个字符
	unit := string(duration[durationLen-1])
	valueDuration := time.Duration(value)
	switch unit {
	case "s":
		return valueDuration * time.Second
	case "m":
		return valueDuration * time.Minute
	case "h":
		return valueDuration * time.Hour
	case "d":
		return valueDuration * time.Hour * 24
	default:
		return 0
	}
}
