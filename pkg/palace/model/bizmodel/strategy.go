package bizmodel

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"google.golang.org/protobuf/types/known/durationpb"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const TableNameStrategy = "strategy"

// Strategy mapped from table <Strategy>
type Strategy struct {
	ID          uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt   types.Time            `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt   types.Time            `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt   soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间" json:"deleted_at"`
	Alert       string                `gorm:"column:alert;type:varchar(64);not null;comment:策略名称" json:"alert"`
	Expr        string                `gorm:"column:expr;type:text;not null;comment:告警表达式" json:"expr"`
	For         durationpb.Duration   `gorm:"column:for;type:varchar(64);not null;comment:告警持续时间" json:"for"`
	Count       uint32                `gorm:"column:count;type:int unsigned;not null;comment:持续次数" json:"count"`
	SustainType vobj.Sustain          `gorm:"column:sustain_type;type:int(11);not null;comment:持续类型" json:"sustain_type"`
	Labels      map[string]string     `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	Annotations map[string]string     `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	Interval    durationpb.Duration   `gorm:"column:interval;type:varchar(64);not null;comment:执行频率" json:"interval"`
	Remark      string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	CreatorID   uint32                `gorm:"column:creator;type:int unsigned;not null;comment:创建者" json:"creator_id"`
	Status      vobj.Status           `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`

	Datasource []*Datasource  `gorm:"many2many:strategy_datasource;" json:"datasource"`
	Creator    *SysTeamMember `gorm:"foreignKey:CreatorID" json:"creator"`
}

// String json string
func (c *Strategy) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *Strategy) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *Strategy) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *Strategy) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *Strategy) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *Strategy) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName Strategy's table name
func (*Strategy) TableName() string {
	return TableNameStrategy
}
