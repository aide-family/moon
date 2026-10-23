package data

import (
	"database/sql"
	"log/slog"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
)

type repoBase struct {
	client      *ent.Client
	db          *sql.DB
	box         *secretBox
	log         *slog.Logger
	history     *historyStore
	messageLogs *messageLogStore
}

func newBase(deps *Dependencies, logger *slog.Logger) *repoBase {
	return &repoBase{
		client: deps.client, db: deps.db, box: deps.box, log: logger,
		history: deps.history, messageLogs: deps.messageLogs,
	}
}

type userRepo struct{ *repoBase }
type captchaRepo struct{ *repoBase }
type emailCodeRepo struct{ *repoBase }
type oauthRepo struct{ *repoBase }
type namespaceRepo struct{ *repoBase }
type memberRepo struct{ *repoBase }
type messageRepo struct{ *repoBase }
type roleRepo struct{ *repoBase }
type datasourceRepo struct{ *repoBase }
type alertDictionaryRepo struct{ *repoBase }
type alertRuleRepo struct{ *repoBase }
type alertEventRepo struct{ *repoBase }
type alertPageRepo struct{ *repoBase }
type notificationRepo struct{ *repoBase }
type agentRepo struct{ *repoBase }
type sshCommandRepo struct{ *repoBase }

func NewUserRepo(deps *Dependencies, logger *slog.Logger) repository.UserRepo {
	return &userRepo{newBase(deps, logger)}
}

func NewCaptchaRepo(deps *Dependencies, logger *slog.Logger) repository.CaptchaRepo {
	return &captchaRepo{newBase(deps, logger)}
}

func NewEmailCodeRepo(deps *Dependencies, logger *slog.Logger) repository.EmailCodeRepo {
	return &emailCodeRepo{newBase(deps, logger)}
}

func NewOAuthRepo(deps *Dependencies, logger *slog.Logger) repository.OAuthRepo {
	return &oauthRepo{newBase(deps, logger)}
}

func NewNamespaceRepo(deps *Dependencies, logger *slog.Logger) repository.NamespaceRepo {
	return &namespaceRepo{newBase(deps, logger)}
}

func NewMemberRepo(deps *Dependencies, logger *slog.Logger) repository.MemberRepo {
	return &memberRepo{newBase(deps, logger)}
}

func NewInviteRepo(deps *Dependencies, logger *slog.Logger) repository.InviteRepo {
	return &inviteRepo{newBase(deps, logger)}
}

func NewMessageRepo(deps *Dependencies, logger *slog.Logger) repository.MessageRepo {
	return &messageRepo{newBase(deps, logger)}
}

func NewRoleRepo(deps *Dependencies, logger *slog.Logger) repository.RoleRepo {
	return &roleRepo{newBase(deps, logger)}
}

func NewDatasourceRepo(deps *Dependencies, logger *slog.Logger) repository.DatasourceRepo {
	return &datasourceRepo{newBase(deps, logger)}
}

func NewAlertDictionaryRepo(deps *Dependencies, logger *slog.Logger) repository.AlertDictionaryRepo {
	return &alertDictionaryRepo{newBase(deps, logger)}
}

func NewAlertRuleRepo(deps *Dependencies, logger *slog.Logger) repository.AlertRuleRepo {
	return &alertRuleRepo{newBase(deps, logger)}
}

func NewAlertEventRepo(deps *Dependencies, logger *slog.Logger) repository.AlertEventRepo {
	return &alertEventRepo{newBase(deps, logger)}
}

func NewAlertPageRepo(deps *Dependencies, logger *slog.Logger) repository.AlertPageRepo {
	return &alertPageRepo{newBase(deps, logger)}
}

func NewNotificationRepo(deps *Dependencies, logger *slog.Logger) repository.NotificationRepo {
	return &notificationRepo{newBase(deps, logger)}
}

func NewAgentRepo(deps *Dependencies, logger *slog.Logger) repository.AgentRepo {
	return &agentRepo{newBase(deps, logger)}
}

func NewSSHCommandRepo(deps *Dependencies, logger *slog.Logger) repository.SSHCommandRepo {
	return &sshCommandRepo{newBase(deps, logger)}
}
