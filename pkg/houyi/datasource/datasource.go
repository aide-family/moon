package datasource

import (
	"context"

	"github.com/aide-family/moon/pkg/vobj"
)

type (
	Value struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"`
	}

	Point struct {
		// 标签集合
		Labels vobj.Labels `json:"labels"`
		// 值
		Value *Value `json:"value"`
	}
)

type Datasource interface {
	Eval(ctx context.Context, expr string) ([]*Point, error)
}
