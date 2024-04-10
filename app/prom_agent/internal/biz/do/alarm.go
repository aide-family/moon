package do

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/strategy"
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
