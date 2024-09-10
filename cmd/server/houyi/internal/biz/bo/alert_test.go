package bo

import (
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

func TestAlert_AlertJson(t *testing.T) {
	alert := &Alert{
		Status:       vobj.AlertStatusFiring,
		Labels:       vobj.NewLabels(map[string]string{}),
		Annotations:  make(vobj.Annotations),
		StartsAt:     types.NewTime(time.Now()),
		EndsAt:       &types.Time{},
		GeneratorURL: "",
		Fingerprint:  "",
		Value:        0,
	}
	t.Log(alert)
}