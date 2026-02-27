package bo

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/oauth"
)

type OAuth2UserBo struct {
	OpenID   string
	Name     string
	Nickname string
	Email    string
	Avatar   string
	App      config.OAuth2_APP
	Raw      []byte
	Remark   string
}

func NewOAuth2UserBo(user *oauth.OAuth2User) *OAuth2UserBo {
	return &OAuth2UserBo{
		OpenID:   user.OpenID,
		Name:     user.Name,
		Nickname: user.Nickname,
		Email:    user.Email,
		Avatar:   user.Avatar,
		App:      user.App,
		Raw:      user.Raw,
		Remark:   user.Remark,
	}
}

type OAuth2LoginBo struct {
	APP         config.OAuth2_APP
	Config      *config.OAuth2_Config
	User        *OAuth2UserBo
	RedirectURL string
}

func NewOAuth2LoginBo(req *oauth.OAuth2LoginRequest) *OAuth2LoginBo {
	return &OAuth2LoginBo{
		APP:         req.App,
		Config:      req.Config,
		User:        NewOAuth2UserBo(req.User),
		RedirectURL: req.Portal,
	}
}
