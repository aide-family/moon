package do

import (
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/strategy"
)

const TableNameEndpoint = "endpoints"

const (
	EndpointFieldName           = "name"
	EndpointFieldEndpoint       = "endpoint"
	EndpointFieldStatus         = "status"
	EndpointFieldRemark         = "remark"
	EndpointFieldDatasourceType = "datasource_type"
)

type Endpoint struct {
	BaseModel
	Name           string                   `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__ep__name,priority:1;comment:名称"`
	Endpoint       string                   `gorm:"column:endpoint;type:varchar(255);not null;uniqueIndex:idx__endpoint,priority:1;comment:地址"`
	Remark         string                   `gorm:"column:remark;type:varchar(255);not null;default:这个Endpoint没有说明, 赶紧补充吧;comment:备注"`
	Status         vobj.Status              `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态: 1启用;2禁用"`
	BasicAuth      *strategy.BasicAuth      `gorm:"column:basic_auth;type:varchar(255);not null;default:'';comment:基础认证"`
	DatasourceType agent.DatasourceCategory `gorm:"column:datasource_type;type:tinyint;not null;default:0;comment:数据源类型: 0prometheus;1victoriametrics;2elasticsearch;3influxdb;4clickhouse;5loki"`
}

// TableName 表名
func (Endpoint) TableName() string {
	return TableNameEndpoint
}

// EndpointInDatasourceType 根据数据源类型搜索
func EndpointInDatasourceType(datasourceTypes ...agent.DatasourceCategory) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(EndpointFieldDatasourceType, datasourceTypes...)
}
