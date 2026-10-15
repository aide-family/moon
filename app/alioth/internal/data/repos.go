package data

import (
	"database/sql"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
)

type repoBase struct {
	client *ent.Client
	db     *sql.DB
	box    *secretBox
	log    *log.Helper
}

func newBase(d *Data, logger log.Logger) *repoBase {
	return &repoBase{client: d.client, db: d.db, box: d.box, log: log.NewHelper(logger)}
}

type userRepo struct{ *repoBase }
type captchaRepo struct{ *repoBase }
type emailCodeRepo struct{ *repoBase }
type oauthRepo struct{ *repoBase }
type namespaceRepo struct{ *repoBase }
type memberRepo struct{ *repoBase }
type roleRepo struct{ *repoBase }
type datasourceRepo struct{ *repoBase }
type alertLevelRepo struct{ *repoBase }
type alertRuleRepo struct{ *repoBase }
type alertEventRepo struct{ *repoBase }

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

func NewDatasourceRepo(d *Data, logger log.Logger) repository.DatasourceRepo {
	return &datasourceRepo{newBase(d, logger)}
}

func NewAlertLevelRepo(d *Data, logger log.Logger) repository.AlertLevelRepo {
	return &alertLevelRepo{newBase(d, logger)}
}

func NewAlertRuleRepo(d *Data, logger log.Logger) repository.AlertRuleRepo {
	return &alertRuleRepo{newBase(d, logger)}
}

func NewAlertEventRepo(d *Data, logger log.Logger) repository.AlertEventRepo {
	return &alertEventRepo{newBase(d, logger)}
}
