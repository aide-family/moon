package feishu

import (
	"encoding/json"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/oauth"
)

type User struct {
	Name            string `json:"name"`
	EnName          string `json:"en_name"`
	AvatarURL       string `json:"avatar_url"`
	AvatarThumb     string `json:"avatar_thumb"`
	AvatarMiddle    string `json:"avatar_middle"`
	AvatarBig       string `json:"avatar_big"`
	OpenID          string `json:"open_id"`
	UnionID         string `json:"union_id"`
	Email           string `json:"email"`
	EnterpriseEmail string `json:"enterprise_email"`
	ID              string `json:"user_id"`
	Mobile          string `json:"mobile"`
	TenantKey       string `json:"tenant_key"`
	EmployeeNo      string `json:"employee_no"`
}

func (u *User) Parse() *oauth.OAuth2User {
	return &oauth.OAuth2User{
		OpenID:   u.OpenID,
		Name:     u.Name,
		Nickname: u.EnName,
		Email:    u.Email,
		Avatar:   u.AvatarURL,
		App:      config.OAuth2_FEISHU,
		Remark:   "",
		Raw:      u.GetRaw(),
	}
}

// GetRaw implements [auth.User].
func (u *User) GetRaw() []byte {
	raw, _ := json.Marshal(u)
	return raw
}
