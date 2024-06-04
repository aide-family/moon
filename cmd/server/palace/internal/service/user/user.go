package user

import (
	"context"

	"github.com/aide-cloud/moon/api/admin"
	pb "github.com/aide-cloud/moon/api/admin/user"
	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/utils/cipher"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type Service struct {
	pb.UnimplementedUserServer

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
func (s *Service) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
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

	createParams := &bo.CreateUserParams{
		Name:      req.GetName(),
		Password:  types.NewPassword(pass),
		Email:     req.GetEmail(),
		Phone:     req.GetPhone(),
		Nickname:  req.GetNickname(),
		Remark:    req.GetRemark(),
		Avatar:    req.GetAvatar(),
		CreatorID: claims.GetUser(),
		Status:    vobj.Status(req.GetStatus()),
		Gender:    vobj.Gender(req.GetGender()),
		Role:      vobj.Role(req.GetRole()),
	}
	_, err = s.userBiz.CreateUser(ctx, createParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.CreateUserReply{}, nil
}

// UpdateUser 更新用户基础信息， 只允许管理员操作
func (s *Service) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	data := req.GetData()
	createParams := bo.CreateUserParams{
		Name:     data.GetName(),
		Email:    data.GetEmail(),
		Phone:    data.GetPhone(),
		Nickname: data.GetNickname(),
		Remark:   data.GetRemark(),
		Avatar:   data.GetAvatar(),
		Status:   vobj.Status(data.GetStatus()),
		Gender:   vobj.Gender(data.GetGender()),
		Role:     vobj.Role(data.GetRole()),
	}
	if err := s.userBiz.UpdateUser(ctx, &bo.UpdateUserParams{
		ID:               req.GetId(),
		CreateUserParams: createParams,
	}); !types.IsNil(err) {
		return nil, err
	}
	return &pb.UpdateUserReply{}, nil
}

func (s *Service) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserReply, error) {
	if err := s.userBiz.DeleteUser(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &pb.DeleteUserReply{}, nil
}

func (s *Service) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserReply, error) {
	userDo, err := s.userBiz.GetUser(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &pb.GetUserReply{
		User: build.NewUserBuild(userDo).ToApi(),
	}, nil
}

func (s *Service) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserReply, error) {
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
	return &pb.ListUserReply{
		List: types.SliceTo(userDos, func(user *model.SysUser) *admin.User {
			return build.NewUserBuild(user).ToApi()
		}),
		Pagination: build.NewPageBuild(queryParams.Page).ToApi(),
	}, nil
}

func (s *Service) BatchUpdateUserStatus(ctx context.Context, req *pb.BatchUpdateUserStatusRequest) (*pb.BatchUpdateUserStatusReply, error) {
	params := &bo.BatchUpdateUserStatusParams{
		Status: vobj.Status(req.GetStatus()),
		IDs:    req.GetIds(),
	}
	if err := s.userBiz.BatchUpdateUserStatus(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &pb.BatchUpdateUserStatusReply{}, nil
}

func (s *Service) ResetUserPassword(ctx context.Context, req *pb.ResetUserPasswordRequest) (*pb.ResetUserPasswordReply, error) {
	return &pb.ResetUserPasswordReply{}, nil
}

func (s *Service) ResetUserPasswordBySelf(ctx context.Context, req *pb.ResetUserPasswordBySelfRequest) (*pb.ResetUserPasswordBySelfReply, error) {
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
	return &pb.ResetUserPasswordBySelfReply{}, nil
}

func (s *Service) GetUserSelectList(ctx context.Context, req *pb.GetUserSelectListRequest) (*pb.GetUserSelectListReply, error) {
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
	return &pb.GetUserSelectListReply{
		List: types.SliceTo(userSelectOptions, func(option *bo.SelectOptionBo) *admin.Select {
			return build.NewSelectBuild(option).ToApi()
		}),
		Pagination: build.NewPageBuild(params.Page).ToApi(),
	}, nil
}

// UpdateUserPhone 更新用户手机号
func (s *Service) UpdateUserPhone(ctx context.Context, req *pb.UpdateUserPhoneRequest) (*pb.UpdateUserPhoneReply, error) {
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
	return &pb.UpdateUserPhoneReply{}, nil
}

// UpdateUserEmail 更新用户邮箱
func (s *Service) UpdateUserEmail(ctx context.Context, req *pb.UpdateUserEmailRequest) (*pb.UpdateUserEmailReply, error) {
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
	return &pb.UpdateUserEmailReply{}, nil
}

// UpdateUserAvatar 更新用户头像
func (s *Service) UpdateUserAvatar(ctx context.Context, req *pb.UpdateUserAvatarRequest) (*pb.UpdateUserAvatarReply, error) {
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
	return &pb.UpdateUserAvatarReply{}, nil
}
