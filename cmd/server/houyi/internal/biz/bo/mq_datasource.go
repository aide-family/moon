package bo

import (
	"strconv"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*MQDatasource)(nil)

// MQDatasource MQ 数据源配置
type MQDatasource struct {
	TeamID uint32      `json:"team_id"`
	ID     uint32      `json:"id"`
	Status vobj.Status `json:"status"`

	Conf *conf.MQ `json:"conf"`
}

// Index 实现 watch.Indexer 接口
func (m *MQDatasource) Index() string {
	return types.TextJoin(strconv.Itoa(int(m.TeamID)), ":", strconv.Itoa(int(m.ID)))
}

// String 实现 fmt.Stringer 接口
func (m *MQDatasource) String() string {
	bs, _ := types.Marshal(m)
	return string(bs)
}

// GetMQConfig 获取 MQ 配置
func (m *MQDatasource) GetMQConfig() *conf.MQ {
	if m == nil {
		return nil
	}
	return m.Conf
}
