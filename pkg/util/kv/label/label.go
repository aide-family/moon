package label

import (
	"encoding/json"
	"strconv"

	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/template"
)

var _ json.Marshaler = (*Label)(nil)
var _ json.Unmarshaler = (*Label)(nil)

func NewLabel(labels map[string]string) *Label {
	return &Label{
		kvMap: labels,
	}
}

type Label struct {
	kvMap kv.StringMap
}

func (a *Label) ToMap() map[string]string {
	return a.kvMap.ToMap()
}

func (a *Label) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &a.kvMap)
}

func (a *Label) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.kvMap)
}

func (a *Label) String() string {
	bs, _ := a.MarshalBinary()
	return string(bs)
}

func (a *Label) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a.kvMap)
}

func (a *Label) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &a.kvMap)
}

func (a *Label) Copy() *Label {
	return &Label{
		kvMap: a.kvMap.Copy(),
	}
}

func (a *Label) Appends(labels map[string]string) *Label {
	for k, v := range labels {
		a.kvMap.Set(k, v)
	}
	return a
}

func (a *Label) GetStrategyId() uint32 {
	v, ok := a.kvMap.Get(cnst.LabelKeyStrategyID)
	if !ok {
		return 0
	}
	strategyId, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(strategyId)
}

func (a *Label) SetStrategyId(strategyId uint32) {
	a.kvMap.Set(cnst.LabelKeyStrategyID, strconv.FormatUint(uint64(strategyId), 10))
}

func (a *Label) GetTeamId() uint32 {
	v, ok := a.kvMap.Get(cnst.LabelKeyTeamID)
	if !ok {
		return 0
	}
	teamId, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(teamId)
}

func (a *Label) SetTeamId(teamId uint32) {
	a.kvMap.Set(cnst.LabelKeyTeamID, strconv.FormatUint(uint64(teamId), 10))
}

func (a *Label) GetDatasourceId() uint32 {
	v, ok := a.kvMap.Get(cnst.LabelKeyDatasourceID)
	if !ok {
		return 0
	}
	datasourceId, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(datasourceId)
}

func (a *Label) SetDatasourceId(datasourceId uint32) {
	a.kvMap.Set(cnst.LabelKeyDatasourceID, strconv.FormatUint(uint64(datasourceId), 10))
}

func (a *Label) GetLevelId() uint32 {
	v, ok := a.kvMap.Get(cnst.LabelKeyLevelID)
	if !ok {
		return 0
	}
	levelId, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(levelId)
}

func (a *Label) SetLevelId(levelId uint32) {
	a.kvMap.Set(cnst.LabelKeyLevelID, strconv.FormatUint(uint64(levelId), 10))
}

func (a *Label) Format(data interface{}) *Label {
	for k, v := range a.kvMap.ToMap() {
		a.kvMap.Set(k, template.TextFormatterX(v, data))
	}
	return a
}
