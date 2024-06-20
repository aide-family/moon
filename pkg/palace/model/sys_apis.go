package model

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/aide-family/moon/pkg/vobj"
)

const TableNameSysAPI = "sys_apis"

// SysAPI mapped from table <sys_apis>
type SysAPI struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间" json:"deleted_at"`
	Name      string                `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__sa__name,priority:1;comment:api名称" json:"name"`  // api名称
	Path      string                `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__sa__path,priority:1;comment:api路径" json:"path"` // api路径
	Status    vobj.Status           `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`                                          // 状态
	Remark    string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                     // 备注
	Module    int32                 `gorm:"column:module;type:int;not null;comment:模块" json:"module"`                                              // 模块
	Domain    int32                 `gorm:"column:domain;type:int;not null;comment:领域" json:"domain"`                                              // 领域
}

// String json string
func (c *SysAPI) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *SysAPI) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *SysAPI) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *SysAPI) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *SysAPI) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *SysAPI) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName SysAPI's table name
func (*SysAPI) TableName() string {
	return TableNameSysAPI
}
