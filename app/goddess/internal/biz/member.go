package biz

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/biz/repository"
)

func NewMember(memberRepo repository.Member, userRepo repository.User, namespaceRepo repository.Namespace, helper *klog.Helper) *Member {
	return &Member{
		memberRepo:    memberRepo,
		userRepo:      userRepo,
		namespaceRepo: namespaceRepo,
		helper:        klog.NewHelper(klog.With(helper.Logger(), "biz", "member")),
	}
}

type Member struct {
	helper        *klog.Helper
	memberRepo    repository.Member
	userRepo      repository.User
	namespaceRepo repository.Namespace
}

func (m *Member) InviteMember(ctx context.Context, req *bo.InviteMemberBo) error {
	user, err := m.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("user with email %s not found, please ask them to sign up first", req.Email)
		}
		m.helper.Errorw("msg", "get user failed", "error", err, "email", req.Email)
		return merr.ErrorInternalServer("invite member failed").WithCause(err)
	}
	namespaceUID := contextx.GetNamespace(ctx)
	if namespaceUID <= 0 {
		return merr.ErrorInvalidArgument("namespace is required")
	}
	_, err = m.memberRepo.GetMemberByNamespaceAndUser(ctx, namespaceUID, user.UID)
	if err == nil {
		return merr.ErrorParams("user %s is already a member", req.Email)
	}
	if !merr.IsNotFound(err) {
		m.helper.Errorw("msg", "check member exists failed", "error", err)
		return merr.ErrorInternalServer("invite member failed").WithCause(err)
	}
	creator := contextx.GetUserUID(ctx)
	createBo := &bo.CreateMemberBo{
		Creator:      creator,
		NamespaceUID: namespaceUID,
		UserUID:      user.UID,
		Name:         user.Name,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
	}
	if err := m.memberRepo.CreateMember(ctx, createBo); err != nil {
		m.helper.Errorw("msg", "create member failed", "error", err, "email", req.Email)
		return merr.ErrorInternalServer("invite member failed").WithCause(err)
	}
	return nil
}

func (m *Member) DismissMember(ctx context.Context, memberUID snowflake.ID) error {
	if err := m.memberRepo.DeleteMember(ctx, memberUID); err != nil {
		m.helper.Errorw("msg", "dismiss member failed", "error", err, "uid", memberUID)
		return merr.ErrorInternalServer("dismiss member failed").WithCause(err)
	}
	return nil
}

func (m *Member) GetMember(ctx context.Context, memberUID snowflake.ID) (*bo.MemberItemBo, error) {
	member, err := m.memberRepo.GetMember(ctx, memberUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("member %s not found", memberUID)
		}
		m.helper.Errorw("msg", "get member failed", "error", err, "uid", memberUID)
		return nil, merr.ErrorInternalServer("get member failed").WithCause(err)
	}
	return member, nil
}

func (m *Member) ListMember(ctx context.Context, req *bo.ListMemberBo) (*bo.PageResponseBo[*bo.MemberItemBo], error) {
	page, err := m.memberRepo.ListMember(ctx, req)
	if err != nil {
		m.helper.Errorw("msg", "list member failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list member failed").WithCause(err)
	}
	return page, nil
}

func (m *Member) SelectMember(ctx context.Context, req *bo.SelectMemberBo) (*bo.SelectMemberBoResult, error) {
	result, err := m.memberRepo.SelectMember(ctx, req)
	if err != nil {
		m.helper.Errorw("msg", "select member failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select member failed").WithCause(err)
	}
	return result, nil
}

func (m *Member) UpdateMemberStatus(ctx context.Context, memberUID snowflake.ID, status int32) error {
	if err := m.memberRepo.UpdateMemberStatus(ctx, memberUID, status); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("member %s not found", memberUID)
		}
		m.helper.Errorw("msg", "update member status failed", "error", err, "uid", memberUID)
		return merr.ErrorInternalServer("update member status failed").WithCause(err)
	}
	return nil
}

func (m *Member) GetNamespaceUIDsByUserUID(ctx context.Context, userUID snowflake.ID) ([]snowflake.ID, error) {
	return m.memberRepo.GetNamespaceUIDsByUserUID(ctx, userUID)
}
