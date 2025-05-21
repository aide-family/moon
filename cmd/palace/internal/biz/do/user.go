package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/cache"
)

type User interface {
	Base
	cache.Object
	GetUsername() string
	GetNickname() string
	GetEmail() string
	GetPhone() string
	GetRemark() string
	GetPassword() string
	GetSalt() string
	GetGender() vobj.Gender
	GetAvatar() string
	GetStatus() vobj.UserStatus
	GetPosition() vobj.Position
	GetRoles() []Role
	GetTeams() []Team
	ValidatePassword(p string) bool
	SetEmail(email string)
}

type UserOAuth interface {
	Base
	GetOpenID() string
	GetAPP() vobj.OAuthAPP
	GetUserID() uint32
	GetRow() string
	GetUser() User
	SetUser(user User)
}
