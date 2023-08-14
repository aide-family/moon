package biz

import (
	"testing"

	"prometheus-manager/api/prom"
)

func TestEnum(t *testing.T) {
	t.Log(prom.Status_Status_ENABLE.String())
	t.Log(prom.Status_Status_ENABLE.Descriptor())
	t.Log(prom.Status_Status_ENABLE.Number())
	t.Log(prom.Status_Status_ENABLE)
}
