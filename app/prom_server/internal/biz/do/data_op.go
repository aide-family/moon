package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNameDataUserOp = "data_user_ops"
const TableNameDataRoleOp = "data_role_ops"

// DataUserOp 数据操作表
type DataUserOp struct {
	BaseModel
	Model   string `gorm:"column:model;type:varchar(255);not null;comment:模型;index:idx__du__model"`
	ModelId uint32 `gorm:"column:model_id;type:bigint(20);not null;comment:模型ID;idx__du__model_id"`
	UserId  uint32 `gorm:"column:user_id;type:bigint(20);not null;comment:用户ID;idx__du__user_id"`
	Op      vo.Op  `gorm:"column:op;type:varchar(255);not null;comment:操作;index:idx__du__op"`
}

// TableName 表名
func (*DataUserOp) TableName() string {
	return TableNameDataUserOp
}

// DataRoleOp 数据操作表
type DataRoleOp struct {
	BaseModel
	Model   string `gorm:"column:model;type:varchar(255);not null;comment:模型;index:idx__dr__model"`
	ModelId uint32 `gorm:"column:model_id;type:bigint(20);not null;comment:模型ID;idx__dr__model_id"`
	RoleId  uint32 `gorm:"column:role_id;type:bigint(20);not null;comment:角色ID;idx__dr__role_id"`
	Op      vo.Op  `gorm:"column:op;type:varchar(255);not null;comment:操作;index:idx__dr__op"`
}

// TableName 表名
func (*DataRoleOp) TableName() string {
	return TableNameDataRoleOp
}
