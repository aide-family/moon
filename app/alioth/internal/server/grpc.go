package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/auth"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/service"
	"github.com/aide-family/Alioth/pkg/merr"
)

func NewGRPCServer(
	c *conf.Server,
	health *service.HealthService,
	captcha *service.CaptchaService,
	authn *service.AuthService,
	self *service.SelfService,
	namespaces *service.NamespaceService,
	members *service.MemberService,
	users *service.UserService,
	roles *service.RoleService,
	uc *auth.Usecase,
	logger log.Logger,
) (*kgrpc.Server, error) {
	if c == nil || c.Grpc == nil || c.Grpc.Addr == "" {
		return nil, merr.Server("grpc addr is empty")
	}
	opts := []kgrpc.ServerOption{
		kgrpc.Address(c.Grpc.Addr),
		kgrpc.Middleware(recovery.Recovery(), authenticate(uc)),
		kgrpc.Logger(logger),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, kgrpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, kgrpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := kgrpc.NewServer(opts...)
	v1.RegisterHealthServiceServer(srv, health)
	v1.RegisterCaptchaServiceServer(srv, captcha)
	v1.RegisterAuthServiceServer(srv, authn)
	v1.RegisterSelfServiceServer(srv, self)
	v1.RegisterNamespaceServiceServer(srv, namespaces)
	v1.RegisterMemberServiceServer(srv, members)
	v1.RegisterUserServiceServer(srv, users)
	v1.RegisterRoleServiceServer(srv, roles)
	return srv, nil
}
