package bo

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	CreateUserParams struct {
		Name     string         `json:"name"`
		Password types.Password `json:"password"`
		Email    string         `json:"email"`
		Phone    string         `json:"phone"`

		Nickname string `json:"nickname"`
		Remark   string `json:"remark"`
		Avatar   string `json:"avatar"`
		// 创建人
		CreatorID uint32 `json:"creatorID"`

		Status vobj.Status `json:"status"`
		Gender vobj.Gender `json:"gender"`
		Role   vobj.Role   `json:"role"`
	}

	UpdateUserParams struct {
		ID uint32 `json:"id"`
		CreateUserParams
	}

	QueryUserSelectParams struct {
		Keyword string           `json:"keyword"`
		Page    types.Pagination `json:"page"`
		Status  vobj.Status      `json:"status"`
		Gender  vobj.Gender      `json:"gender"`
		Role    vobj.Role        `json:"role"`
	}

	QueryUserListParams struct {
		Keyword string           `json:"keyword"`
		Page    types.Pagination `json:"page"`
		Status  vobj.Status      `json:"status"`
		Gender  vobj.Gender      `json:"gender"`
		Role    vobj.Role        `json:"role"`
	}

	BatchUpdateUserStatusParams struct {
		Status vobj.Status `json:"status"`
		IDs    []uint32    `json:"ids"`
	}

	ResetUserPasswordBySelfParams struct {
		UserId   uint32         `json:"userId"`
		Password types.Password `json:"password"`
	}

	UpdateUserPhoneRequest struct {
		UserId uint32 `json:"userId"`
		Phone  string `json:"phone"`
	}

	UpdateUserEmailRequest struct {
		UserId uint32 `json:"userId"`
		Email  string `json:"email"`
	}

	UpdateUserAvatarRequest struct {
		UserId uint32 `json:"userId"`
		Avatar string `json:"avatar"`
	}

	UserSelectOptionBuild struct {
		*model.SysUser
	}
)

// NewUserSelectOptionBuild 创建选择项构建器
func NewUserSelectOptionBuild(user *model.SysUser) *UserSelectOptionBuild {
	return &UserSelectOptionBuild{
		SysUser: user,
	}
}

// ToSelectOption 转换为选择项
func (u *UserSelectOptionBuild) ToSelectOption() *SelectOptionBo {
	if types.IsNil(u) || types.IsNil(u.SysUser) {
		return nil
	}
	return &SelectOptionBo{
		Value:    u.ID,
		Label:    u.Username,
		Disabled: u.DeletedAt > 0 || !u.Status.IsEnable(),
	}
}
