package auth

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (

	// FeiShuUser 飞书用户

	FeiShuUser struct {
		Name            string `json:"name"`
		EnName          string `json:"en_name"`
		AvatarURL       string `json:"avatar_url"`
		AvatarThumb     string `json:"avatar_thumb"`
		AvatarMiddle    string `json:"avatar_middle"`
		AvatarBig       string `json:"avatar_big"`
		OpenID          string `json:"open_id"`          // 用户在当前应用中的唯一标识
		UnionID         string `json:"union_id"`         // 用户在飞书开放平台中的唯一标识
		Email           string `json:"email"`            // 用户邮箱
		EnterpriseEmail string `json:"enterprise_email"` // 企业邮箱
		UserID          string `json:"user_id"`          // 用户ID（旧版字段）
		Mobile          string `json:"mobile"`           // 手机号（带国际码）
		TenantKey       string `json:"tenant_key"`       // 企业唯一标识
		EmployeeNo      string `json:"employee_no"`      // 工号
	}
)

func (f *FeiShuUser) GetOAuthID() string {
	return f.OpenID
}

func (f *FeiShuUser) GetEmail() string {
	return f.Email
}

func (f *FeiShuUser) GetRemark() string {
	return f.AvatarMiddle
}

func (f *FeiShuUser) GetUsername() string {
	return f.Name
}

func (f *FeiShuUser) GetNickname() string {
	return f.Name
}

func (f *FeiShuUser) GetAvatar() string {
	return f.AvatarURL
}

func (f *FeiShuUser) GetAPP() vobj.OAuthAPP {
	return vobj.OAuthFeiShu
}

func (f *FeiShuUser) String() string {
	bs, _ := types.Marshal(f)
	return string(bs)
}
