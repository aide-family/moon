package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
)

func NewTeamNotice(
	transaction repository.Transaction,
	teamNoticeRepo repository.TeamNotice,
	teamHookRepo repository.TeamHook,
	teamMemberRepo repository.Member,
	teamConfigEmailRepo repository.TeamEmailConfig,
	teamConfigSMSRepo repository.TeamSMSConfig,
) *TeamNotice {
	return &TeamNotice{
		transaction:         transaction,
		teamNoticeRepo:      teamNoticeRepo,
		teamHookRepo:        teamHookRepo,
		teamMemberRepo:      teamMemberRepo,
		teamConfigEmailRepo: teamConfigEmailRepo,
		teamConfigSMSRepo:   teamConfigSMSRepo,
	}
}

type TeamNotice struct {
	transaction         repository.Transaction
	teamNoticeRepo      repository.TeamNotice
	teamHookRepo        repository.TeamHook
	teamMemberRepo      repository.Member
	teamConfigEmailRepo repository.TeamEmailConfig
	teamConfigSMSRepo   repository.TeamSMSConfig
}

func (t *TeamNotice) SaveNoticeGroup(ctx context.Context, req *bo.SaveNoticeGroupReq) error {
	hookDos, err := t.teamHookRepo.Find(ctx, req.HookIds)
	if err != nil {
		return err
	}
	req.WithHooks(hookDos)
	memberDos, err := t.teamMemberRepo.Find(ctx, req.GetMemberIds())
	if err != nil {
		return err
	}
	req.WithNoticeMembers(memberDos)
	if req.EmailConfigID > 0 {
		emailConfig, err := t.teamConfigEmailRepo.Get(ctx, req.EmailConfigID)
		if err != nil {
			return err
		}
		req.WithEmailConfig(emailConfig)
	}
	if req.SMSConfigID > 0 {
		smsConfig, err := t.teamConfigSMSRepo.Get(ctx, req.SMSConfigID)
		if err != nil {
			return err
		}
		req.WithSMSConfig(smsConfig)
	}
	return t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if req.GetID() <= 0 {
			return t.teamNoticeRepo.Create(ctx, req)
		}
		noticeGroupDo, err := t.teamNoticeRepo.Get(ctx, req.GetID())
		if err != nil {
			return err
		}
		req.WithNoticeGroup(noticeGroupDo)

		return t.teamNoticeRepo.Update(ctx, req.WithNoticeGroup(noticeGroupDo))
	})
}

func (t *TeamNotice) UpdateNoticeGroupStatus(ctx context.Context, req *bo.UpdateTeamNoticeGroupStatusRequest) error {
	if len(req.GroupIds) == 0 {
		return nil
	}
	return t.teamNoticeRepo.UpdateStatus(ctx, req)
}

func (t *TeamNotice) DeleteNoticeGroup(ctx context.Context, groupID uint32) error {
	return t.teamNoticeRepo.Delete(ctx, groupID)
}

func (t *TeamNotice) GetNoticeGroup(ctx context.Context, groupID uint32) (do.NoticeGroup, error) {
	return t.teamNoticeRepo.Get(ctx, groupID)
}

func (t *TeamNotice) TeamNoticeGroups(ctx context.Context, req *bo.ListNoticeGroupReq) (*bo.ListNoticeGroupReply, error) {
	return t.teamNoticeRepo.List(ctx, req)
}
