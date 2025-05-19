package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToUser(ctx context.Context, userDo do.User) *system.User {
	if validate.IsNil(userDo) {
		return nil
	}
	user, ok := userDo.(*system.User)
	if ok {
		user.WithContext(ctx)
		return user
	}
	user = &system.User{
		BaseModel: ToBaseModel(ctx, userDo),
		Username:  userDo.GetUsername(),
		Nickname:  userDo.GetNickname(),
		Email:     userDo.GetEmail(),
		Phone:     userDo.GetPhone(),
		Remark:    userDo.GetRemark(),
		Avatar:    userDo.GetAvatar(),
		Gender:    userDo.GetGender(),
		Position:  userDo.GetPosition(),
		Status:    userDo.GetStatus(),
		Roles:     ToRoles(ctx, userDo.GetRoles()),
		Teams:     ToTeams(ctx, userDo.GetTeams()),
	}
	user.WithContext(ctx)
	return user
}

func ToUsers(ctx context.Context, userDos []do.User) []*system.User {
	return slices.MapFilter(userDos, func(userDo do.User) (*system.User, bool) {
		if validate.IsNil(userDo) {
			return nil, false
		}
		return ToUser(ctx, userDo), true
	})
}
