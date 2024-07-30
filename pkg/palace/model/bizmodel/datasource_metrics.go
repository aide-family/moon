package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/plugin/soft_delete"
)

const tableNameDatasourceMetric = "datasource_metrics"

// DatasourceMetric mapped from table <datasource_metrics>
type DatasourceMetric struct {
	model.AllFieldModel
	Name         string          `gorm:"column:name;type:varchar(255);not null;comment:指标名称;uniqueIndex:idx__name__datasource_id__deleted_at" json:"name"`
	Category     vobj.MetricType `gorm:"column:category;type:int;not null;comment:指标类型（对应prometheus四种数据类型）" json:"category"`
	Unit         string          `gorm:"column:unit;type:varchar(255);not null;comment:单位" json:"unit"`                                                                       // 单位
	Remark       string          `gorm:"column:remark;type:text;not null;comment:备注" json:"remark"`                                                                           // 备注
	DatasourceID uint32          `gorm:"column:datasource_id;type:int unsigned;not null;comment:所属数据源;uniqueIndex:idx__name__datasource_id__deleted_at" json:"datasource_id"` // 所属数据源
	// 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间;uniqueIndex:idx__name__datasource_id__deleted_at" json:"deleted_at"` // 删除时间
	Labels    []*MetricLabel        `gorm:"foreignKey:MetricID" json:"labels"`
}

// String json string
func (c *DatasourceMetric) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *DatasourceMetric) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *DatasourceMetric) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName DatasourceMetric's table name
func (*DatasourceMetric) TableName() string {
	return tableNameDatasourceMetric
}
