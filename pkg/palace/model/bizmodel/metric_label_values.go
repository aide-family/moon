package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"

	"gorm.io/plugin/soft_delete"
)

const tableNameMetricLabelValue = "metric_label_values"

// MetricLabelValue mapped from table <datasource_label_values>
type MetricLabelValue struct {
	model.AllFieldModel
	Name      string                `gorm:"column:name;type:varchar(255);not null;comment:值;uniqueIndex:idx__name__label_id__deleted_at,priority:3" json:"name"`               // 数据源名称
	LabelID   uint32                `gorm:"column:label_id;type:int unsigned;not null;comment:所属label;uniqueIndex:idx__name__label_id__deleted_at,priority:1" json:"label_id"` // 所属数据源
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间;uniqueIndex:idx__name__label_id__deleted_at,priority:2" json:"deleted_at"`      // 删除时间
}

// String json string
func (c *MetricLabelValue) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *MetricLabelValue) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *MetricLabelValue) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName MetricLabelValue's table name
func (*MetricLabelValue) TableName() string {
	return tableNameMetricLabelValue
}
