package bizmodel

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const TableNameMetricLabel = "metric_labels"

// MetricLabel mapped from table <metric_labels>
type MetricLabel struct {
	ID          uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name        string                `gorm:"column:name;type:varchar(255);not null;comment:标签名称名称;uniqueIndex:idx__name__metric_id__deleted_at" json:"name"`
	MetricID    uint32                `gorm:"column:metric_id;type:int unsigned;not null;comment:所属指标;uniqueIndex:idx__name__metric_id__deleted_at" json:"metric_id"` // 所属指标
	CreatedAt   types.Time            `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`                      // 创建时间
	UpdatedAt   types.Time            `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`                      // 更新时间
	DeletedAt   soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间;uniqueIndex:idx__name__metric_id__deleted_at" json:"deleted_at"`     // 删除时间
	Remark      string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                                      // 备注
	LabelValues []*MetricLabelValue   `gorm:"foreignKey:LabelID" json:"label_values"`
}

// String json string
func (c *MetricLabel) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *MetricLabel) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *MetricLabel) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *MetricLabel) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *MetricLabel) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *MetricLabel) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName MetricLabel's table name
func (*MetricLabel) TableName() string {
	return TableNameMetricLabel
}
