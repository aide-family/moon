package fingerprint

import "testing"

func TestOfStable(t *testing.T) {
	left := Of("rule-1", map[string]string{"job": "api", "instance": "a"})
	right := Of("rule-1", map[string]string{"instance": "a", "job": "api"})
	if left != right {
		t.Fatalf("fingerprint changed with key order: %s %s", left, right)
	}
	if Of("rule-1", map[string]string{"job": "api"}) == Of("rule-2", map[string]string{"job": "api"}) {
		t.Fatal("different rules share a fingerprint")
	}
}
