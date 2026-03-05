package impl

import (
	"context"
	"net/url"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/jwt"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
	"github.com/asaskevich/govalidator"
	"github.com/go-kratos/kratos/v2/errors"
	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/goddess/internal/data"
	"github.com/aide-family/goddess/internal/data/impl/convert"
	"github.com/aide-family/goddess/internal/data/impl/do"
	"github.com/aide-family/goddess/internal/data/impl/query"
)

func NewLoginRepository(bc *conf.Bootstrap, d *data.Data) (repository.LoginRepository, error) {
	return NewLoginRepositoryWithDB(d.DB(), bc.GetJwt()), nil
}

func NewLoginRepositoryWithDB(db *gorm.DB, jwtConfig *config.JWT) repository.LoginRepository {
	query.SetDefault(db)
	return &loginRepository{db: db, jwtConfig: jwtConfig}
}

type loginRepository struct {
	db        *gorm.DB
	jwtConfig *config.JWT
}

func (g *loginRepository) LoginByEmail(ctx context.Context, email string) (string, error) {
	user, err := g.findOrCreateUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	return g.generateToken(user)
}

func (g *loginRepository) findOrCreateUserByEmail(ctx context.Context, email string) (*do.User, error) {
	userMutation := query.Use(getDBWithTransaction(ctx, g.db)).User
	userDO, err := userMutation.WithContext(ctx).Where(userMutation.Email.Eq(email)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorInternalServer("find user by email failed").WithCause(err)
		}
		userDO = &do.User{
			Email:  email,
			Name:   email,
			Status: enum.UserStatus_ACTIVE,
		}
		if err := userMutation.WithContext(ctx).Create(userDO); err != nil {
			return nil, merr.ErrorInternalServer("create user by email failed").WithCause(err)
		}
	}
	return userDO, nil
}

// Login implements [authv1.Repository].
func (g *loginRepository) LoginByOAuth2(ctx context.Context, req *bo.OAuth2LoginBo) (string, error) {
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
	if !govalidator.IsEmail(user.Email) {
		klog.Context(ctx).Debugw("msg", "email is invalid", "email", user.Email)
		return "", merr.ErrorInvalidArgument("email is invalid")
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

func (g *loginRepository) oauth2UserEqual(oauth2UserDO *do.OAuth2User, user *bo.OAuth2UserBo) bool {
	return oauth2UserDO.Email == user.Email &&
		oauth2UserDO.Name == user.Name &&
		oauth2UserDO.Avatar == user.Avatar &&
		oauth2UserDO.Remark == user.Remark &&
		oauth2UserDO.Nickname == user.Nickname
}

func (g *loginRepository) userEqual(userDO *do.User, user *do.OAuth2User) bool {
	return userDO.Email == user.Email &&
		userDO.Name == user.Name &&
		userDO.Avatar == user.Avatar &&
		userDO.Remark == user.Remark &&
		userDO.Nickname == user.Nickname
}

func (g *loginRepository) findOrCreateOAuth2User(ctx context.Context, user *bo.OAuth2UserBo) (*do.OAuth2User, error) {
	oauth2Mutation := query.Use(getDBWithTransaction(ctx, g.db)).OAuth2User
	oauth2UserDO, err := oauth2Mutation.WithContext(ctx).Where(oauth2Mutation.OpenID.Eq(user.OpenID), oauth2Mutation.APP.Eq(user.App.String())).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			klog.Context(ctx).Debugw("msg", "get oauth2 user failed", "error", err, "openID", user.OpenID)
			return nil, merr.ErrorInternalServer("get oauth2 user failed").WithCause(err)
		}
		oauth2UserDO = convert.OAuth2UserToDo(user)
		if err := oauth2Mutation.WithContext(ctx).Create(oauth2UserDO); err != nil {
			klog.Context(ctx).Debugw("msg", "create oauth2 user failed", "error", err, "oauth2UserUID", oauth2UserDO.UID)
			return nil, merr.ErrorInternalServer("create oauth2 user failed").WithCause(err)
		}
	}
	if g.oauth2UserEqual(oauth2UserDO, user) {
		return oauth2UserDO, nil
	}
	updateColumns := []field.AssignExpr{
		oauth2Mutation.Email.Value(user.Email),
		oauth2Mutation.Name.Value(user.Name),
		oauth2Mutation.Avatar.Value(user.Avatar),
		oauth2Mutation.Remark.Value(user.Remark),
		oauth2Mutation.Nickname.Value(user.Nickname),
		oauth2Mutation.Raw.Value(user.Raw),
	}
	_, err = oauth2Mutation.WithContext(ctx).Where(oauth2Mutation.UID.Eq(int64(oauth2UserDO.UID))).UpdateColumnSimple(updateColumns...)
	if err != nil {
		klog.Context(ctx).Debugw("msg", "update oauth2 user email failed", "error", err, "oauth2UserUID", oauth2UserDO.UID, "email", oauth2UserDO.Email)
		return nil, merr.ErrorInternalServer("update oauth2 user email failed").WithCause(err)
	}
	return oauth2UserDO, nil
}

func (g *loginRepository) findOrCreateUser(ctx context.Context, user *do.OAuth2User) (*do.User, error) {
	userMutation := query.Use(getDBWithTransaction(ctx, g.db)).User
	var userDO *do.User
	var err error
	userDO, err = userMutation.WithContext(ctx).Where(userMutation.Email.Eq(user.Email)).First()
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
	if g.userEqual(userDO, user) {
		return userDO, nil
	}
	updateColumns := []field.AssignExpr{
		userMutation.Name.Value(user.Name),
		userMutation.Nickname.Value(user.Nickname),
		userMutation.Avatar.Value(user.Avatar),
		userMutation.Remark.Value(user.Remark),
	}
	_, err = userMutation.WithContext(ctx).Where(userMutation.UID.Eq(int64(userDO.UID))).UpdateColumnSimple(updateColumns...)
	if err != nil {
		klog.Context(ctx).Warnw("msg", "update user failed", "error", err, "userUID", userDO.UID)
	}
	return userDO, nil
}

func (g *loginRepository) bindUserAndOAuth2User(ctx context.Context, user *do.User, oauth2User *do.OAuth2User) error {
	if int64(oauth2User.UID) == int64(user.UID) {
		return nil
	}
	oauth2Mutation := query.Use(getDBWithTransaction(ctx, g.db)).OAuth2User
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
	if strutil.IsEmpty(redirectURL) {
		return "", merr.ErrorInvalidArgument("redirect URL is empty")
	}
	urlObj, err := url.Parse(redirectURL)
	if err != nil {
		return "", merr.ErrorInvalidArgument("invalid redirect URL").WithCause(err)
	}
	if urlObj.Scheme != "" && urlObj.Scheme != "https" && urlObj.Scheme != "http" {
		return "", merr.ErrorInvalidArgument("redirect URL scheme not allowed")
	}
	query := urlObj.Query()
	query.Set("token", token)
	urlObj.RawQuery = query.Encode()
	return urlObj.String(), nil
}
