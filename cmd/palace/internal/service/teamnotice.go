package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
)

type TeamNoticeService struct {
	palace.UnimplementedTeamNoticeServer
	teamHookBiz   *biz.TeamHook
	teamNoticeBiz *biz.TeamNotice
}

func NewTeamNoticeService(
	teamHookBiz *biz.TeamHook,
	teamNoticeBiz *biz.TeamNotice,
) *TeamNoticeService {
	return &TeamNoticeService{
		teamHookBiz:   teamHookBiz,
		teamNoticeBiz: teamNoticeBiz,
	}
}

// SaveTeamNoticeHook saves a team notice hook
func (s *TeamNoticeService) SaveTeamNoticeHook(ctx context.Context, req *palace.SaveTeamNoticeHookRequest) (*common.EmptyReply, error) {
	if err := s.teamHookBiz.SaveHook(ctx, build.ToSaveTeamNoticeHookRequest(req)); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

// UpdateTeamNoticeHookStatus updates the status of a hook
func (s *TeamNoticeService) UpdateTeamNoticeHookStatus(ctx context.Context, req *palace.UpdateTeamNoticeHookStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateTeamNoticeHookStatusRequest{
		HookID: req.GetHookId(),
		Status: vobj.GlobalStatus(req.GetStatus()),
	}
	if err := s.teamHookBiz.UpdateHookStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

// DeleteTeamNoticeHook deletes a hook
func (s *TeamNoticeService) DeleteTeamNoticeHook(ctx context.Context, req *palace.DeleteTeamNoticeHookRequest) (*common.EmptyReply, error) {
	if err := s.teamHookBiz.DeleteHook(ctx, req.GetHookId()); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

// GetTeamNoticeHook gets the details of a hook
func (s *TeamNoticeService) GetTeamNoticeHook(ctx context.Context, req *palace.GetTeamNoticeHookRequest) (*common.NoticeHookItem, error) {
	hook, err := s.teamHookBiz.GetHook(ctx, req.GetHookId())
	if err != nil {
		return nil, err
	}
	return build.ToNoticeHookItem(hook), nil
}

// ListTeamNoticeHook gets the list of hooks
func (s *TeamNoticeService) ListTeamNoticeHook(ctx context.Context, req *palace.ListTeamNoticeHookRequest) (*palace.ListTeamNoticeHookReply, error) {
	reply, err := s.teamHookBiz.ListHook(ctx, build.ToListTeamNoticeHookRequest(req))
	if err != nil {
		return nil, err
	}

	return &palace.ListTeamNoticeHookReply{
		Items:      build.ToNoticeHookItems(reply.Items),
		Pagination: build.ToPaginationReply(reply.PaginationReply),
	}, nil
}

func (s *TeamNoticeService) TeamNoticeHookSelect(ctx context.Context, req *palace.TeamNoticeHookSelectRequest) (*palace.TeamNoticeHookSelectReply, error) {
	reply, err := s.teamHookBiz.SelectHook(ctx, build.ToTeamNoticeHookSelectRequest(req))
	if err != nil {
		return nil, err
	}
	return &palace.TeamNoticeHookSelectReply{
		Pagination: build.ToPaginationReply(reply.PaginationReply),
		Items:      build.ToSelectItems(reply.Items),
	}, nil
}

func (s *TeamNoticeService) SaveTeamNoticeGroup(ctx context.Context, req *palace.SaveTeamNoticeGroupRequest) (*common.EmptyReply, error) {
	members := slices.Map(req.GetMembers(), func(member *palace.SaveTeamNoticeGroupRequest_Member) *bo.SaveNoticeMemberItem {
		return &bo.SaveNoticeMemberItem{
			MemberID:     member.GetMemberId(),
			UserID:       0,
			NoticeType:   vobj.NoticeType(member.GetNoticeType()),
			DutyCycleIds: member.GetDutyCycleIds(),
		}
	})
	params := &bo.SaveNoticeGroupReq{
		GroupID:       req.GetGroupId(),
		Name:          req.GetName(),
		Remark:        req.GetRemark(),
		HookIds:       req.GetHookIds(),
		NoticeMembers: members,
		EmailConfigID: req.GetEmailConfigId(),
		SMSConfigID:   req.GetSmsConfigId(),
	}
	if err := s.teamNoticeBiz.SaveNoticeGroup(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *TeamNoticeService) UpdateTeamNoticeGroupStatus(ctx context.Context, req *palace.UpdateTeamNoticeGroupStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateTeamNoticeGroupStatusRequest{
		GroupIds: []uint32{req.GetGroupId()},
		Status:   vobj.GlobalStatus(req.GetStatus()),
	}
	if err := s.teamNoticeBiz.UpdateNoticeGroupStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *TeamNoticeService) DeleteTeamNoticeGroup(ctx context.Context, req *palace.DeleteTeamNoticeGroupRequest) (*common.EmptyReply, error) {
	if err := s.teamNoticeBiz.DeleteNoticeGroup(ctx, req.GetGroupId()); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *TeamNoticeService) GetTeamNoticeGroup(ctx context.Context, req *palace.GetTeamNoticeGroupRequest) (*common.NoticeGroupItem, error) {
	noticeGroup, err := s.teamNoticeBiz.GetNoticeGroup(ctx, req.GetGroupId())
	if err != nil {
		return nil, err
	}
	return build.ToNoticeGroupItem(noticeGroup), nil
}

func (s *TeamNoticeService) ListTeamNoticeGroup(ctx context.Context, req *palace.ListTeamNoticeGroupRequest) (*palace.ListTeamNoticeGroupReply, error) {
	params := &bo.ListNoticeGroupReq{
		PaginationRequest: build.ToPaginationRequest(req.GetPagination()),
		MemberIds:         req.GetMemberIds(),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Keyword:           req.GetKeyword(),
	}
	reply, err := s.teamNoticeBiz.TeamNoticeGroups(ctx, params)
	if err != nil {
		return nil, err
	}

	return &palace.ListTeamNoticeGroupReply{
		Pagination: build.ToPaginationReply(reply.PaginationReply),
		Items:      build.ToNoticeGroupItems(reply.Items),
	}, nil
}

func (s *TeamNoticeService) TeamNoticeGroupSelect(ctx context.Context, req *palace.TeamNoticeGroupSelectRequest) (*palace.TeamNoticeGroupSelectReply, error) {
	reply, err := s.teamNoticeBiz.SelectNoticeGroup(ctx, build.ToTeamNoticeGroupSelectRequest(req))
	if err != nil {
		return nil, err
	}
	return &palace.TeamNoticeGroupSelectReply{
		Pagination: build.ToPaginationReply(reply.PaginationReply),
		Items:      build.ToSelectItems(reply.Items),
	}, nil
}
