package impl

import (
	"context"

	authv1 "github.com/aide-family/magicbox/domain/auth/v1"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/oauth"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

type loginRepository struct {
	repo authv1.Repository
}

func NewLoginRepository(c *conf.Bootstrap, d *data.Data) (repository.LoginRepository, error) {
	repoConfig := c.GetLoginConfig()
	version := repoConfig.GetVersion()
	driver := repoConfig.GetDriver()
	switch version {
	default:
		factory, ok := authv1.GetAuthV1Factory(driver)
		if !ok {
			return nil, merr.ErrorInternalServer("auth repository factory not found")
		}
		repoImpl, close, err := factory(repoConfig, c.GetJwt())
		if err != nil {
			return nil, err
		}
		d.AppendClose("loginRepo", close)

		return &loginRepository{repo: repoImpl}, nil
	}
}

func (l *loginRepository) Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (string, error) {
	redirectURL, err := l.repo.Login(ctx, req)
	if err != nil {
		klog.Context(ctx).Debugw("msg", "login failed", "error", err)
		return "", err
	}
	return redirectURL, nil
}
