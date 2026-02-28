package biz

import (
	"context"

	magicboxapiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewRecipientGroup(
	memberRepo repository.Member,
	recipientGroupRepo repository.RecipientGroup,
	recipientMemberRepo repository.RecipientMember,
	helper *klog.Helper,
) *RecipientGroup {
	return &RecipientGroup{
		memberRepo:          memberRepo,
		recipientGroupRepo:  recipientGroupRepo,
		recipientMemberRepo: recipientMemberRepo,
		helper:              klog.NewHelper(klog.With(helper.Logger(), "biz", "recipient_group")),
	}
}

type RecipientGroup struct {
	helper              *klog.Helper
	memberRepo          repository.Member
	recipientGroupRepo  repository.RecipientGroup
	recipientMemberRepo repository.RecipientMember
}

func (b *RecipientGroup) CreateRecipientGroup(ctx context.Context, req *bo.CreateRecipientGroupBo) (snowflake.ID, error) {
	if recipientGroup, err := b.recipientGroupRepo.GetRecipientGroupByName(ctx, req.Name); err == nil {
		return 0, merr.ErrorParams("recipient group %s already exists, uid: %s", req.Name, recipientGroup.UID)
	} else if !merr.IsNotFound(err) {
		b.helper.Errorw("msg", "check recipient group exists failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create recipient group failed").WithCause(err)
	}
	if len(req.Members) > 0 {
		if err := b.createRecipientMember(ctx, req.Members); err != nil {
			b.helper.Errorw("msg", "create recipient member failed", "error", err)
			return 0, merr.ErrorInternalServer("create recipient member failed").WithCause(err)
		}
	}
	uid, err := b.recipientGroupRepo.CreateRecipientGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "create recipient group failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create recipient group failed").WithCause(err)
	}
	return uid, nil
}

func (b *RecipientGroup) createRecipientMember(ctx context.Context, memberUIDs []int64) error {
	membersReq := &magicboxapiv1.ListMemberRequest{
		Page:     1,
		PageSize: 200,
		Status:   enum.MemberStatus_JOINED,
		Uids:     memberUIDs,
	}
	members, err := b.memberRepo.ListMember(ctx, membersReq)
	if err != nil {
		b.helper.Errorw("msg", "list member failed", "error", err)
		return merr.ErrorInternalServer("list member failed").WithCause(err)
	}

	if err := b.recipientMemberRepo.CreateRecipientMember(ctx, bo.NewRecipientMemberItemBo(members.Items)); err != nil {
		b.helper.Errorw("msg", "create recipient member failed", "error", err)
		return merr.ErrorInternalServer("create recipient member failed").WithCause(err)
	}
	return nil
}

func (b *RecipientGroup) GetRecipientGroup(ctx context.Context, uid snowflake.ID) (*bo.RecipientGroupDetailBo, error) {
	detail, err := b.recipientGroupRepo.GetRecipientGroup(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, err
		}
		b.helper.Errorw("msg", "get recipient group failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get recipient group failed").WithCause(err)
	}
	return detail, nil
}

func (b *RecipientGroup) UpdateRecipientGroup(ctx context.Context, req *bo.UpdateRecipientGroupBo) error {
	recipientGroup, err := b.recipientGroupRepo.GetRecipientGroup(ctx, req.UID)
	if err != nil && !merr.IsNotFound(err) {
		b.helper.Errorw("msg", "get recipient group failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("get recipient group failed").WithCause(err)
	} else if recipientGroup != nil && recipientGroup.Name != req.Name {
		return merr.ErrorParams("recipient group %s already exists", req.Name)
	}
	if err := b.recipientGroupRepo.UpdateRecipientGroup(ctx, req); err != nil {
		b.helper.Errorw("msg", "update recipient group failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update recipient group failed").WithCause(err)
	}
	return nil
}

func (b *RecipientGroup) UpdateRecipientGroupStatus(ctx context.Context, req *bo.UpdateRecipientGroupStatusBo) error {
	if err := b.recipientGroupRepo.UpdateRecipientGroupStatus(ctx, req); err != nil {
		b.helper.Errorw("msg", "update recipient group status failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update recipient group status failed").WithCause(err)
	}
	return nil
}

func (b *RecipientGroup) DeleteRecipientGroup(ctx context.Context, uid snowflake.ID) error {
	if err := b.recipientGroupRepo.DeleteRecipientGroup(ctx, uid); err != nil {
		b.helper.Errorw("msg", "delete recipient group failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete recipient group failed").WithCause(err)
	}
	return nil
}

func (b *RecipientGroup) ListRecipientGroup(ctx context.Context, req *bo.ListRecipientGroupBo) (*bo.PageResponseBo[*bo.RecipientGroupItemBo], error) {
	page, err := b.recipientGroupRepo.ListRecipientGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "list recipient group failed", "error", err)
		return nil, merr.ErrorInternalServer("list recipient group failed").WithCause(err)
	}
	return page, nil
}

func (b *RecipientGroup) SelectRecipientGroup(ctx context.Context, req *bo.SelectRecipientGroupBo) (*bo.SelectRecipientGroupBoResult, error) {
	result, err := b.recipientGroupRepo.SelectRecipientGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "select recipient group failed", "error", err)
		return nil, merr.ErrorInternalServer("select recipient group failed").WithCause(err)
	}
	return result, nil
}
