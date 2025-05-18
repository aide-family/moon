package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToNoticeMemberItem(noticeMember do.NoticeMember) *common.NoticeMemberItem {
	if validate.IsNil(noticeMember) {
		return nil
	}
	return &common.NoticeMemberItem{
		NoticeGroupId: noticeMember.GetNoticeGroupID(),
		UserId:        noticeMember.GetUserID(),
		NoticeType:    common.NoticeType(noticeMember.GetNoticeType()),
		NoticeGroup:   ToNoticeGroupItem(noticeMember.GetNoticeGroup()),
		Member:        ToTeamMemberBaseItem(noticeMember.GetMember()),
	}
}

func ToNoticeMemberItems(noticeMembers []do.NoticeMember) []*common.NoticeMemberItem {
	return slices.Map(noticeMembers, ToNoticeMemberItem)
}

func ToNoticeGroupItem(noticeGroup do.NoticeGroup) *common.NoticeGroupItem {
	if validate.IsNil(noticeGroup) {
		return nil
	}
	return &common.NoticeGroupItem{
		NoticeGroupId: noticeGroup.GetID(),
		CreatedAt:     timex.Format(noticeGroup.GetCreatedAt()),
		UpdatedAt:     timex.Format(noticeGroup.GetUpdatedAt()),
		Name:          noticeGroup.GetName(),
		Remark:        noticeGroup.GetRemark(),
		Status:        common.GlobalStatus(noticeGroup.GetStatus().GetValue()),
		NoticeMembers: ToNoticeMemberItems(noticeGroup.GetNoticeMembers()),
		Hooks:         ToNoticeHookItems(noticeGroup.GetHooks()),
		Creator:       ToUserBaseItem(noticeGroup.GetCreator()),
	}
}

func ToNoticeGroupItems(noticeGroups []do.NoticeGroup) []*common.NoticeGroupItem {
	return slices.Map(noticeGroups, ToNoticeGroupItem)
}

func ToTeamNoticeGroupSelectRequest(req *palace.TeamNoticeGroupSelectRequest) *bo.TeamNoticeGroupSelectRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.TeamNoticeGroupSelectRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Keyword:           req.GetKeyword(),
		Status:            vobj.GlobalStatus(req.GetStatus()),
	}
}
