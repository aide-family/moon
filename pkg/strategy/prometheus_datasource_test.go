package strategy

import (
	"context"
	"testing"
	"time"
)

const defaultDatasource = "https://prom-server.aide-cloud.cn"

var expr = "up == 1"
var groupInfo = &Group{
	Name: "test-group",
	Id:   1,
}

var ruleInfo = &Rule{
	Id:    1,
	Alert: "test-alert",
	Expr:  expr,
	For:   "3m",
	Labels: map[string]string{
		metricLevelId: "1",
	},
	Annotations: map[string]string{
		metricSummary:     "instance {{ $labels.instance }} is up",
		metricDescription: "This value is {{ $value }}",
	},
	endpoint:       defaultDatasource,
	datasourceName: string(PrometheusDatasource),
}

func TestNewPrometheusDatasource(t *testing.T) {
	d := NewPrometheusDatasource(defaultDatasource)
	queryResponse, err := d.Query(context.Background(), expr, time.Now().Unix())
	if err != nil {
		t.Error(err)
		return
	}

	alarm := NewAlarm(groupInfo, ruleInfo, queryResponse.Data.Result)

	t.Log(string(alarm.Bytes()))
}
