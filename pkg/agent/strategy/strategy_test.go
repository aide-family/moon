package strategy_test

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/agent"
	nutsdbCache "github.com/aide-family/moon/pkg/agent/cacher/nutsdbcache"
	"github.com/aide-family/moon/pkg/agent/datasource/p8s"
	"github.com/aide-family/moon/pkg/agent/strategy"
	"github.com/nutsdb/nutsdb"
)

func init() {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir("./cache"), // 数据库会自动创建这个目录文件
	)
	if err != nil {
		panic(err)
	}

	cache := nutsdbCache.NewNutsDbCache(db, "default_bucket")

	agent.SetGlobalCache(cache)
}

var expr = `node_cpu_seconds_total{app="prometheus", app_kubernetes_io_managed_by="Helm", chart="prometheus-15.9.2", component="node-exporter", cpu="0", heritage="Helm", instance="10.0.4.5:9100", job="kubernetes-service-endpoints", mode="idle", namespace="pixiu-system", node="vm-4-5-centos", release="prometheus", service="prometheus-node-exporter"} == 0`

func getRule() *strategy.EvalRule {
	rule := &strategy.EvalRule{
		ID:    "1",
		Alert: "test",
		Expr:  expr,
		For:   "10s",
		Labels: map[string]string{
			"severity": "critical",
		},
		Annotations: map[string]string{
			"summary":     "test {{ $value }}",
			"description": "test description {{ $value }}",
		},
	}
	datasourceOpts := []p8s.Option{
		p8s.WithEndpoint("https://prom-server.aide-cloud.cn/"),
	}
	rule.SetDatasource(p8s.NewPrometheusDatasource(datasourceOpts...))
	return rule
}

func TestEvalRule(t *testing.T) {
	count := 0
	for count < 10 {
		count++
		rule := getRule()
		alarm, err := rule.Eval(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		t.Log(alarm.String())
		time.Sleep(1 * time.Second)
		if alarm == nil {
			continue
		}
	}

	cache := agent.GetGlobalCache()
	defer cache.Close()

	rule := getRule()
	var alarm agent.Alarm
	if err := cache.Get(rule.GetID(), &alarm); err != nil {
		t.Fatal(err)
	}
	t.Log(alarm.String())
}
