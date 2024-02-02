package strategy

import (
	"context"
	"testing"
	"time"
)

func TestNewDatasource(t *testing.T) {
	d := NewDatasource(PrometheusDatasource, defaultDatasource)
	queryResponse, err := d.Query(context.Background(), expr, time.Now().Unix())
	if err != nil {
		t.Error(err)
		return
	}
	alarm, _, _ := NewAlarm(&Group{
		Name: "test-group",
		Id:   1,
	}, &Rule{
		Id:    1,
		Alert: "test-alert",
		Expr:  expr,
		For:   "3m",
		Labels: map[string]string{
			MetricLevelId: "1",
		},
		Annotations: map[string]string{
			MetricSummary:     "instance {{ $labels.instance }} is up",
			MetricDescription: "This value is {{ $value }} {{ $labels.__name__ }} {{ $labels.__name__ }}",
		},
		endpoint:       defaultDatasource,
		datasourceName: string(PrometheusDatasource),
	}, queryResponse.Data.Result)

	t.Log(string(alarm.Bytes()))
}
