package strategy

import (
	"testing"
)

func TestStore_StoreStrategy(t *testing.T) {
	source := Groups{Groups: []*Group{
		{
			Name: "test_1",
			Rules: []*Rule{{
				Alert: "node_load1",
				Expr:  "node_load1{scrape_job=\"node_export_qingcdn\"} > 45",
				For:   "3m",
				Labels: map[string]string{
					"severity": "critical",
				},
				Annotations: map[string]string{
					"summary": "{{ $labels.instance }} has high load",
					"message": "node_load1 {{ $value }}",
				},
			}},
		},
	}}
	store := NewStrategyStore(".")

	if err := store.Store(&source); err != nil {
		t.Error(err)
		return
	}
}
