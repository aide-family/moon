package authorization

import (
	"context"

	pb "github.com/aide-cloud/moon/api/admin/authorization"
)

type Service struct {
	pb.UnimplementedAuthorizationServer
}

func NewAuthorizationService() *Service {
	return &Service{}
}

func (s *Service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	return &pb.LoginReply{}, nil
}
func (s *Service) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutReply, error) {
	return &pb.LogoutReply{}, nil
}
func (s *Service) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenReply, error) {
	return &pb.RefreshTokenReply{}, nil
}
func (s *Service) Captcha(ctx context.Context, req *pb.CaptchaReq) (*pb.CaptchaReply, error) {
	return &pb.CaptchaReply{}, nil
}
