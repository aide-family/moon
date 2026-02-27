package convert

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

func OAuth2UserToBo(oauth2User *do.OAuth2User) *bo.OAuth2UserBo {
	if oauth2User == nil {
		return nil
	}
	return &bo.OAuth2UserBo{
		OpenID:   oauth2User.OpenID,
		Name:     oauth2User.Name,
		Nickname: oauth2User.Nickname,
		Email:    oauth2User.Email,
		Avatar:   oauth2User.Avatar,
		Remark:   oauth2User.Remark,
		App:      config.OAuth2_APP(config.OAuth2_APP_value[oauth2User.APP]),
		Raw:      oauth2User.Raw,
	}
}

func OAuth2UserToDo(oauth2User *bo.OAuth2UserBo) *do.OAuth2User {
	if oauth2User == nil {
		return nil
	}

	return &do.OAuth2User{
		OpenID:   oauth2User.OpenID,
		Name:     oauth2User.Name,
		Email:    oauth2User.Email,
		Avatar:   oauth2User.Avatar,
		APP:      oauth2User.App.String(),
		Raw:      oauth2User.Raw,
		Remark:   oauth2User.Remark,
		Nickname: oauth2User.Nickname,
	}
}

func OAuth2UserToUserDo(oauth2User *do.OAuth2User) *do.User {
	if oauth2User == nil {
		return nil
	}
	return &do.User{
		Email:    oauth2User.Email,
		Name:     oauth2User.Name,
		Avatar:   oauth2User.Avatar,
		Remark:   oauth2User.Remark,
		Nickname: oauth2User.Nickname,
		Status:   enum.UserStatus_ACTIVE,
	}
}
