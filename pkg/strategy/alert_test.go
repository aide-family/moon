package strategy

import (
	"context"
	"testing"
	"time"
)

func TestBuildDuration(t *testing.T) {
	t.Log(BuildDuration("10s"))
	t.Log(BuildDuration("10m"))
	t.Log(BuildDuration("10h"))
	t.Log(BuildDuration("10d"))
}

func TestNewAlerting(t *testing.T) {
	rule := &Rule{
		Alert: "test-alert",
		Expr:  "up == 1",
		For:   "3s",
		Labels: map[string]string{
			"job": "test-job",
		},
		Annotations: map[string]string{
			"summary":     "test-summary",
			"description": "test-description",
		},
		endpoint: "",
	}
	rule.SetEndpoint(defaultDatasource)

	group := &Group{
		Name:  "test-group",
		Rules: []*Rule{rule},
	}
	a := NewAlerting(group, PrometheusDatasource, nil)

	for _, r := range group.Rules {
		results, err := a.Eval(context.Background(), r)
		if err != nil {
			t.Error(err)
		}
		t.Log(results)
	}
	time.Sleep(time.Second * 60)
}
