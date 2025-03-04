package logs

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/loki-client-go/loki"
	"github.com/prometheus/common/model"
)

func TestPushLogsLokiHandle(t *testing.T) {

	config, err := loki.NewDefaultConfig("http://localhost:3100/loki/api/v1/push")

	if err != nil {
		t.Errorf("failed to create config: %v", err)
		return
	}
	client, err := loki.New(config)

	if err != nil {
		t.Errorf("failed to create client: %v", err)
	}

	defer client.Stop()
	labels := model.LabelSet{
		"job":      "test_app",
		"instance": "host-01",
	}
	// 4. 推送日志
	err = client.Handle(
		labels,
		time.Now(),
		"测试日志内容",
	)

	if err != nil {
		t.Errorf("failed to push logs: %v", err)
	}
}

func TestQueryLogs(t *testing.T) {

	lokiData := NewLokiDatasource(WithLokiEndpoint("http://127.0.0.1:3100"))

	queryLogs, err := lokiData.QueryLogs(context.Background(), "{job=~\".+\"}", time.Now().Add(-time.Hour).Unix(), time.Now().Unix())
	if err != nil {
		t.Fatalf("failed to query logs: %v", err)
		return
	}

	values := queryLogs.Values

	t.Log(values)
}
