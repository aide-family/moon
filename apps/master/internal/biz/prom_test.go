package biz

import (
	"prometheus-manager/api/prom"
	"testing"
)

func TestEnum(t *testing.T) {
	t.Log(prom.Status_ENABLE.String())
	t.Log(prom.Status_ENABLE.Descriptor())
	t.Log(prom.Status_ENABLE.Number())
	t.Log(prom.Status_ENABLE)
}
