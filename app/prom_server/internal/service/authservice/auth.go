package authservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/auth"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/pkg/helper"
	"prometheus-manager/pkg/util/password"
)

type AuthService struct {
	pb.UnimplementedAuthServer
	log *log.Helper

	userBiz *biz.UserBiz
}

func NewAuthService(userBiz *biz.UserBiz, logger log.Logger) *AuthService {
	return &AuthService{
		log:     log.NewHelper(log.With(logger, "module", "service.auth")),
		userBiz: userBiz,
	}
}

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	// TODO 认证传递的code, 前端需要校验code的合法性
	pwd := req.GetPassword()
	// 解密前端传递的密码, 拒绝明文传输
	dePwd, err := password.DecryptPassword(pwd, password.DefaultIv)
	if err != nil {
		return nil, perrors.ErrorInvalidParams("密码不规范")
	}
	// 颁发token, 时间建议设置为半天以内
	token, err := s.userBiz.LoginByUsernameAndPassword(ctx, req.GetUsername(), dePwd)
	if err != nil {
		s.log.Warnf("LoginByUsernameAndPassword error: %v", err)
		return nil, perrors.ErrorUnknown("颁发token失败")
	}
	return &pb.LoginReply{Token: token}, nil
}

func (s *AuthService) Logout(ctx context.Context, _ *pb.LogoutRequest) (*pb.LogoutReply, error) {
	authClaims, ok := helper.GetAuthClaims(ctx)
	if !ok {
		return nil, helper.ErrTokenInvalid
	}
	// 记录token md5然后存储到redis
	if err := s.userBiz.Logout(ctx, authClaims); err != nil {
		return nil, err
	}

	return &pb.LogoutReply{}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenReply, error) {
	authClaims, ok := helper.GetAuthClaims(ctx)
	if !ok {
		return nil, helper.ErrTokenInvalid
	}

	token, err := s.userBiz.RefreshToken(ctx, authClaims)
	if err != nil {
		return nil, err
	}
	return &pb.RefreshTokenReply{Token: token}, nil
}
