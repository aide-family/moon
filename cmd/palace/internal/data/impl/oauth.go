package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
)

func NewOAuthRepo(data *data.Data, logger log.Logger) repository.OAuth {
	return &oauthRepoImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.oauth")),
	}
}

type oauthRepoImpl struct {
	*data.Data
	helper *log.Helper
}

func toOAuthUserDo(u bo.IOAuthUser) *system.UserOAuth {
	if u == nil {
		return nil
	}
	return &system.UserOAuth{
		OpenID: u.GetOpenID(),
		UserID: u.GetUserID(),
		Row:    u.String(),
		APP:    u.GetAPP(),
		User:   nil,
	}
}

func (o *oauthRepoImpl) Create(ctx context.Context, user bo.IOAuthUser) (do.UserOAuth, error) {
	oauthUserDo := toOAuthUserDo(user)
	mainQuery := getMainQuery(ctx, o)
	oauthUserMutation := mainQuery.UserOAuth
	if err := oauthUserMutation.WithContext(ctx).Create(oauthUserDo); err != nil {
		return nil, err
	}
	return oauthUserMutation.WithContext(ctx).
		Where(oauthUserMutation.OpenID.Eq(oauthUserDo.OpenID)).
		Where(oauthUserMutation.APP.Eq(oauthUserDo.APP.GetValue())).
		Preload(oauthUserMutation.User).
		First()
}

func (o *oauthRepoImpl) FindByOpenID(ctx context.Context, openID string, app vobj.OAuthAPP) (do.UserOAuth, error) {
	mainQuery := getMainQuery(ctx, o)
	oauthUserMutation := mainQuery.UserOAuth
	oauthUserDo, err := oauthUserMutation.WithContext(ctx).
		Where(oauthUserMutation.OpenID.Eq(openID)).
		Where(oauthUserMutation.APP.Eq(app.GetValue())).
		Preload(oauthUserMutation.User).
		First()
	if err != nil {
		return nil, oauthUserNotFound(err)
	}
	return oauthUserDo, nil
}

func (o *oauthRepoImpl) SetUser(ctx context.Context, user do.UserOAuth) (do.UserOAuth, error) {
	mainQuery := getMainQuery(ctx, o)
	oauthUserMutation := mainQuery.UserOAuth
	wrapper := []gen.Condition{
		oauthUserMutation.ID.Eq(user.GetID()),
		oauthUserMutation.APP.Eq(user.GetAPP().GetValue()),
		oauthUserMutation.OpenID.Eq(user.GetOpenID()),
	}

	if _, err := oauthUserMutation.WithContext(ctx).Where(wrapper...).UpdateSimple(oauthUserMutation.UserID.Value(user.GetUserID())); err != nil {
		return nil, err
	}
	return oauthUserMutation.WithContext(ctx).Where(wrapper...).Preload(oauthUserMutation.User).First()
}
