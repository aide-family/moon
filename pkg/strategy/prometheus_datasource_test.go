package strategy

import (
	"context"
	"testing"
	"time"
)

const defaultDatasource = "https://prom-server.aide-cloud.cn"

func TestNewDatasource(t *testing.T) {
	d := NewPrometheusDatasource(defaultDatasource)
	queryResponse, err := d.Query(context.Background(), "up == 1", time.Now().Unix())
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(queryResponse.Status)
	t.Log(queryResponse.Data)
	for _, v := range queryResponse.Data.Result {
		t.Log(v.Metric.Name())
		t.Log(v.Metric.String())
		t.Log(v.Value)
	}
	t.Log(queryResponse.String())
}
