package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	userapi "github.com/aide-family/moon/api/admin/user"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IUserModuleBuilder = (*userModuleBuilder)(nil)

type (
	userModuleBuilder struct {
		ctx context.Context
	}

	IUserModuleBuilder interface {
		WithCreateUserRequest(*userapi.CreateUserRequest) ICreateUserRequestBuilder
		WithUpdateUserRequest(*userapi.UpdateUserRequest) IUpdateUserRequestBuilder
		WithListUserRequest(*userapi.ListUserRequest) IListUserRequestBuilder
		WithBatchUpdateUserStatusRequest(*userapi.BatchUpdateUserStatusRequest) IBatchUpdateUserStatusRequestBuilder
		WithResetUserPasswordBySelfRequest(*userapi.ResetUserPasswordBySelfRequest) IResetUserPasswordBySelfRequestBuilder
		WithUpdateUserPhoneRequest(*userapi.UpdateUserPhoneRequest) IUpdateUserPhoneRequestBuilder
		WithUpdateUserEmailRequest(*userapi.UpdateUserEmailRequest) IUpdateUserEmailRequestBuilder
		WithUpdateUserAvatarRequest(*userapi.UpdateUserAvatarRequest) IUpdateUserAvatarRequestBuilder
		WithUpdateUserBaseInfoRequest(*userapi.UpdateUserBaseInfoRequest) IUpdateUserBaseInfoRequestBuilder
		DoUserBuilder() IDoUserBuilder
		DoNoticeUserBuilder() INoticeUserBuilder
	}

	ICreateUserRequestBuilder interface {
		ToBo() *bo.CreateUserParams
	}

	createUserRequestBuilder struct {
		ctx context.Context
		*userapi.CreateUserRequest
	}

	IUpdateUserRequestBuilder interface {
		ToBo() *bo.UpdateUserParams
	}

	updateUserRequestBuilder struct {
		ctx context.Context
		*userapi.UpdateUserRequest
	}

	IListUserRequestBuilder interface {
		ToBo() *bo.QueryUserListParams
	}

	listUserRequestBuilder struct {
		ctx context.Context
		*userapi.ListUserRequest
	}

	IBatchUpdateUserStatusRequestBuilder interface {
		ToBo() *bo.BatchUpdateUserStatusParams
	}

	batchUpdateUserStatusRequestBuilder struct {
		ctx context.Context
		*userapi.BatchUpdateUserStatusRequest
	}

	IResetUserPasswordBySelfRequestBuilder interface {
		ToBo() *bo.ResetUserPasswordBySelfParams
		WithUserInfo(f func(ctx context.Context, id uint32) (*model.SysUser, error)) (IResetUserPasswordBySelfRequestBuilder, error)
	}

	resetUserPasswordBySelfRequestBuilder struct {
		ctx    context.Context
		userDo *model.SysUser
		*userapi.ResetUserPasswordBySelfRequest
	}

	IUpdateUserPhoneRequestBuilder interface {
		ToBo() *bo.UpdateUserPhoneRequest
	}

	updateUserPhoneRequestBuilder struct {
		ctx context.Context
		*userapi.UpdateUserPhoneRequest
	}

	IUpdateUserEmailRequestBuilder interface {
		ToBo() *bo.UpdateUserEmailRequest
	}

	updateUserEmailRequestBuilder struct {
		ctx context.Context
		*userapi.UpdateUserEmailRequest
	}

	IUpdateUserAvatarRequestBuilder interface {
		ToBo() *bo.UpdateUserAvatarRequest
	}

	updateUserAvatarRequestBuilder struct {
		ctx context.Context
		*userapi.UpdateUserAvatarRequest
	}

	IUpdateUserBaseInfoRequestBuilder interface {
		ToBo() *bo.UpdateUserBaseParams
	}

	updateUserBaseInfoRequestBuilder struct {
		ctx context.Context
		*userapi.UpdateUserBaseInfoRequest
	}

	IDoUserBuilder interface {
		ToAPI(*model.SysUser) *adminapi.UserItem
		ToAPIs([]*model.SysUser) []*adminapi.UserItem
		ToSelect(*model.SysUser) *adminapi.SelectItem
		ToSelects([]*model.SysUser) []*adminapi.SelectItem
	}

	doUserBuilder struct {
		ctx context.Context
	}

	INoticeUserBuilder interface {
		ToAPI(*bizmodel.AlarmNoticeMember) *adminapi.NoticeItem
		ToAPIs([]*bizmodel.AlarmNoticeMember) []*adminapi.NoticeItem
	}

	doNoticeUserBuilder struct {
		ctx context.Context
	}
)

func (d *doNoticeUserBuilder) ToAPI(user *bizmodel.AlarmNoticeMember) *adminapi.NoticeItem {
	if types.IsNil(user) {
		return nil
	}

	return &adminapi.NoticeItem{
		User:       nil, // TODO impl
		NotifyType: api.NotifyType(user.AlarmNoticeType),
	}
}

func (d *doNoticeUserBuilder) ToAPIs(users []*bizmodel.AlarmNoticeMember) []*adminapi.NoticeItem {
	if types.IsNil(users) {
		return nil
	}

	return types.SliceTo(users, func(user *bizmodel.AlarmNoticeMember) *adminapi.NoticeItem {
		return d.ToAPI(user)
	})
}

func (u *userModuleBuilder) DoNoticeUserBuilder() INoticeUserBuilder {
	return &doNoticeUserBuilder{ctx: u.ctx}
}

func (u *updateUserBaseInfoRequestBuilder) ToBo() *bo.UpdateUserBaseParams {
	claims, ok := middleware.ParseJwtClaims(u.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(u.ctx))
	}
	return &bo.UpdateUserBaseParams{
		ID:       claims.GetUser(),
		Gender:   vobj.Gender(u.GetGender()),
		Remark:   u.GetRemark(),
		Nickname: u.GetNickname(),
	}
}

func (u *userModuleBuilder) WithUpdateUserBaseInfoRequest(request *userapi.UpdateUserBaseInfoRequest) IUpdateUserBaseInfoRequestBuilder {
	return &updateUserBaseInfoRequestBuilder{ctx: u.ctx, UpdateUserBaseInfoRequest: request}
}

func (u *updateUserAvatarRequestBuilder) ToBo() *bo.UpdateUserAvatarRequest {
	if types.IsNil(u) || types.IsNil(u.UpdateUserAvatarRequest) {
		return nil
	}
	claims, ok := middleware.ParseJwtClaims(u.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(u.ctx))
	}

	return &bo.UpdateUserAvatarRequest{
		UserID: claims.GetUser(),
		Avatar: u.GetAvatar(),
	}
}

func (u *userModuleBuilder) WithUpdateUserAvatarRequest(request *userapi.UpdateUserAvatarRequest) IUpdateUserAvatarRequestBuilder {
	return &updateUserAvatarRequestBuilder{ctx: u.ctx, UpdateUserAvatarRequest: request}
}

func (u *updateUserEmailRequestBuilder) ToBo() *bo.UpdateUserEmailRequest {
	if types.IsNil(u) || types.IsNil(u.UpdateUserEmailRequest) {
		return nil
	}
	claims, ok := middleware.ParseJwtClaims(u.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(u.ctx))
	}
	return &bo.UpdateUserEmailRequest{
		UserID: claims.GetUser(),
		Email:  u.GetEmail(),
	}
}

func (u *userModuleBuilder) WithUpdateUserEmailRequest(request *userapi.UpdateUserEmailRequest) IUpdateUserEmailRequestBuilder {
	return &updateUserEmailRequestBuilder{ctx: u.ctx, UpdateUserEmailRequest: request}
}

func (u *updateUserPhoneRequestBuilder) ToBo() *bo.UpdateUserPhoneRequest {
	if types.IsNil(u) || types.IsNil(u.UpdateUserPhoneRequest) {
		return nil
	}

	claims, ok := middleware.ParseJwtClaims(u.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(u.ctx))
	}
	return &bo.UpdateUserPhoneRequest{
		UserID: claims.GetUser(),
		Phone:  u.GetPhone(),
	}
}

func (u *userModuleBuilder) WithUpdateUserPhoneRequest(request *userapi.UpdateUserPhoneRequest) IUpdateUserPhoneRequestBuilder {
	return &updateUserPhoneRequestBuilder{ctx: u.ctx, UpdateUserPhoneRequest: request}
}

func (r *resetUserPasswordBySelfRequestBuilder) WithUserInfo(f func(ctx context.Context, id uint32) (*model.SysUser, error)) (IResetUserPasswordBySelfRequestBuilder, error) {
	claims, ok := middleware.ParseJwtClaims(r.ctx)
	if !ok {
		return nil, merr.ErrorI18nUnauthorized(r.ctx)
	}
	// 查询用户详情
	userDo, err := f(r.ctx, claims.GetUser())
	if !types.IsNil(err) {
		return nil, err
	}
	newPass := types.NewPassword(r.GetNewPassword(), userDo.Salt)
	oldPass := userDo.Password
	// 对比旧密码正确
	if oldPass != r.GetOldPassword() {
		return nil, merr.ErrorI18nAlertPasswordErr(r.ctx)
	}

	// 对比两次密码相同, 相同修改无意义
	if newPass.String() == oldPass {
		return nil, merr.ErrorI18nAlertPasswordSameErr(r.ctx)
	}
	return r, nil
}

func (r *resetUserPasswordBySelfRequestBuilder) ToBo() *bo.ResetUserPasswordBySelfParams {
	if types.IsNil(r) || types.IsNil(r.ResetUserPasswordBySelfRequest) || types.IsNil(r.userDo) {
		return nil
	}

	return &bo.ResetUserPasswordBySelfParams{
		UserID: r.userDo.ID,
		// 使用新的盐
		Password: types.NewPassword(r.GetNewPassword()),
	}
}

func (b *batchUpdateUserStatusRequestBuilder) ToBo() *bo.BatchUpdateUserStatusParams {
	if types.IsNil(b) || types.IsNil(b.BatchUpdateUserStatusRequest) {
		return nil
	}

	return &bo.BatchUpdateUserStatusParams{
		Status: vobj.Status(b.GetStatus()),
		IDs:    b.GetIds(),
	}
}

func (l *listUserRequestBuilder) ToBo() *bo.QueryUserListParams {
	if types.IsNil(l) || types.IsNil(l.ListUserRequest) {
		return nil
	}

	return &bo.QueryUserListParams{
		Keyword: l.GetKeyword(),
		Page:    types.NewPagination(l.GetPagination()),
		Status:  vobj.Status(l.GetStatus()),
		Gender:  vobj.Gender(l.GetGender()),
		Role:    vobj.Role(l.GetRole()),
		IDs:     l.GetIds(),
	}
}

func (u *updateUserRequestBuilder) ToBo() *bo.UpdateUserParams {
	if types.IsNil(u) || types.IsNil(u.UpdateUserRequest) {
		return nil
	}
	creatBuilder := &createUserRequestBuilder{ctx: u.ctx, CreateUserRequest: u.GetData()}
	createParams := creatBuilder.ToBo()
	return &bo.UpdateUserParams{
		ID:               u.GetId(),
		CreateUserParams: createParams,
	}
}

func (c *createUserRequestBuilder) ToBo() *bo.CreateUserParams {
	if types.IsNil(c) || types.IsNil(c.CreateUserRequest) {
		return nil
	}
	pass := types.NewPassword(c.GetPassword())
	claims, ok := middleware.ParseJwtClaims(c.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(c.ctx))
	}
	return &bo.CreateUserParams{
		Name:      c.GetName(),
		Password:  pass,
		Email:     c.GetEmail(),
		Phone:     c.GetPhone(),
		Nickname:  c.GetNickname(),
		Remark:    c.GetRemark(),
		Avatar:    c.GetAvatar(),
		CreatorID: claims.GetUser(),
		Status:    vobj.Status(c.GetStatus()),
		Gender:    vobj.Gender(c.GetGender()),
		Role:      vobj.Role(c.GetRole()),
	}
}

func (u *userModuleBuilder) WithCreateUserRequest(request *userapi.CreateUserRequest) ICreateUserRequestBuilder {
	return &createUserRequestBuilder{ctx: u.ctx, CreateUserRequest: request}
}

func (u *userModuleBuilder) WithUpdateUserRequest(request *userapi.UpdateUserRequest) IUpdateUserRequestBuilder {
	return &updateUserRequestBuilder{ctx: u.ctx, UpdateUserRequest: request}
}

func (u *userModuleBuilder) WithListUserRequest(request *userapi.ListUserRequest) IListUserRequestBuilder {
	return &listUserRequestBuilder{ctx: u.ctx, ListUserRequest: request}
}

func (u *userModuleBuilder) WithBatchUpdateUserStatusRequest(request *userapi.BatchUpdateUserStatusRequest) IBatchUpdateUserStatusRequestBuilder {
	return &batchUpdateUserStatusRequestBuilder{ctx: u.ctx, BatchUpdateUserStatusRequest: request}
}

func (u *userModuleBuilder) WithResetUserPasswordBySelfRequest(request *userapi.ResetUserPasswordBySelfRequest) IResetUserPasswordBySelfRequestBuilder {
	return &resetUserPasswordBySelfRequestBuilder{ctx: u.ctx, ResetUserPasswordBySelfRequest: request}
}

func (d *doUserBuilder) ToAPI(user *model.SysUser) *adminapi.UserItem {
	if types.IsNil(user) || types.IsNil(d) {
		return nil
	}

	return &adminapi.UserItem{
		Id:        user.ID,
		Name:      user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Phone:     user.Phone,
		Status:    api.Status(user.Status),
		Gender:    api.Gender(user.Gender),
		Role:      api.Role(user.Role),
		Avatar:    user.Avatar,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	}
}

func (d *doUserBuilder) ToAPIs(users []*model.SysUser) []*adminapi.UserItem {
	if types.IsNil(users) || types.IsNil(d) {
		return nil
	}

	return types.SliceTo(users, func(item *model.SysUser) *adminapi.UserItem {
		return d.ToAPI(item)
	})
}

func (d *doUserBuilder) ToSelect(user *model.SysUser) *adminapi.SelectItem {
	if types.IsNil(user) || types.IsNil(d) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    user.ID,
		Label:    user.Nickname,
		Disabled: user.DeletedAt > 0 || !user.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: user.Remark,
			Image:  user.Avatar,
		},
	}
}

func (d *doUserBuilder) ToSelects(users []*model.SysUser) []*adminapi.SelectItem {
	if types.IsNil(users) || types.IsNil(d) {
		return nil
	}

	return types.SliceTo(users, func(item *model.SysUser) *adminapi.SelectItem {
		return d.ToSelect(item)
	})
}

func (u *userModuleBuilder) DoUserBuilder() IDoUserBuilder {
	return &doUserBuilder{ctx: u.ctx}
}
