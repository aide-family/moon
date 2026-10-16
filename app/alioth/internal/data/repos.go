package data

import (
	"database/sql"
	"log/slog"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
)

type repoBase struct {
	client *ent.Client
	db     *sql.DB
	box    *secretBox
	log    *slog.Logger
}

func newBase(d *Data, logger *slog.Logger) *repoBase {
	return &repoBase{client: d.client, db: d.db, box: d.box, log: logger}
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
type alertPageRepo struct{ *repoBase }
type notificationRepo struct{ *repoBase }
type agentRepo struct{ *repoBase }
type sshCommandRepo struct{ *repoBase }

func NewUserRepo(d *Data, logger *slog.Logger) repository.UserRepo {
	return &userRepo{newBase(d, logger)}
}

func NewCaptchaRepo(d *Data, logger *slog.Logger) repository.CaptchaRepo {
	return &captchaRepo{newBase(d, logger)}
}

func NewEmailCodeRepo(d *Data, logger *slog.Logger) repository.EmailCodeRepo {
	return &emailCodeRepo{newBase(d, logger)}
}

func NewOAuthRepo(d *Data, logger *slog.Logger) repository.OAuthRepo {
	return &oauthRepo{newBase(d, logger)}
}

func NewNamespaceRepo(d *Data, logger *slog.Logger) repository.NamespaceRepo {
	return &namespaceRepo{newBase(d, logger)}
}

func NewMemberRepo(d *Data, logger *slog.Logger) repository.MemberRepo {
	return &memberRepo{newBase(d, logger)}
}

func NewRoleRepo(d *Data, logger *slog.Logger) repository.RoleRepo {
	return &roleRepo{newBase(d, logger)}
}

func NewDatasourceRepo(d *Data, logger *slog.Logger) repository.DatasourceRepo {
	return &datasourceRepo{newBase(d, logger)}
}

func NewAlertLevelRepo(d *Data, logger *slog.Logger) repository.AlertLevelRepo {
	return &alertLevelRepo{newBase(d, logger)}
}

func NewAlertRuleRepo(d *Data, logger *slog.Logger) repository.AlertRuleRepo {
	return &alertRuleRepo{newBase(d, logger)}
}

func NewAlertEventRepo(d *Data, logger *slog.Logger) repository.AlertEventRepo {
	return &alertEventRepo{newBase(d, logger)}
}

func NewAlertPageRepo(d *Data, logger *slog.Logger) repository.AlertPageRepo {
	return &alertPageRepo{newBase(d, logger)}
}

func NewNotificationRepo(d *Data, logger *slog.Logger) repository.NotificationRepo {
	return &notificationRepo{newBase(d, logger)}
}

func NewAgentRepo(d *Data, logger *slog.Logger) repository.AgentRepo {
	return &agentRepo{newBase(d, logger)}
}

func NewSSHCommandRepo(d *Data, logger *slog.Logger) repository.SSHCommandRepo {
	return &sshCommandRepo{newBase(d, logger)}
}
