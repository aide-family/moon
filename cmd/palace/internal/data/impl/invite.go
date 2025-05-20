package impl

import (
	"context"
	_ "embed"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/template"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewInviteRepo(
	bc *conf.Bootstrap,
	d *data.Data,
	logger log.Logger,
) repository.Invite {
	return &inviteRepoImpl{
		bc:     bc,
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.invite")),
	}
}

type inviteRepoImpl struct {
	bc *conf.Bootstrap
	*data.Data
	helper *log.Helper
}

//go:embed template/team_invite_user.html
var inviteEmailBody string

func (i *inviteRepoImpl) TeamInviteUser(ctx context.Context, req bo.InviteMember) error {
	roleIds := slices.MapFilter(req.GetRoles(), func(role do.TeamRole) (uint32, bool) {
		if validate.IsNil(role) || role.GetID() <= 0 {
			return 0, false
		}
		return role.GetID(), true
	})
	teamInviteUserDo := &system.TeamInviteUser{
		TeamID:       req.GetTeam().GetID(),
		InviteUserID: req.GetInviteUser().GetID(),
		Position:     req.GetPosition(),
		Roles:        roleIds,
	}
	mainMutation := getMainQuery(ctx, i)
	teamInviteUserMutation := mainMutation.TeamInviteUser
	teamInviteUserDo.WithContext(ctx)
	if err := teamInviteUserMutation.WithContext(ctx).Create(teamInviteUserDo); err != nil {
		return err
	}
	sendEmail := req.GetSendEmailFun()
	templateParams := map[string]string{
		"RedirectURI": i.bc.GetAuth().GetOauth2().GetRedirectUri(),
		"InviterName": req.GetOperator().GetMemberName(),
		"Position":    req.GetPosition().String(),
		"TeamName":    req.GetTeam().GetName(),
		"Roles":       strings.Join(slices.Map(req.GetRoles(), func(role do.TeamRole) string { return role.GetName() }), ","),
	}
	body, err := template.HtmlFormatter(inviteEmailBody, templateParams)
	if err != nil {
		return err
	}

	sendParams := &bo.SendEmailParams{
		Email:       string(req.GetInviteUser().GetEmail()),
		Body:        body,
		Subject:     "Welcome to the Moon Monitoring System.",
		ContentType: "text/html",
		RequestID:   uuid.New().String(),
	}
	return sendEmail(ctx, sendParams)
}
