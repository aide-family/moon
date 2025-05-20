package system

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/validate"
)

var _ do.UserOAuth = (*UserOAuth)(nil)

const tableNameOAuthUser = "sys_oauth_users"

type UserOAuth struct {
	do.BaseModel
	OpenID string        `gorm:"column:open_id;type:varchar(255);not null;comment:open_id;index:uk__open_id__app,unique" json:"open_id"`
	UserID uint32        `gorm:"column:user_id;type:int unsigned;not null;comment:associated user ID;index:uk__open_id__app,unique" json:"user_id"`
	Row    string        `gorm:"column:row;type:text;comment:user information" json:"row"`
	APP    vobj.OAuthAPP `gorm:"column:app;type:tinyint;not null;comment:OAuth application;index:uk__open_id__app,unique" json:"app"`
	User   *User         `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (s *UserOAuth) TableName() string {
	return tableNameOAuthUser
}

func (s *UserOAuth) GetOpenID() string {
	return s.OpenID
}

func (s *UserOAuth) GetAPP() vobj.OAuthAPP {
	return s.APP
}

func (s *UserOAuth) GetUserID() uint32 {
	return s.UserID
}

func (s *UserOAuth) GetRow() string {
	return s.Row
}

func (s *UserOAuth) GetUser() do.User {
	return s.User
}

func (s *UserOAuth) SetUser(user do.User) {
	if validate.IsNil(user) {
		return
	}
	userDo, ok := user.(*User)
	s.UserID = user.GetID()
	if !ok {
		return
	}
	s.User = userDo
}
