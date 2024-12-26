package bo

import (
	"strconv"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*EventDatasource)(nil)

// EventDatasource MQ 数据源配置
type EventDatasource struct {
	TeamID uint32      `json:"team_id"`
	ID     uint32      `json:"id"`
	Status vobj.Status `json:"status"`

	Conf *conf.Event `json:"conf"`
}

// Index 实现 watch.Indexer 接口
func (m *EventDatasource) Index() string {
	return types.TextJoin(strconv.Itoa(int(m.TeamID)), ":", strconv.Itoa(int(m.ID)))
}

// String 实现 fmt.Stringer 接口
func (m *EventDatasource) String() string {
	bs, _ := types.Marshal(m)
	return string(bs)
}

// GetConfig 获取 MQ 配置
func (m *EventDatasource) GetConfig() *conf.Event {
	if m == nil {
		return nil
	}
	return m.Conf
}
