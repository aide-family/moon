package strategy

import (
	"strconv"
	"testing"
)

func TestRuleLabelString(t *testing.T) {
	labels := Labels{
		"test":  "test",
		"test2": "test2",
		"tt":    "xx",
	}
	for i := 0; i < 100; i++ {
		labels[strconv.Itoa(i)] = strconv.Itoa(i)
	}

	for i := 0; i < 1000; i++ {
		if labels.String() != labels.String() {
			t.Error("label string error")
		}

		rule := &Rule{
			Labels: labels,
		}

		if rule.MD5() != rule.MD5() {
			t.Error("rule md5 error")
		}
	}
}
