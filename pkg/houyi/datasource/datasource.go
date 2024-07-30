package datasource

import (
	"context"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

type (
	// Value 数据源查询值
	Value struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"`
	}

	// Point 数据点
	Point struct {
		// 标签集合
		Labels *vobj.Labels `json:"labels"`
		// 值
		Values []*Value `json:"value"`
	}
)

// Datasource 数据源通用接口
type Datasource interface {
	Eval(ctx context.Context, expr string, step uint32) (map[watch.Indexer]*Point, error)
	Step() uint32
}

// NewDatasource 根据配置创建对应的数据源
func NewDatasource(config *api.Datasource) Datasource {
	// TODO 根据配置创建对应的数据源
	return NewMockDatasource()
}

type mockDatasource struct {
}

func (m *mockDatasource) Eval(_ context.Context, _ string, _ uint32) (map[watch.Indexer]*Point, error) {
	res := make(map[watch.Indexer]*Point)
	labels := vobj.NewLabels(map[string]string{"env": "mock"})
	values := make([]*Value, 0, 100)
	for i := 0; i < 100; i++ {
		values = append(values, &Value{
			Value:     float64(i + 1),
			Timestamp: time.Now().Unix(),
		})
	}
	res[labels] = &Point{
		Labels: labels,
		Values: values,
	}
	return res, nil
}

func (m *mockDatasource) Step() uint32 {
	return 10
}

// NewMockDatasource 创建一个mock数据源
func NewMockDatasource() Datasource {
	return &mockDatasource{}
}
