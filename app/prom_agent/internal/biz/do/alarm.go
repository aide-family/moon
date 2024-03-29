package do

import (
	"encoding/json"

	"prometheus-manager/pkg/strategy"
)

type (
	AlarmDo struct {
		*strategy.Alarm
	}
)

func (a *AlarmDo) Bytes() []byte {
	if a == nil {
		return []byte("{}")
	}
	bs, _ := json.Marshal(a)
	return bs
}
