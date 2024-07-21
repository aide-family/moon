package user

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	userapi "github.com/aide-family/moon/api/admin/user"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/cipher"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type Service struct {
	userapi.UnimplementedUserServer

	userBiz *biz.UserBiz
}

func NewUserService(userBiz *biz.UserBiz) *Service {
	return &Service{
		userBiz: userBiz,
	}
}

const (
	defaultKey = "1234567890123456"
	defaultIv  = "1234567890123456"
)

// 解密传输密码字符串
func decryptPassword(ctx context.Context, password string) (string, error) {
	aes, err := cipher.NewAes(defaultKey, defaultIv)
	if !types.IsNil(err) {
		return "", merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	decryptBase64Pass, err := aes.DecryptBase64(password)
	if !types.IsNil(err) {
		return "", merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	pass := string(decryptBase64Pass)
	return pass, nil
}

// 加密传输密码字符串
func encryptPassword(ctx context.Context, password string) (string, error) {
	aes, err := cipher.NewAes(defaultKey, defaultIv)
	if !types.IsNil(err) {
		return "", merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	encryptBase64Pass, err := aes.EncryptBase64([]byte(password))
	if !types.IsNil(err) {
		return "", merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return encryptBase64Pass, nil
}

// CreateUser 创建用户 只允许管理员操作
func (s *Service) CreateUser(ctx context.Context, req *userapi.CreateUserRequest) (*userapi.CreateUserReply, error) {
	pass, err := decryptPassword(ctx, req.GetPassword())
	if !types.IsNil(err) {
		return nil, merr.ErrorAlert("请使用加密后的密文传输").WithMetadata(map[string]string{
			"password": "请使用加密密文",
		})
	}
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	createParams := build.NewBuilder().WithCreateUserBo(req).ToCreateUserBO(claims.GetUser(), pass)
	_, err = s.userBiz.CreateUser(ctx, createParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.CreateUserReply{}, nil
}

// UpdateUser 更新用户基础信息， 只允许管理员操作
func (s *Service) UpdateUser(ctx context.Context, req *userapi.UpdateUserRequest) (*userapi.UpdateUserReply, error) {
	updateParams := build.NewBuilder().WithUpdateUserBo(req).ToUpdateUserBO()
	if err := s.userBiz.UpdateUser(ctx, updateParams); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserReply{}, nil
}

func (s *Service) DeleteUser(ctx context.Context, req *userapi.DeleteUserRequest) (*userapi.DeleteUserReply, error) {
	if err := s.userBiz.DeleteUser(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.DeleteUserReply{}, nil
}

func (s *Service) GetUser(ctx context.Context, req *userapi.GetUserRequest) (*userapi.GetUserReply, error) {
	userDo, err := s.userBiz.GetUser(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.GetUserReply{
		User: build.NewBuilder().WithApiUserBo(userDo).ToApi(),
	}, nil
}

func (s *Service) ListUser(ctx context.Context, req *userapi.ListUserRequest) (*userapi.ListUserReply, error) {
	queryParams := &bo.QueryUserListParams{
		Keyword: req.GetKeyword(),
		Page:    types.NewPagination(req.GetPagination()),
		Status:  vobj.Status(req.GetStatus()),
		Gender:  vobj.Gender(req.GetGender()),
		Role:    vobj.Role(req.GetRole()),
	}
	userDos, err := s.userBiz.ListUser(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.ListUserReply{
		List: types.SliceTo(userDos, func(user *model.SysUser) *admin.User {
			return build.NewBuilder().WithApiUserBo(user).ToApi()
		}),
		Pagination: build.NewPageBuilder(queryParams.Page).ToApi(),
	}, nil
}

func (s *Service) BatchUpdateUserStatus(ctx context.Context, req *userapi.BatchUpdateUserStatusRequest) (*userapi.BatchUpdateUserStatusReply, error) {
	params := &bo.BatchUpdateUserStatusParams{
		Status: vobj.Status(req.GetStatus()),
		IDs:    req.GetIds(),
	}
	if err := s.userBiz.BatchUpdateUserStatus(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.BatchUpdateUserStatusReply{}, nil
}

func (s *Service) ResetUserPassword(ctx context.Context, req *userapi.ResetUserPasswordRequest) (*userapi.ResetUserPasswordReply, error) {
	return &userapi.ResetUserPasswordReply{}, nil
}

func (s *Service) ResetUserPasswordBySelf(ctx context.Context, req *userapi.ResetUserPasswordBySelfRequest) (*userapi.ResetUserPasswordBySelfReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	// 查询用户详情
	userDo, err := s.userBiz.GetUser(ctx, claims.GetUser())
	if !types.IsNil(err) {
		return nil, err
	}
	newPass, err := decryptPassword(ctx, req.GetNewPassword())
	if !types.IsNil(err) {
		return nil, err
	}
	oldPass, err := decryptPassword(ctx, req.GetOldPassword())
	if !types.IsNil(err) {
		return nil, err
	}
	// 对比旧密码正确
	oldPassword := types.NewPassword(oldPass, userDo.Salt)
	old := types.NewPassword(userDo.Password, userDo.Salt)
	if !oldPassword.Equal(old) {
		return nil, merr.ErrorI18nPasswordErr(ctx)
	}

	// 对比两次密码相同, 相同修改无意义
	if newPass == oldPass {
		return nil, merr.ErrorI18nPasswordSameErr(ctx)
	}

	params := &bo.ResetUserPasswordBySelfParams{
		UserId: claims.GetUser(),
		// 使用新的盐
		Password: types.NewPassword(newPass),
	}
	if err = s.userBiz.ResetUserPasswordBySelf(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.ResetUserPasswordBySelfReply{}, nil
}

func (s *Service) GetUserSelectList(ctx context.Context, req *userapi.GetUserSelectListRequest) (*userapi.GetUserSelectListReply, error) {
	params := &bo.QueryUserSelectParams{
		Keyword: req.GetKeyword(),
		Page:    types.NewPagination(req.GetPagination()),
		Status:  vobj.Status(req.GetStatus()),
		Gender:  vobj.Gender(req.GetGender()),
		Role:    vobj.Role(req.GetRole()),
	}
	userSelectOptions, err := s.userBiz.GetUserSelectList(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.GetUserSelectListReply{
		List: types.SliceTo(userSelectOptions, func(option *bo.SelectOptionBo) *admin.Select {
			return build.NewSelectBuilder(option).ToApi()
		}),
		Pagination: build.NewPageBuilder(params.Page).ToApi(),
	}, nil
}

// UpdateUserPhone 更新用户手机号
func (s *Service) UpdateUserPhone(ctx context.Context, req *userapi.UpdateUserPhoneRequest) (*userapi.UpdateUserPhoneReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	// TODO 验证手机号短信验证码
	params := &bo.UpdateUserPhoneRequest{
		UserId: claims.GetUser(),
		Phone:  req.GetPhone(),
	}
	if err := s.userBiz.UpdateUserPhone(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserPhoneReply{}, nil
}

// UpdateUserEmail 更新用户邮箱
func (s *Service) UpdateUserEmail(ctx context.Context, req *userapi.UpdateUserEmailRequest) (*userapi.UpdateUserEmailReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	// TODO 验证邮箱验证码
	params := &bo.UpdateUserEmailRequest{
		UserId: claims.GetUser(),
		Email:  req.GetEmail(),
	}
	if err := s.userBiz.UpdateUserEmail(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserEmailReply{}, nil
}

// UpdateUserAvatar 更新用户头像
func (s *Service) UpdateUserAvatar(ctx context.Context, req *userapi.UpdateUserAvatarRequest) (*userapi.UpdateUserAvatarReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	params := &bo.UpdateUserAvatarRequest{
		UserId: claims.GetUser(),
		Avatar: req.GetAvatar(),
	}
	if err := s.userBiz.UpdateUserAvatar(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserAvatarReply{}, nil
}
