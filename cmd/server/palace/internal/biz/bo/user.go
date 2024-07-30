package bo

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// CreateUserParams 创建用户参数
	CreateUserParams struct {
		Name     string         `json:"name"`
		Password types.Password `json:"password"`
		Email    string         `json:"email"`
		Phone    string         `json:"phone"`
		Nickname string         `json:"nickname"`
		Remark   string         `json:"remark"`
		Avatar   string         `json:"avatar"`
		// 创建人
		CreatorID uint32 `json:"creatorID"`

		Status vobj.Status `json:"status"`
		Gender vobj.Gender `json:"gender"`
		Role   vobj.Role   `json:"role"`
	}

	// UpdateUserParams 更新用户参数
	UpdateUserParams struct {
		ID uint32 `json:"id"`
		CreateUserParams
	}

	// UpdateUserBaseParams 更新用户基础信息参数
	UpdateUserBaseParams struct {
		ID       uint32      `json:"id"`
		Nickname string      `json:"nickname"`
		Remark   string      `json:"remark"`
		Gender   vobj.Gender `json:"gender"`
	}

	// QueryUserSelectParams 查询用户选择参数
	QueryUserSelectParams struct {
		Keyword string           `json:"keyword"`
		Page    types.Pagination `json:"page"`
		Status  vobj.Status      `json:"status"`
		Gender  vobj.Gender      `json:"gender"`
		Role    vobj.Role        `json:"role"`
	}

	// QueryUserListParams 查询用户列表参数
	QueryUserListParams struct {
		Keyword string           `json:"keyword"`
		Page    types.Pagination `json:"page"`
		Status  vobj.Status      `json:"status"`
		Gender  vobj.Gender      `json:"gender"`
		Role    vobj.Role        `json:"role"`
	}

	// BatchUpdateUserStatusParams 批量更新用户状态参数
	BatchUpdateUserStatusParams struct {
		Status vobj.Status `json:"status"`
		IDs    []uint32    `json:"ids"`
	}

	// ResetUserPasswordBySelfParams 重置用户密码参数
	ResetUserPasswordBySelfParams struct {
		UserID   uint32         `json:"userId"`
		Password types.Password `json:"password"`
	}

	// UpdateUserPhoneRequest 更新用户手机号参数
	UpdateUserPhoneRequest struct {
		UserID uint32 `json:"userId"`
		Phone  string `json:"phone"`
	}

	// UpdateUserEmailRequest 更新用户邮箱参数
	UpdateUserEmailRequest struct {
		UserID uint32 `json:"userId"`
		Email  string `json:"email"`
	}

	// UpdateUserAvatarRequest 更新用户头像参数
	UpdateUserAvatarRequest struct {
		UserID uint32 `json:"userId"`
		Avatar string `json:"avatar"`
	}

	// UserSelectOptionBuild 用户选择项构建器
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
