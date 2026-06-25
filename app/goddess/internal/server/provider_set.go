// Package server is a server package for kratos.
package server

import (
	"context"
	"embed"
	nethttp "net/http"
	"strings"

	magicboxv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/auth/basic"
	"github.com/aide-family/magicbox/oauth"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/goddess/internal/service"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

//go:embed swagger
var docFS embed.FS

var (
	ProviderSetServerAll  = wire.NewSet(NewHTTPServer, NewGRPCServer, RegisterService)
	ProviderSetServerHTTP = wire.NewSet(NewHTTPServer, RegisterHTTPService)
	ProviderSetServerGRPC = wire.NewSet(NewGRPCServer, RegisterGRPCService)
)

type Server interface {
	transport.Server
	Name() string
	Instance() transport.Server
}

type server struct {
	transport.Server
	name string
}

func (s *server) Name() string {
	return s.name
}

func (s *server) Instance() transport.Server {
	return s.Server
}

func newServer(name string, srv transport.Server) Server {
	return &server{
		Server: srv,
		name:   name,
	}
}

type Servers []Server

func BindSwagger(httpSrv *http.Server, bc *conf.Bootstrap) {
	binding := basic.HandlerBinding{
		Name:      "Swagger",
		Enabled:   strings.EqualFold(bc.GetEnableSwagger(), "true"),
		BasicAuth: bc.GetSwaggerBasicAuth(),
		Handler:   nethttp.StripPrefix("/doc/", nethttp.FileServer(nethttp.FS(docFS))),
		Path:      "/doc/",
	}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

func BindMetrics(httpSrv *http.Server, bc *conf.Bootstrap) {
	binding := basic.HandlerBinding{
		Name:      "Metrics",
		Enabled:   strings.EqualFold(bc.GetEnableMetrics(), "true"),
		BasicAuth: bc.GetMetricsBasicAuth(),
		Handler:   promhttp.Handler(),
		Path:      "/metrics",
	}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

// RegisterService registers the service.
func RegisterService(
	c *conf.Bootstrap,
	httpSrv *http.Server,
	grpcSrv *grpc.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
	captchaService *service.CaptchaService,
	sysRoleService *service.SysRoleService,
	sysMenuService *service.SysMenuService,
	sysFunctionService *service.SysFunctionService,
	sysUserRoleService *service.SysUserRoleService,
	sysRoleMenuService *service.SysRoleMenuService,
	sysRoleFunctionService *service.SysRoleFunctionService,
	sysUserMenuService *service.SysUserMenuService,
	sysUserFunctionService *service.SysUserFunctionService,
	nmspacesRoleService *service.NmspacesRoleService,
	nmspacesMenuService *service.NmspacesMenuService,
	nmspacesFunctionService *service.NmspacesFunctionService,
	nmspacesMemberRoleService *service.NmspacesMemberRoleService,
	nmspacesRoleMenuService *service.NmspacesRoleMenuService,
	nmspacesRoleFunctionService *service.NmspacesRoleFunctionService,
	nmspacesMemberMenuService *service.NmspacesMemberMenuService,
	nmspacesMemberFunctionService *service.NmspacesMemberFunctionService,
) Servers {
	var srvs Servers

	srvs = append(srvs, RegisterHTTPService(c, httpSrv,
		authService,
		healthService,
		namespaceService,
		userService,
		memberService,
		selfService,
		captchaService,
		sysRoleService,
		sysMenuService,
		sysFunctionService,
		sysUserRoleService,
		sysRoleMenuService,
		sysRoleFunctionService,
		sysUserMenuService,
		sysUserFunctionService,
		nmspacesRoleService,
		nmspacesMenuService,
		nmspacesFunctionService,
		nmspacesMemberRoleService,
		nmspacesRoleMenuService,
		nmspacesRoleFunctionService,
		nmspacesMemberMenuService,
		nmspacesMemberFunctionService,
	)...)
	srvs = append(srvs, RegisterGRPCService(c, grpcSrv,
		authService,
		healthService,
		namespaceService,
		userService,
		memberService,
		selfService,
		captchaService,
		sysRoleService,
		sysMenuService,
		sysFunctionService,
		sysUserRoleService,
		sysRoleMenuService,
		sysRoleFunctionService,
		sysUserMenuService,
		sysUserFunctionService,
		nmspacesRoleService,
		nmspacesMenuService,
		nmspacesFunctionService,
		nmspacesMemberRoleService,
		nmspacesRoleMenuService,
		nmspacesRoleFunctionService,
		nmspacesMemberMenuService,
		nmspacesMemberFunctionService,
	)...)
	return srvs
}

// RegisterHTTPService registers only HTTP service.
func RegisterHTTPService(
	c *conf.Bootstrap,
	httpSrv *http.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
	captchaService *service.CaptchaService,
	sysRoleService *service.SysRoleService,
	sysMenuService *service.SysMenuService,
	sysFunctionService *service.SysFunctionService,
	sysUserRoleService *service.SysUserRoleService,
	sysRoleMenuService *service.SysRoleMenuService,
	sysRoleFunctionService *service.SysRoleFunctionService,
	sysUserMenuService *service.SysUserMenuService,
	sysUserFunctionService *service.SysUserFunctionService,
	nmspacesRoleService *service.NmspacesRoleService,
	nmspacesMenuService *service.NmspacesMenuService,
	nmspacesFunctionService *service.NmspacesFunctionService,
	nmspacesMemberRoleService *service.NmspacesMemberRoleService,
	nmspacesRoleMenuService *service.NmspacesRoleMenuService,
	nmspacesRoleFunctionService *service.NmspacesRoleFunctionService,
	nmspacesMemberMenuService *service.NmspacesMemberMenuService,
	nmspacesMemberFunctionService *service.NmspacesMemberFunctionService,
) Servers {
	magicboxv1.RegisterHealthHTTPServer(httpSrv, healthService)
	goddessv1.RegisterAuthServiceHTTPServer(httpSrv, authService)
	goddessv1.RegisterNamespaceHTTPServer(httpSrv, namespaceService)
	goddessv1.RegisterUserHTTPServer(httpSrv, userService)
	goddessv1.RegisterMemberHTTPServer(httpSrv, memberService)
	goddessv1.RegisterSelfHTTPServer(httpSrv, selfService)
	goddessv1.RegisterCaptchaHTTPServer(httpSrv, captchaService)
	goddessv1.RegisterSysRoleHTTPServer(httpSrv, sysRoleService)
	goddessv1.RegisterSysMenuHTTPServer(httpSrv, sysMenuService)
	goddessv1.RegisterSysFunctionHTTPServer(httpSrv, sysFunctionService)
	goddessv1.RegisterSysUserRoleHTTPServer(httpSrv, sysUserRoleService)
	goddessv1.RegisterSysRoleMenuHTTPServer(httpSrv, sysRoleMenuService)
	goddessv1.RegisterSysRoleFunctionHTTPServer(httpSrv, sysRoleFunctionService)
	goddessv1.RegisterSysUserMenuHTTPServer(httpSrv, sysUserMenuService)
	goddessv1.RegisterSysUserFunctionHTTPServer(httpSrv, sysUserFunctionService)
	goddessv1.RegisterNmspacesRoleHTTPServer(httpSrv, nmspacesRoleService)
	goddessv1.RegisterNmspacesMenuHTTPServer(httpSrv, nmspacesMenuService)
	goddessv1.RegisterNmspacesFunctionHTTPServer(httpSrv, nmspacesFunctionService)
	goddessv1.RegisterNmspacesMemberRoleHTTPServer(httpSrv, nmspacesMemberRoleService)
	goddessv1.RegisterNmspacesRoleMenuHTTPServer(httpSrv, nmspacesRoleMenuService)
	goddessv1.RegisterNmspacesRoleFunctionHTTPServer(httpSrv, nmspacesRoleFunctionService)
	goddessv1.RegisterNmspacesMemberMenuHTTPServer(httpSrv, nmspacesMemberMenuService)
	goddessv1.RegisterNmspacesMemberFunctionHTTPServer(httpSrv, nmspacesMemberFunctionService)

	oauth2Handler := oauth.NewOAuth2Handler(c.GetOauth2(), func(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error) {
		reply, err := authService.OAuth2Login(ctx, req)
		if err != nil {
			return "", err
		}
		return reply.Token, nil
	})
	if err := oauth2Handler.Handler(httpSrv); err != nil {
		panic(err)
	}
	return Servers{newServer("http", httpSrv)}
}

// RegisterGRPCService registers only gRPC service.
func RegisterGRPCService(
	c *conf.Bootstrap,
	grpcSrv *grpc.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
	captchaService *service.CaptchaService,
	sysRoleService *service.SysRoleService,
	sysMenuService *service.SysMenuService,
	sysFunctionService *service.SysFunctionService,
	sysUserRoleService *service.SysUserRoleService,
	sysRoleMenuService *service.SysRoleMenuService,
	sysRoleFunctionService *service.SysRoleFunctionService,
	sysUserMenuService *service.SysUserMenuService,
	sysUserFunctionService *service.SysUserFunctionService,
	nmspacesRoleService *service.NmspacesRoleService,
	nmspacesMenuService *service.NmspacesMenuService,
	nmspacesFunctionService *service.NmspacesFunctionService,
	nmspacesMemberRoleService *service.NmspacesMemberRoleService,
	nmspacesRoleMenuService *service.NmspacesRoleMenuService,
	nmspacesRoleFunctionService *service.NmspacesRoleFunctionService,
	nmspacesMemberMenuService *service.NmspacesMemberMenuService,
	nmspacesMemberFunctionService *service.NmspacesMemberFunctionService,
) Servers {
	magicboxv1.RegisterHealthServer(grpcSrv, healthService)
	goddessv1.RegisterAuthServiceServer(grpcSrv, authService)
	goddessv1.RegisterNamespaceServer(grpcSrv, namespaceService)
	goddessv1.RegisterUserServer(grpcSrv, userService)
	goddessv1.RegisterMemberServer(grpcSrv, memberService)
	goddessv1.RegisterSelfServer(grpcSrv, selfService)
	goddessv1.RegisterCaptchaServer(grpcSrv, captchaService)
	goddessv1.RegisterSysRoleServer(grpcSrv, sysRoleService)
	goddessv1.RegisterSysMenuServer(grpcSrv, sysMenuService)
	goddessv1.RegisterSysFunctionServer(grpcSrv, sysFunctionService)
	goddessv1.RegisterSysUserRoleServer(grpcSrv, sysUserRoleService)
	goddessv1.RegisterSysRoleMenuServer(grpcSrv, sysRoleMenuService)
	goddessv1.RegisterSysRoleFunctionServer(grpcSrv, sysRoleFunctionService)
	goddessv1.RegisterSysUserMenuServer(grpcSrv, sysUserMenuService)
	goddessv1.RegisterSysUserFunctionServer(grpcSrv, sysUserFunctionService)
	goddessv1.RegisterNmspacesRoleServer(grpcSrv, nmspacesRoleService)
	goddessv1.RegisterNmspacesMenuServer(grpcSrv, nmspacesMenuService)
	goddessv1.RegisterNmspacesFunctionServer(grpcSrv, nmspacesFunctionService)
	goddessv1.RegisterNmspacesMemberRoleServer(grpcSrv, nmspacesMemberRoleService)
	goddessv1.RegisterNmspacesRoleMenuServer(grpcSrv, nmspacesRoleMenuService)
	goddessv1.RegisterNmspacesRoleFunctionServer(grpcSrv, nmspacesRoleFunctionService)
	goddessv1.RegisterNmspacesMemberMenuServer(grpcSrv, nmspacesMemberMenuService)
	goddessv1.RegisterNmspacesMemberFunctionServer(grpcSrv, nmspacesMemberFunctionService)
	return Servers{newServer("grpc", grpcSrv)}
}

var namespaceAllowList = []string{
	goddessv1.OperationNamespaceCreateNamespace,
	goddessv1.OperationNamespaceUpdateNamespace,
	goddessv1.OperationNamespaceUpdateNamespaceStatus,
	goddessv1.OperationNamespaceDeleteNamespace,
	goddessv1.OperationNamespaceGetNamespace,
	goddessv1.OperationNamespaceListNamespace,
	goddessv1.OperationNamespaceSelectNamespace,
	goddessv1.OperationSelfInfo,
	goddessv1.OperationSelfNamespaces,
	goddessv1.OperationSelfChangeEmail,
	goddessv1.OperationSelfChangeAvatar,
	goddessv1.OperationSelfRefreshToken,
	goddessv1.OperationUserGetUser,
	goddessv1.OperationUserListUser,
	goddessv1.OperationUserSelectUser,
	goddessv1.OperationUserBanUser,
	goddessv1.OperationUserPermitUser,
	goddessv1.OperationSysRoleListSysRole,
	goddessv1.OperationSysMenuListSysMenu,
	goddessv1.OperationSysMenuTreeSysMenu,
	goddessv1.OperationSysMenuGetSysMenu,
	goddessv1.OperationSysMenuCreateSysMenu,
	goddessv1.OperationSysMenuUpdateSysMenu,
	goddessv1.OperationSysMenuDeleteSysMenu,
	goddessv1.OperationSysFunctionCreateSysFunction,
	goddessv1.OperationSysFunctionListSysFunction,
	goddessv1.OperationSysFunctionGetSysFunction,
	goddessv1.OperationSysFunctionUpdateSysFunction,
	goddessv1.OperationSysFunctionDeleteSysFunction,
	goddessv1.OperationSysUserRoleListSysUserRole,
	goddessv1.OperationSysUserRoleBindSysUserRoles,
	goddessv1.OperationSysUserRoleUnbindSysUserRoles,
	goddessv1.OperationSysRoleMenuListRoleMenu,
	goddessv1.OperationSysRoleMenuGrantRoleMenus,
	goddessv1.OperationSysRoleMenuRevokeRoleMenus,
	goddessv1.OperationSysRoleFunctionListRoleFunction,
	goddessv1.OperationSysRoleFunctionGrantRoleFunctions,
	goddessv1.OperationSysRoleFunctionRevokeRoleFunctions,
	goddessv1.OperationSysUserMenuListUserMenu,
	goddessv1.OperationSysUserMenuGrantUserMenus,
	goddessv1.OperationSysUserMenuRevokeUserMenus,
	goddessv1.OperationSysUserFunctionListUserFunction,
	goddessv1.OperationSysUserFunctionGrantUserFunctions,
	goddessv1.OperationSysUserFunctionRevokeUserFunctions,
	goddessv1.OperationUserListUserFinalMenus,
	goddessv1.OperationUserListUserFinalFunctions,
	goddessv1.OperationNmspacesRoleListNmspacesRole,
	goddessv1.OperationNmspacesRoleGetNmspacesRole,
	goddessv1.OperationNmspacesRoleCreateNmspacesRole,
	goddessv1.OperationNmspacesRoleUpdateNmspacesRole,
	goddessv1.OperationNmspacesRoleDeleteNmspacesRole,
	goddessv1.OperationNmspacesMenuListNmspacesMenu,
	goddessv1.OperationNmspacesMenuGetNmspacesMenu,
	goddessv1.OperationNmspacesMenuCreateNmspacesMenu,
	goddessv1.OperationNmspacesMenuUpdateNmspacesMenu,
	goddessv1.OperationNmspacesMenuDeleteNmspacesMenu,
	goddessv1.OperationNmspacesFunctionCreateNmspacesFunction,
	goddessv1.OperationNmspacesFunctionListNmspacesFunction,
	goddessv1.OperationNmspacesFunctionGetNmspacesFunction,
	goddessv1.OperationNmspacesFunctionUpdateNmspacesFunction,
	goddessv1.OperationNmspacesFunctionDeleteNmspacesFunction,
	goddessv1.OperationNmspacesMemberRoleListNmspacesMemberRole,
	goddessv1.OperationNmspacesMemberRoleBindNmspacesMemberRoles,
	goddessv1.OperationNmspacesMemberRoleUnbindNmspacesMemberRoles,
	goddessv1.OperationNmspacesRoleMenuListNmspacesRoleMenu,
	goddessv1.OperationNmspacesRoleMenuGrantNmspacesRoleMenus,
	goddessv1.OperationNmspacesRoleMenuRevokeNmspacesRoleMenus,
	goddessv1.OperationNmspacesRoleFunctionListNmspacesRoleFunction,
	goddessv1.OperationNmspacesRoleFunctionGrantNmspacesRoleFunctions,
	goddessv1.OperationNmspacesRoleFunctionRevokeNmspacesRoleFunctions,
	goddessv1.OperationNmspacesMemberMenuListNmspacesMemberMenu,
	goddessv1.OperationNmspacesMemberMenuGrantNmspacesMemberMenus,
	goddessv1.OperationNmspacesMemberMenuRevokeNmspacesMemberMenus,
	goddessv1.OperationNmspacesMemberFunctionListNmspacesMemberFunction,
	goddessv1.OperationNmspacesMemberFunctionGrantNmspacesMemberFunctions,
	goddessv1.OperationNmspacesMemberFunctionRevokeNmspacesMemberFunctions,
	goddessv1.OperationUserListNmspaceFinalMenus,
	goddessv1.OperationUserListNmspaceFinalFunctions,
}

var authAllowList = []string{
	magicboxv1.OperationHealthHealthCheck,
	oauth.OperationOAuth2Reports,
	oauth.OperationOAuth2Login,
	oauth.OperationOAuth2Callback,
	goddessv1.OperationAuthServiceEmailLogin,
	goddessv1.OperationAuthServiceSendEmailLoginCode,
	goddessv1.OperationCaptchaGetCaptcha,
	goddessv1.OperationNamespaceGetNamespaceSimple,
}
