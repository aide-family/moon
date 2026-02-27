package impl

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/jwt"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/do"
	"github.com/aide-family/goddess/internal/data/impl/query"
)

type loginRepository struct {
	*data.Data
	jwtConfig *config.JWT
	node      *snowflake.Node
}

func NewLoginRepository(bc *conf.Bootstrap, d *data.Data) (repository.LoginRepository, error) {
	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		return nil, err
	}
	query.SetDefault(d.DB())
	return &loginRepository{Data: d, node: node, jwtConfig: bc.GetJwt()}, nil
}

// Login implements [authv1.Repository].
func (g *loginRepository) Login(ctx context.Context, req *bo.OAuth2LoginBo) (string, error) {
	user, oauthConfig := req.User, req.Config
	if pointer.IsNil(user) {
		klog.Context(ctx).Debugw("msg", "user is nil")
		return "", merr.ErrorInvalidArgument("user is nil")
	}
	if pointer.IsNil(oauthConfig) {
		klog.Context(ctx).Debugw("msg", "oauthConfig is nil")
		return "", merr.ErrorInvalidArgument("oauthConfig is nil")
	}
	if strutil.IsEmpty(user.Email) {
		klog.Context(ctx).Debugw("msg", "email is empty")
		return "", merr.ErrorInvalidArgument("email is empty")
	}

	// 1. check if outh2 user exists
	oauth2UserDO, err := g.findOrCreateOAuth2User(ctx, user)
	if err != nil {
		return "", err
	}

	// 2. check if user exists
	userDO, err := g.findOrCreateUser(ctx, oauth2UserDO)
	if err != nil {
		return "", err
	}

	// bind user and oauth2 user
	if err := g.bindUserAndOAuth2User(ctx, userDO, oauth2UserDO); err != nil {
		return "", err
	}

	// generate token
	token, err := g.generateToken(userDO)
	if err != nil {
		return "", err
	}

	// build redirect url
	redirectURL, err := g.buildRedirectURL(token, req.RedirectURL)
	if err != nil {
		return "", err
	}
	return redirectURL, nil
}

func (g *loginRepository) findOrCreateOAuth2User(ctx context.Context, user *bo.OAuth2UserBo) (*do.OAuth2User, error) {
	oauth2Mutation := query.OAuth2User
	oauth2UserDO, err := oauth2Mutation.WithContext(ctx).Where(oauth2Mutation.OpenID.Eq(user.OpenID), oauth2Mutation.APP.Eq(user.App.String())).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			klog.Context(ctx).Debugw("msg", "get oauth2 user failed", "error", err, "openID", user.OpenID)
			return nil, merr.ErrorInternalServer("get oauth2 user failed").WithCause(err)
		}
		oauth2UserDO = convert.OAuth2UserToDo(user)
		if err := oauth2Mutation.WithContext(ctx).Create(oauth2UserDO); err != nil {
			klog.Context(ctx).Debugw("msg", "create oauth2 user failed", "error", err, "oauth2UserUID", oauth2UserDO.UID)
			return nil, merr.ErrorInternalServer("create oauth2 user failed").WithCause(err).WithCause(err)
		}
	}
	if strings.EqualFold(user.Email, oauth2UserDO.Email) {
		return oauth2UserDO, nil
	}
	oauth2UserDO.Email = user.Email
	_, err = oauth2Mutation.WithContext(ctx).Where(oauth2Mutation.UID.Eq(int64(oauth2UserDO.UID))).UpdateColumn(oauth2Mutation.Email, oauth2UserDO.Email)
	if err != nil {
		klog.Context(ctx).Debugw("msg", "update oauth2 user email failed", "error", err, "oauth2UserUID", oauth2UserDO.UID, "email", oauth2UserDO.Email)
		return nil, merr.ErrorInternalServer("update oauth2 user email failed").WithCause(err)
	}
	return oauth2UserDO, nil
}

func (g *loginRepository) findOrCreateUser(ctx context.Context, user *do.OAuth2User) (*do.User, error) {
	userMutation := query.User
	var userDO *do.User
	var err error
	if user.UID > 0 {
		userDO, err = userMutation.WithContext(ctx).Where(userMutation.UID.Eq(int64(user.UID))).First()
	} else {
		userDO, err = userMutation.WithContext(ctx).Where(userMutation.Email.Eq(user.Email)).First()
	}
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			klog.Context(ctx).Debugw("msg", "get user failed", "error", err, "userUID", user.UID)
			return nil, merr.ErrorInternalServer("get user failed").WithCause(err)
		}
		userDO = convert.OAuth2UserToUserDo(user)
		if err := userMutation.WithContext(ctx).Create(userDO); err != nil {
			klog.Context(ctx).Debugw("msg", "create user failed", "error", err, "userUID", userDO.UID)
			return nil, merr.ErrorInternalServer("create user failed").WithCause(err)
		}
		return userDO, nil
	}
	if !strings.EqualFold(user.Email, userDO.Email) {
		return nil, merr.ErrorInvalidArgument("email mismatch")
	}
	return userDO, nil
}

func (g *loginRepository) bindUserAndOAuth2User(ctx context.Context, user *do.User, oauth2User *do.OAuth2User) error {
	if int64(oauth2User.UID) == int64(user.UID) {
		return nil
	}
	oauth2Mutation := query.OAuth2User
	if _, err := oauth2Mutation.WithContext(ctx).Where(oauth2Mutation.UID.Eq(int64(oauth2User.UID))).Update(oauth2Mutation.UID, int64(user.UID)); err != nil {
		klog.Context(ctx).Debugw("msg", "update oauth2 user failed", "error", err, "oauth2UserUID", oauth2User.UID, "userUID", user.UID)
		return merr.ErrorInternalServer("update oauth2 user failed").WithCause(err)
	}
	return nil
}

func (g *loginRepository) generateToken(user *do.User) (string, error) {
	claims := jwt.NewJwtClaims(g.jwtConfig, jwt.BaseInfo{
		UID:      user.UID,
		Username: user.Name,
	})
	return claims.GenerateToken()
}

// RefreshToken generates a new JWT from BaseInfo, reusing the same token generation logic as Login.
func (g *loginRepository) RefreshToken(ctx context.Context, baseInfo jwt.BaseInfo) (string, error) {
	claims := jwt.NewJwtClaims(g.jwtConfig, baseInfo)
	return claims.GenerateToken()
}

func (g *loginRepository) buildRedirectURL(token, redirectURL string) (string, error) {
	urlObj, err := url.Parse(redirectURL)
	if err != nil {
		return "", merr.ErrorInvalidArgument("invalid redirect URL").WithCause(err)
	}
	query := urlObj.Query()
	query.Set("token", token)
	urlObj.RawQuery = query.Encode()
	return urlObj.String(), nil
}
