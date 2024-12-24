package user

import (
	"context"

	userapi "github.com/aide-family/moon/api/admin/user"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/util/types"
)

// Service 用户管理服务
type Service struct {
	userapi.UnimplementedUserServer

	userBiz *biz.UserBiz
}

// NewUserService 创建用户服务
func NewUserService(userBiz *biz.UserBiz) *Service {
	return &Service{
		userBiz: userBiz,
	}
}

// CreateUser 创建用户 只允许管理员操作
func (s *Service) CreateUser(ctx context.Context, req *userapi.CreateUserRequest) (*userapi.CreateUserReply, error) {
	createParams := builder.NewParamsBuild(ctx).UserModuleBuilder().WithCreateUserRequest(req).ToBo()
	_, err := s.userBiz.CreateUser(ctx, createParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.CreateUserReply{}, nil
}

// UpdateUser 更新用户基础信息， 只允许管理员操作
func (s *Service) UpdateUser(ctx context.Context, req *userapi.UpdateUserRequest) (*userapi.UpdateUserReply, error) {
	updateParams := builder.NewParamsBuild(ctx).UserModuleBuilder().WithUpdateUserRequest(req).ToBo()
	if err := s.userBiz.UpdateUser(ctx, updateParams); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserReply{}, nil
}

// DeleteUser 删除用户 只允许管理员操作
func (s *Service) DeleteUser(ctx context.Context, req *userapi.DeleteUserRequest) (*userapi.DeleteUserReply, error) {
	if err := s.userBiz.DeleteUser(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.DeleteUserReply{}, nil
}

// GetUser 获取用户详情
func (s *Service) GetUser(ctx context.Context, req *userapi.GetUserRequest) (*userapi.GetUserReply, error) {
	userDo, err := s.userBiz.GetUser(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.GetUserReply{
		Detail: builder.NewParamsBuild(ctx).UserModuleBuilder().DoUserBuilder().ToAPI(userDo),
	}, nil
}

// ListUser 获取用户列表
func (s *Service) ListUser(ctx context.Context, req *userapi.ListUserRequest) (*userapi.ListUserReply, error) {
	queryParams := builder.NewParamsBuild(ctx).UserModuleBuilder().WithListUserRequest(req).ToBo()
	userDos, err := s.userBiz.ListUser(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.ListUserReply{
		List:       builder.NewParamsBuild(ctx).UserModuleBuilder().DoUserBuilder().ToAPIs(userDos),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(queryParams.Page),
	}, nil
}

// BatchUpdateUserStatus 批量更新用户状态
func (s *Service) BatchUpdateUserStatus(ctx context.Context, req *userapi.BatchUpdateUserStatusRequest) (*userapi.BatchUpdateUserStatusReply, error) {
	params := builder.NewParamsBuild(ctx).UserModuleBuilder().WithBatchUpdateUserStatusRequest(req).ToBo()
	if err := s.userBiz.BatchUpdateUserStatus(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.BatchUpdateUserStatusReply{}, nil
}

// ResetUserPassword 重置用户密码
func (s *Service) ResetUserPassword(ctx context.Context, req *userapi.ResetUserPasswordRequest) (*userapi.ResetUserPasswordReply, error) {
	// TODO 发送邮件等相关操作
	return &userapi.ResetUserPasswordReply{}, nil
}

// ResetUserPasswordBySelf 重置用户密码
func (s *Service) ResetUserPasswordBySelf(ctx context.Context, req *userapi.ResetUserPasswordBySelfRequest) (*userapi.ResetUserPasswordBySelfReply, error) {
	builderUser := builder.NewParamsBuild(ctx).UserModuleBuilder().WithResetUserPasswordBySelfRequest(req)
	checkBuilder, err := builderUser.WithUserInfo(s.userBiz.GetUser)
	if !types.IsNil(err) {
		return nil, err
	}
	params := checkBuilder.ToBo()
	if err := s.userBiz.ResetUserPasswordBySelf(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.ResetUserPasswordBySelfReply{}, nil
}

// GetUserSelectList 获取用户下拉列表
func (s *Service) GetUserSelectList(ctx context.Context, req *userapi.ListUserRequest) (*userapi.GetUserSelectListReply, error) {
	params := builder.NewParamsBuild(ctx).UserModuleBuilder().WithListUserRequest(req).ToBo()
	userSelectOptions, err := s.userBiz.ListUser(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.GetUserSelectListReply{
		List:       builder.NewParamsBuild(ctx).UserModuleBuilder().DoUserBuilder().ToSelects(userSelectOptions),
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}

// UpdateUserPhone 更新用户手机号
func (s *Service) UpdateUserPhone(ctx context.Context, req *userapi.UpdateUserPhoneRequest) (*userapi.UpdateUserPhoneReply, error) {
	// TODO 验证手机号短信验证码
	params := builder.NewParamsBuild(ctx).UserModuleBuilder().WithUpdateUserPhoneRequest(req).ToBo()
	if err := s.userBiz.UpdateUserPhone(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserPhoneReply{}, nil
}

// UpdateUserEmail 更新用户邮箱
func (s *Service) UpdateUserEmail(ctx context.Context, req *userapi.UpdateUserEmailRequest) (*userapi.UpdateUserEmailReply, error) {
	// TODO 验证邮箱验证码
	params := builder.NewParamsBuild(ctx).UserModuleBuilder().WithUpdateUserEmailRequest(req).ToBo()
	if err := s.userBiz.UpdateUserEmail(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserEmailReply{}, nil
}

// UpdateUserAvatar 更新用户头像
func (s *Service) UpdateUserAvatar(ctx context.Context, req *userapi.UpdateUserAvatarRequest) (*userapi.UpdateUserAvatarReply, error) {
	params := builder.NewParamsBuild(ctx).UserModuleBuilder().WithUpdateUserAvatarRequest(req).ToBo()
	if err := s.userBiz.UpdateUserAvatar(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserAvatarReply{}, nil
}

// UpdateUserBaseInfo 更新用户基础信息
func (s *Service) UpdateUserBaseInfo(ctx context.Context, req *userapi.UpdateUserBaseInfoRequest) (*userapi.UpdateUserBaseInfoReply, error) {
	updateParams := builder.NewParamsBuild(ctx).UserModuleBuilder().WithUpdateUserBaseInfoRequest(req).ToBo()
	if err := s.userBiz.UpdateUserBaseInfo(ctx, updateParams); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.UpdateUserBaseInfoReply{}, nil
}

// GetUserSelfBasic 获取用户个人基础信息
func (s *Service) GetUserSelfBasic(ctx context.Context, _ *userapi.GetUserSelfBasicRequest) (*userapi.GetUserSelfBasicReply, error) {
	userDo, err := s.userBiz.GetUser(ctx, middleware.GetUserID(ctx))
	if !types.IsNil(err) {
		return nil, err
	}
	return &userapi.GetUserSelfBasicReply{
		Detail: builder.NewParamsBuild(ctx).UserModuleBuilder().DoUserBuilder().ToAPI(userDo),
	}, nil
}

// SetUserRole 设置用户角色
func (s *Service) SetUserRole(ctx context.Context, req *userapi.SetUserRoleRequest) (*userapi.SetUserRoleReply, error) {
	params := builder.NewParamsBuild(ctx).UserModuleBuilder().WithSetUserRoleRequest(req).ToBo()
	if err := s.userBiz.SetUserRole(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &userapi.SetUserRoleReply{}, nil
}
