package repoimpl

import (
	"context"
	_ "embed"
	"time"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo/auth"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/cipher"
	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gorm/clause"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func NewGithubUserRepository(bc *palaceconf.Bootstrap, data *data.Data) repository.OAuth {
	return &githubUserRepositoryImpl{data: data, bc: bc}
}

type githubUserRepositoryImpl struct {
	data *data.Data
	bc   *palaceconf.Bootstrap
}

func buildSysUserModel(u auth.IOAuthUser, pass types.Password) *model.SysUser {
	return &model.SysUser{
		Username: u.GetUsername(),
		Nickname: u.GetNickname(),
		Password: pass.String(),
		Email:    u.GetEmail(),
		Remark:   u.GetRemark(),
		Avatar:   u.GetAvatar(),
		Salt:     pass.GetSalt(),
		Gender:   vobj.GenderUnknown,
		Role:     vobj.RoleUser,
		Status:   vobj.StatusEnable,
	}
}

func genPassword() (string, types.Password) {
	randPass := cipher.MD5(time.Now().String())[:8]
	password := types.NewPassword(cipher.MD5(randPass + "3c4d9a0a5a703938dd1d2d46e1c924f9"))
	return randPass, password
}

func (g *githubUserRepositoryImpl) OAuthUserFirstOrCreate(ctx context.Context, user auth.IOAuthUser) (sysUser *model.SysUser, err error) {
	userQuery := query.Use(g.data.GetMainDB(ctx)).SysOAuthUser
	first, err := userQuery.WithContext(ctx).Where(userQuery.OAuthID.Eq(user.GetOAuthID()), userQuery.APP.Eq(user.GetAPP().GetValue())).First()
	if types.IsNil(err) {
		return g.getSysUserByID(ctx, first.SysUserID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err = types.CheckEmail(user.GetEmail()); err != nil {
		return nil, err
	}

	// 根据邮箱查询系统用户
	sysUser, err = g.getSysUserByEmail(ctx, user.GetEmail())
	if !types.IsNil(err) {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		randPass, password := genPassword()
		sysUser = buildSysUserModel(user, password)
		defer func() {
			if err == nil {
				// 发送用户密码到用户邮箱
				g.sendUserPassword(ctx, sysUser, randPass)
			}
		}()
	}
	// 调试用
	//sysUser.Email = "1058165620@qq.com"
	err = query.Use(g.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 创建系统用户
		if err = query.Use(g.data.GetMainDB(ctx)).SysUser.Clauses(clause.OnConflict{DoNothing: true}).Create(sysUser); !types.IsNil(err) {
			return err
		}
		sysOAuthUser := &model.SysOAuthUser{
			OAuthID:   user.GetOAuthID(),
			SysUserID: sysUser.ID,
			Row:       user.String(),
			APP:       user.GetAPP(),
		}
		if err = userQuery.WithContext(ctx).Create(sysOAuthUser); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return nil, err
	}
	return sysUser, nil
}

func (g *githubUserRepositoryImpl) getSysUserByID(ctx context.Context, id uint32) (*model.SysUser, error) {
	userQuery := query.Use(g.data.GetMainDB(ctx)).SysUser
	return userQuery.WithContext(ctx).Where(userQuery.ID.Eq(id)).First()
}

func (g *githubUserRepositoryImpl) getSysUserByEmail(ctx context.Context, email string) (*model.SysUser, error) {
	userQuery := query.Use(g.data.GetMainDB(ctx)).SysUser
	return userQuery.WithContext(ctx).Where(userQuery.Email.Eq(email)).First()
}

//go:embed welcome.html
var body string

func (g *githubUserRepositoryImpl) sendUserPassword(_ context.Context, user *model.SysUser, pass string) error {
	if err := types.CheckEmail(user.Email); err != nil {
		return err
	}

	body = format.Formatter(body, map[string]string{
		"Username":    user.Username,
		"Password":    pass,
		"RedirectURI": g.bc.GetRedirectUri(),
		"APP":         g.bc.GetServer().GetName(),
		"Remark":      g.bc.GetServer().GetMetadata()["description"],
	})
	// 发送用户密码到用户邮箱
	return g.data.GetEmailer().SetSubject("欢迎使用"+g.bc.GetServer().GetName()).SetTo(user.Email).SetBody(body, "text/html").Send()
}
