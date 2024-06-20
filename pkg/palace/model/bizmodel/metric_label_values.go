package bizmodel

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const TableNameMetricLabelValue = "metric_label_values"

// MetricLabelValue mapped from table <datasource_label_values>
type MetricLabelValue struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name      string                `gorm:"column:name;type:varchar(255);not null;comment:值;uniqueIndex:idx__name__label_id__deleted_at,priority:3" json:"name"`               // 数据源名称
	LabelID   uint32                `gorm:"column:label_id;type:int unsigned;not null;comment:所属label;uniqueIndex:idx__name__label_id__deleted_at,priority:1" json:"label_id"` // 所属数据源
	CreatedAt types.Time            `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`                                 // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`                                 // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间;uniqueIndex:idx__name__label_id__deleted_at,priority:2" json:"deleted_at"`      // 删除时间
}

// String json string
func (c *MetricLabelValue) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *MetricLabelValue) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *MetricLabelValue) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *MetricLabelValue) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *MetricLabelValue) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *MetricLabelValue) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName MetricLabelValue's table name
func (*MetricLabelValue) TableName() string {
	return TableNameMetricLabelValue
}
