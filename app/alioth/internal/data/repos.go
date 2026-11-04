package data

import (
	"database/sql"
	"log/slog"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
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

// probeRepo 管理探测目标组、组内地址与探测记录。
type probeRepo struct{ *repoBase }
type alertDictionaryRepo struct{ *repoBase }
type alertRuleRepo struct{ *repoBase }
type alertEventRepo struct{ *repoBase }
type alertPageRepo struct{ *repoBase }
type notificationRepo struct{ *repoBase }
type agentRepo struct{ *repoBase }
type sshCommandRepo struct{ *repoBase }
type featureRepo struct{ *repoBase }
type boardRepo struct{ *repoBase }
type alertMaintenanceRepo struct{ *repoBase }

// alertEscalationRepo 管理升级策略、步骤、事件升级记录与调度生成。
type alertEscalationRepo struct{ *repoBase }
type statusPageRepo struct{ *repoBase }
type probeSlaRepo struct{ *repoBase }

// siteRepo 维护官网的站点设置单例。
type siteRepo struct{ *repoBase }

// loginMethodRepo 维护登录方式开关单例。
type loginMethodRepo struct{ *repoBase }

// sitePageRepo 维护官网页面与发布快照。
type sitePageRepo struct{ *repoBase }

// emailSettingRepo 维护登录与邀请邮件的 SMTP 配置单例。
type emailSettingRepo struct{ *repoBase }

// oauthProviderSettingRepo 维护第三方登录来源的后台配置。
type oauthProviderSettingRepo struct{ *repoBase }

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

// NewProbeRepo 返回探测目标组的仓储实现。
func NewProbeRepo(deps *Dependencies, logger *slog.Logger) repository.ProbeTargetRepo {
	return &probeRepo{newBase(deps, logger)}
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

func NewStatusPageRepo(deps *Dependencies, logger *slog.Logger) repository.StatusPageRepo {
	return &statusPageRepo{newBase(deps, logger)}
}

func NewProbeSlaRepo(deps *Dependencies, logger *slog.Logger) repository.ProbeSlaRepo {
	return &probeSlaRepo{newBase(deps, logger)}
}

func NewSiteRepo(deps *Dependencies, logger *slog.Logger) repository.SiteRepo {
	return &siteRepo{newBase(deps, logger)}
}

func NewLoginMethodRepo(deps *Dependencies, logger *slog.Logger) repository.LoginMethodRepo {
	return &loginMethodRepo{newBase(deps, logger)}
}

func NewSitePageRepo(deps *Dependencies, logger *slog.Logger) repository.SitePageRepo {
	return &sitePageRepo{newBase(deps, logger)}
}

func NewEmailSettingRepo(deps *Dependencies, logger *slog.Logger) repository.EmailSettingRepo {
	return &emailSettingRepo{newBase(deps, logger)}
}

func NewOAuthProviderSettingRepo(deps *Dependencies, logger *slog.Logger) repository.OAuthProviderSettingRepo {
	return &oauthProviderSettingRepo{newBase(deps, logger)}
}

func NewAlertMaintenanceRepo(deps *Dependencies, logger *slog.Logger) repository.AlertMaintenanceRepo {
	return &alertMaintenanceRepo{newBase(deps, logger)}
}

func NewAlertEscalationRepo(deps *Dependencies, logger *slog.Logger) repository.AlertEscalationRepo {
	return &alertEscalationRepo{newBase(deps, logger)}
}

func NewOnCallRotationRepo(deps *Dependencies, logger *slog.Logger) repository.OnCallRotationRepo {
	return &onCallRotationRepo{newBase(deps, logger)}
}

func NewBoardRepo(deps *Dependencies, logger *slog.Logger) repository.BoardRepo {
	return &boardRepo{newBase(deps, logger)}
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

// NewFeatureRepo 返回功能套餐和租户功能差异的仓储实现。
func NewFeatureRepo(deps *Dependencies, logger *slog.Logger) repository.FeatureRepo {
	return &featureRepo{newBase(deps, logger)}
}

// NewNamespaceOrderRepo 返回租户创建订单的仓储实现。
func NewNamespaceOrderRepo(deps *Dependencies, logger *slog.Logger) repository.NamespaceOrderRepo {
	return &namespaceOrderRepo{newBase(deps, logger)}
}
