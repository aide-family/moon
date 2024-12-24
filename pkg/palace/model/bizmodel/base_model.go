package bizmodel

import (
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gorm"
)

// AllFieldModel 基础模型
type AllFieldModel struct {
	model.AllFieldModel
	// TeamID 团队ID
	TeamID uint32 `gorm:"column:team_id;type:int unsigned;not null;comment:团队ID" json:"team_id"`
}

// GetTeamID 获取团队ID
func (u *AllFieldModel) GetTeamID() uint32 {
	if types.IsNil(u) {
		return 0
	}
	if u.TeamID == 0 {
		u.TeamID = middleware.GetTeamID(u.GetContext())
	}
	return u.TeamID
}

// BeforeCreate 创建前的hook
func (u *AllFieldModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.GetContext() == nil {
		return
	}
	if err := u.AllFieldModel.BeforeCreate(tx); err != nil {
		return err
	}
	u.TeamID = middleware.GetTeamID(u.GetContext())
	return
}
