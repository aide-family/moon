package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
)

type repoBase struct {
	client *ent.Client
	log    *log.Helper
}

func newBase(d *Data, logger log.Logger) *repoBase {
	return &repoBase{client: d.client, log: log.NewHelper(logger)}
}

type userRepo struct{ *repoBase }
type captchaRepo struct{ *repoBase }
type emailCodeRepo struct{ *repoBase }
type oauthRepo struct{ *repoBase }
type namespaceRepo struct{ *repoBase }
type memberRepo struct{ *repoBase }
type roleRepo struct{ *repoBase }

func NewUserRepo(d *Data, logger log.Logger) repository.UserRepo {
	return &userRepo{newBase(d, logger)}
}

func NewCaptchaRepo(d *Data, logger log.Logger) repository.CaptchaRepo {
	return &captchaRepo{newBase(d, logger)}
}

func NewEmailCodeRepo(d *Data, logger log.Logger) repository.EmailCodeRepo {
	return &emailCodeRepo{newBase(d, logger)}
}

func NewOAuthRepo(d *Data, logger log.Logger) repository.OAuthRepo {
	return &oauthRepo{newBase(d, logger)}
}

func NewNamespaceRepo(d *Data, logger log.Logger) repository.NamespaceRepo {
	return &namespaceRepo{newBase(d, logger)}
}

func NewMemberRepo(d *Data, logger log.Logger) repository.MemberRepo {
	return &memberRepo{newBase(d, logger)}
}

func NewRoleRepo(d *Data, logger log.Logger) repository.RoleRepo {
	return &roleRepo{newBase(d, logger)}
}
