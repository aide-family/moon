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

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func NewGithubUserRepository(bc *palaceconf.Bootstrap, data *data.Data) repository.GithubUser {
	return &githubUserRepositoryImpl{data: data, bc: bc}
}

type githubUserRepositoryImpl struct {
	data *data.Data
	bc   *palaceconf.Bootstrap
}

func (g *githubUserRepositoryImpl) FirstOrCreate(ctx context.Context, user *auth.GithubLoginResponse) (*model.SysUser, error) {
	userQuery := query.Use(g.data.GetMainDB(ctx)).SysGithubUser
	first, err := userQuery.WithContext(ctx).Where(userQuery.GithubUserID.Eq(user.Id)).First()
	if types.IsNil(err) {
		return g.getSysUserByID(ctx, first.SysUserID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	randPass := cipher.MD5(time.Now().String())[:8]
	password := types.NewPassword(cipher.MD5(randPass + "3c4d9a0a5a703938dd1d2d46e1c924f9"))
	sysUser := &model.SysUser{
		Username: user.Login,
		Nickname: user.Name,
		Password: password.String(),
		Email:    user.Email,
		Phone:    user.Login,
		Remark:   user.Bio,
		Avatar:   user.AvatarUrl,
		Salt:     password.GetSalt(),
		Gender:   vobj.GenderUnknown,
		Role:     vobj.RoleUser,
		Status:   vobj.StatusEnable,
	}
	// 调试用
	//sysUser.Email = "1058165620@qq.com"
	err = query.Use(g.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 创建系统用户
		if err = query.Use(g.data.GetMainDB(ctx)).SysUser.Create(sysUser); !types.IsNil(err) {
			return err
		}
		sysGitHubUser := &model.SysGithubUser{
			GithubUserID: user.Id,
			SysUserID:    sysUser.ID,
			Row:          user.String(),
		}
		if err = userQuery.WithContext(ctx).Create(sysGitHubUser); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return nil, err
	}
	// 发送用户密码到用户邮箱
	return sysUser, g.sendUserPassword(ctx, sysUser, randPass)
}

func (g *githubUserRepositoryImpl) getSysUserByID(ctx context.Context, id uint32) (*model.SysUser, error) {
	userQuery := query.Use(g.data.GetMainDB(ctx)).SysUser
	return userQuery.WithContext(ctx).Where(userQuery.ID.Eq(id)).First()
}

//go:embed welcome.html
var body string

func (g *githubUserRepositoryImpl) sendUserPassword(_ context.Context, user *model.SysUser, pass string) error {
	if types.TextIsNull(user.Email) {
		return nil
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
