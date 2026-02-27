// Package oauth is the oauth package for the magicbox service.
package oauth

import (
	"context"
	nethttp "net/http"
	"net/url"
	"strings"

	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/safety"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/oauth2"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
)

const (
	loginRoutePath  = "login"
	reportRoutePath = "reports"

	OperationOAuth2Reports  = "/magicbox.oauth.OAuth2/OAuth2Reports"
	OperationOAuth2Login    = "/magicbox.oauth.OAuth2/OAuth2Login"
	OperationOAuth2Callback = "/magicbox.oauth.OAuth2/OAuth2Callback"
)

func NewOAuth2Handler(conf *config.OAuth2, redirectURLFunc RedirectURLFunc) *OAuth2Handler {
	return &OAuth2Handler{
		conf:            conf,
		redirectURLFunc: redirectURLFunc,
		oauth2RoutePath: "/oauth2",
		loginPath:       "/",
		callbackPath:    "/callback",
		loginHandler:    DefaultLoginHandler,
		callbackHandler: DefaultCallbackHandler,
		oauth2Configs:   safety.NewMap(make(map[config.OAuth2_APP]*oauth2.Config)),
	}
}

type OAuth2Handler struct {
	conf            *config.OAuth2
	loginHandler    OAuth2LoginHandlerFunc
	callbackHandler OAuth2CallbackHandlerFunc
	redirectURLFunc RedirectURLFunc

	oauth2RoutePath string
	loginPath       string
	callbackPath    string

	oauth2Configs *safety.Map[config.OAuth2_APP, *oauth2.Config]
}

type OAuth2HandlerOption func(*OAuth2Handler)

type (
	OAuth2CallbackHandlerFunc func(app config.OAuth2_APP, oauthConfig *oauth2.Config, redirectURLFunc RedirectURLFunc) (http.HandlerFunc, error)
	OAuth2LoginHandlerFunc    func(app config.OAuth2_APP, oauthConfig *oauth2.Config) (http.HandlerFunc, error)
	OAuth2LoginFun            func(ctx http.Context, oauthConfig *oauth2.Config) (*OAuth2User, error)
	RedirectURLFunc           func(ctx context.Context, req *OAuth2LoginRequest) (string, error)
)

func RegisterLoginHandler(handler OAuth2LoginHandlerFunc) OAuth2HandlerOption {
	return func(h *OAuth2Handler) {
		h.loginHandler = handler
	}
}

func RegisterCallbackHandler(handler OAuth2CallbackHandlerFunc) OAuth2HandlerOption {
	return func(h *OAuth2Handler) {
		h.callbackHandler = handler
	}
}

func BindOAuth2RoutePath(routePath string) OAuth2HandlerOption {
	return func(h *OAuth2Handler) {
		h.oauth2RoutePath = routePath
	}
}

func BindLoginPath(loginPath string) OAuth2HandlerOption {
	return func(h *OAuth2Handler) {
		h.loginPath = loginPath
	}
}

func BindCallbackPath(callbackPath string) OAuth2HandlerOption {
	return func(h *OAuth2Handler) {
		h.callbackPath = callbackPath
	}
}

func (h *OAuth2Handler) Handler(srv *http.Server) error {
	if pointer.IsNil(h.conf) || !strings.EqualFold(h.conf.GetEnable(), "true") {
		klog.Debug("oauth2 is not enabled")
		return nil
	}

	routePrintList := make([]string, 0, len(h.conf.GetConfigs()))
	oauth2Route := srv.Route(h.oauth2RoutePath)
	loginRoute := oauth2Route.Group(loginRoutePath)
	for _, config := range h.conf.GetConfigs() {
		app := config.GetApp()
		authConfigItem := &oauth2.Config{
			ClientID:     config.GetClientId(),
			ClientSecret: config.GetClientSecret(),
			RedirectURL:  config.GetCallbackUri(),
			Scopes:       config.GetScopes(),
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.GetAuthUrl(),
				TokenURL: config.GetTokenUrl(),
			},
		}
		h.oauth2Configs.Set(app, authConfigItem)
		appPath := strings.ToLower(app.String())
		appRoute := loginRoute.Group(appPath)
		loginHandler, err := h.loginHandler(app, authConfigItem)
		if err != nil {
			klog.Warnf("login handler failed: %v", err)
			return err
		}
		appRoute.GET(h.loginPath, loginHandler)
		callbackHandler, err := h.callbackHandler(app, authConfigItem, h.redirectURLFunc)
		if err != nil {
			klog.Warnf("callback handler failed: %v", err)
			return err
		}
		appRoute.GET(h.callbackPath, callbackHandler)
		loginURL, _ := url.JoinPath(h.oauth2RoutePath, loginRoutePath, appPath, h.loginPath)
		callbackURL, _ := url.JoinPath(h.oauth2RoutePath, loginRoutePath, appPath, h.callbackPath)
		routePrintList = append(routePrintList, loginURL, callbackURL)
	}
	oauth2Route.GET(reportRoutePath, h.OAuth2Reports())
	reportURL, _ := url.JoinPath(h.oauth2RoutePath, reportRoutePath)
	routePrintList = append(routePrintList, reportURL)
	for _, route := range routePrintList {
		klog.Debugf("OAuth2 route: %s", route)
	}
	return nil
}

func (h *OAuth2Handler) OAuth2Reports() http.HandlerFunc {
	reports := make([]OAuth2ReportItem, 0, len(h.conf.GetConfigs()))
	for _, config := range h.conf.GetConfigs() {
		reports = append(reports, OAuth2ReportItem{
			App:      config.GetApp().String(),
			LoginUrl: config.GetLoginUrl(),
		})
	}
	return func(ctx http.Context) error {
		http.SetOperation(ctx, OperationOAuth2Reports)
		h := ctx.Middleware(func(ctx context.Context, _ interface{}) (interface{}, error) {
			return append([]OAuth2ReportItem{}, reports...), nil
		})
		out, err := h(ctx, nil)
		if err != nil {
			klog.Warnf("oauth2 reports failed: %v", err)
			return err
		}
		reply := out.([]OAuth2ReportItem)
		return ctx.Result(nethttp.StatusOK, reply)
	}
}

func DefaultLoginHandler(app config.OAuth2_APP, oauthConfig *oauth2.Config) (http.HandlerFunc, error) {
	return func(ctx http.Context) error {
		http.SetOperation(ctx, OperationOAuth2Login)
		h := ctx.Middleware(func(ctx context.Context, _ interface{}) (interface{}, error) {
			return nil, nil
		})
		_, _ = h(ctx, nil)

		// Redirect to the specified URL
		url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOnline)
		req := ctx.Request()
		resp := ctx.Response()
		resp.Header().Set("Location", url)
		resp.WriteHeader(nethttp.StatusTemporaryRedirect)
		ctx.Reset(resp, req)
		return nil
	}, nil
}

func DefaultCallbackHandler(app config.OAuth2_APP, oauthConfig *oauth2.Config, redirectURLFunc RedirectURLFunc) (http.HandlerFunc, error) {
	login, ok := GetOAuth2LoginFun(app)
	if !ok {
		return nil, merr.ErrorInternalServer("app %s login fun not registered", app)
	}

	return func(ctx http.Context) error {
		http.SetOperation(ctx, OperationOAuth2Callback)
		oauth2LoginUser, err := login(ctx, oauthConfig)
		if err != nil {
			return err
		}

		oauth2LoginRequest := &OAuth2LoginRequest{
			App:  app,
			User: oauth2LoginUser,
			Config: &config.OAuth2_Config{
				ClientId:     oauthConfig.ClientID,
				ClientSecret: oauthConfig.ClientSecret,
				CallbackUri:  oauthConfig.RedirectURL,
				AuthUrl:      oauthConfig.Endpoint.AuthURL,
				TokenUrl:     oauthConfig.Endpoint.TokenURL,
				Scopes:       oauthConfig.Scopes,
			},
		}
		if err := ctx.BindQuery(oauth2LoginRequest); err != nil {
			return err
		}
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			redirectURLStr, err := redirectURLFunc(ctx, req.(*OAuth2LoginRequest))
			if err != nil {
				return nil, err
			}
			return redirectURLStr, nil
		})

		redirectURLStr, err := h(ctx, oauth2LoginRequest)
		if err != nil {
			return err
		}

		redirectURL, err := url.Parse(redirectURLStr.(string))
		if err != nil {
			return err
		}

		resp := ctx.Response()
		resp.Header().Set("Location", redirectURL.String())
		resp.WriteHeader(nethttp.StatusTemporaryRedirect)
		ctx.Reset(resp, ctx.Request())
		return nil
	}, nil
}
