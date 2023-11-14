package authservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/auth"
	"prometheus-manager/pkg/helper"
)

type AuthService struct {
	pb.UnimplementedAuthServer
	log *log.Helper
}

func NewAuthService(logger log.Logger) *AuthService {
	return &AuthService{
		log: log.NewHelper(log.With(logger, "module", "service.auth")),
	}
}

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	token, err := helper.IssueToken(1, "admin")
	if err != nil {
		return nil, err
	}
	return &pb.LoginReply{Token: token}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutReply, error) {
	return &pb.LogoutReply{}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenReply, error) {
	authClaims, ok := helper.GetAuthClaims(ctx)
	if !ok {
		return nil, helper.ErrTokenInvalid
	}

	token, err := helper.IssueToken(authClaims.ID, authClaims.Role)
	if err != nil {
		return nil, err
	}
	return &pb.RefreshTokenReply{Token: token}, nil
}
