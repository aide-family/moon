package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/plugin/soft_delete"
)

const tableNameMetricLabel = "metric_labels"

// MetricLabel mapped from table <metric_labels>
type MetricLabel struct {
	AllFieldModel
	Name        string                `gorm:"column:name;type:varchar(255);not null;comment:标签名称名称;uniqueIndex:idx__name__metric_id__deleted_at" json:"name"`
	MetricID    uint32                `gorm:"column:metric_id;type:int unsigned;not null;comment:所属指标;uniqueIndex:idx__name__metric_id__deleted_at" json:"metric_id"` // 所属指标
	DeletedAt   soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间;uniqueIndex:idx__name__metric_id__deleted_at" json:"deleted_at"`     // 删除时间
	Remark      string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                                      // 备注
	LabelValues string                `gorm:"column:label_values;type:text;not null;comment:标签值" json:"label_values"`
}

// GetLabelValues get label values
func (c *MetricLabel) GetLabelValues() []string {
	var values []string
	_ = types.Unmarshal([]byte(c.LabelValues), &values)
	return values
}

// String json string
func (c *MetricLabel) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *MetricLabel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *MetricLabel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName MetricLabel's table name
func (*MetricLabel) TableName() string {
	return tableNameMetricLabel
}
