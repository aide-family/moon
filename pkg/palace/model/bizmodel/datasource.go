package bizmodel

import (
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameDatasource = "datasource"

// Datasource mapped from table <datasource>
type Datasource struct {
	AllFieldModel
	Name        string              `gorm:"column:name;type:varchar(64);not null;comment:数据源名称" json:"name"` // 数据源名称
	Category    vobj.DatasourceType `gorm:"column:category;type:int;not null;comment:数据源类型" json:"category"` // 数据源类型
	StorageType vobj.StorageType    `gorm:"column:storage_type;type:int;not null;comment:存储类型" json:"storage_type"`
	Config      *datasource.Config  `gorm:"column:config;type:JSON;not null;comment:数据源配置参数" json:"config"`           // 数据源配置参数
	Endpoint    string              `gorm:"column:endpoint;type:varchar(255);not null;comment:数据源地址" json:"endpoint"` // 数据源地址
	Status      vobj.Status         `gorm:"column:status;type:int;not null;comment:数据源状态" json:"status"`              // 数据源状态
	Remark      string              `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`      // 描述信息
}

// String json string
func (c *Datasource) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *Datasource) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *Datasource) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName Datasource's table name
func (*Datasource) TableName() string {
	return tableNameDatasource
}
