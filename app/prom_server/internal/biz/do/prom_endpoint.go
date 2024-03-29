package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

const TableNameEndpoint = "endpoints"

const (
	EndpointFieldName     = "name"
	EndpointFieldEndpoint = "endpoint"
	EndpointFieldStatus   = "status"
	EndpointFieldRemark   = "remark"
)

type Endpoint struct {
	BaseModel
	Name     string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__ep__name,priority:1;comment:名称"`
	Endpoint string      `gorm:"column:endpoint;type:varchar(255);not null;uniqueIndex:idx__endpoint,priority:1;comment:地址"`
	Remark   string      `gorm:"column:remark;type:varchar(255);not null;default:这个Endpoint没有说明, 赶紧补充吧;comment:备注"`
	Status   vobj.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态: 1启用;2禁用"`
}

// TableName 表名
func (Endpoint) TableName() string {
	return TableNameEndpoint
}
