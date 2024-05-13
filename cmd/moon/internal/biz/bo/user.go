package bo

import (
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/vobj"
	"github.com/aide-cloud/moon/pkg/notify/email"
	"github.com/aide-cloud/moon/pkg/notify/phone"
	"github.com/aide-cloud/moon/pkg/types"
)

type (
	CreateUserParams struct {
		Name     string         `json:"name"`
		Password types.Password `json:"password"`
		Email    email.Type     `json:"email"`
		Phone    phone.Type     `json:"phone"`

		Nickname string `json:"nickname"`
		Remark   string `json:"remark"`
		Avatar   string `json:"avatar"`
		// 创建人
		CreatorID int `json:"creatorID"`

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
	}

	QueryUserListParams struct {
		Keyword string           `json:"keyword"`
		Page    types.Pagination `json:"page"`
		Status  vobj.Status      `json:"status"`
		Gender  vobj.Gender      `json:"gender"`
	}
)
