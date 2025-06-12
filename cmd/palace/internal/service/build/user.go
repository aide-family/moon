package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/strutil"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

// ToUserItem converts a system.User to a common.UserItem
func ToUserItem(user do.User) *common.UserItem {
	if validate.IsNil(user) {
		return nil
	}

	return &common.UserItem{
		Username:  user.GetUsername(),
		Nickname:  user.GetNickname(),
		Avatar:    user.GetAvatar(),
		Gender:    common.Gender(user.GetGender().GetValue()),
		Email:     strutil.MaskEmail(string(user.GetEmail())),
		Phone:     strutil.MaskPhone(string(user.GetPhone())),
		Remark:    user.GetRemark(),
		Position:  common.UserPosition(user.GetPosition().GetValue()),
		Status:    common.UserStatus(user.GetStatus().GetValue()),
		CreatedAt: timex.Format(user.GetCreatedAt()),
		UpdatedAt: timex.Format(user.GetUpdatedAt()),
		UserId:    user.GetID(),
	}
}

func ToUserItemPlaintext(user do.User) *common.UserItem {
	if validate.IsNil(user) {
		return nil
	}

	return &common.UserItem{
		Username:  user.GetUsername(),
		Nickname:  user.GetNickname(),
		Avatar:    user.GetAvatar(),
		Gender:    common.Gender(user.GetGender().GetValue()),
		Email:     string(user.GetEmail()),
		Phone:     string(user.GetPhone()),
		Remark:    user.GetRemark(),
		Position:  common.UserPosition(user.GetPosition().GetValue()),
		Status:    common.UserStatus(user.GetStatus().GetValue()),
		CreatedAt: timex.Format(user.GetCreatedAt()),
		UpdatedAt: timex.Format(user.GetUpdatedAt()),
		UserId:    user.GetID(),
	}
}

func ToUserBaseItem(user do.User) *common.UserBaseItem {
	if validate.IsNil(user) {
		return nil
	}

	return &common.UserBaseItem{
		Username: user.GetUsername(),
		Nickname: user.GetNickname(),
		Avatar:   user.GetAvatar(),
		Gender:   common.Gender(user.GetGender().GetValue()),
		UserId:   user.GetID(),
	}
}

// ToUserItems converts a slice of system.User to a slice of common.UserItem
func ToUserItems(users []do.User) []*common.UserItem {
	return slices.Map(users, ToUserItem)
}

// ToUserBaseItems converts a slice of system.User to a slice of common.UserBaseItem
func ToUserBaseItems(users []do.User) []*common.UserBaseItem {
	return slices.Map(users, ToUserBaseItem)
}

func ToSelfUpdateInfo(req *palace.UpdateSelfInfoRequest) *bo.UserUpdateInfo {
	if req == nil {
		panic("UpdateSelfInfoRequest is nil")
	}

	return &bo.UserUpdateInfo{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   vobj.Gender(req.Gender),
	}
}

func ToUserUpdateInfo(req *palace.UpdateUserRequest) *bo.UserUpdateInfo {
	if req == nil {
		panic("UpdateUserRequest is nil")
	}

	return &bo.UserUpdateInfo{
		UserID:   req.GetUserId(),
		Nickname: req.GetNickname(),
		Avatar:   req.GetAvatar(),
		Gender:   vobj.Gender(req.GetGender()),
	}
}

// ToPasswordUpdateInfo converts an API password update request to a business object
func ToPasswordUpdateInfo(req *palace.UpdateSelfPasswordRequest, sendEmailFun bo.SendEmailFun) *bo.PasswordUpdateInfo {
	if req == nil {
		panic("UpdateSelfPasswordRequest is nil")
	}

	return &bo.PasswordUpdateInfo{
		OldPassword:  req.OldPassword,
		NewPassword:  req.NewPassword,
		SendEmailFun: sendEmailFun,
	}
}

func ToUserListRequest(req *palace.GetUserListRequest) *bo.UserListRequest {
	if validate.IsNil(req) {
		panic("GetUserListRequest is nil")
	}
	return &bo.UserListRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Status: slices.MapFilter(req.GetStatus(), func(status common.UserStatus) (vobj.UserStatus, bool) {
			vobjStatus := vobj.UserStatus(status)
			return vobjStatus, vobjStatus.Exist() && !vobjStatus.IsUnknown()
		}),
		Position: slices.MapFilter(req.GetPosition(), func(position common.UserPosition) (vobj.Position, bool) {
			vobjPosition := vobj.Position(position)
			return vobjPosition, vobjPosition.Exist() && !vobjPosition.IsUnknown()
		}),
		Keyword: req.GetKeyword(),
	}
}
