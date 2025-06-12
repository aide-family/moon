package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToTeamNoticeMember(ctx context.Context, member do.NoticeMember) *team.NoticeMember {
	if validate.IsNil(member) {
		return nil
	}
	memberDo := &team.NoticeMember{
		TeamModel:     ToTeamModel(ctx, member),
		NoticeGroupID: member.GetNoticeGroupID(),
		UserID:        member.GetUserID(),
		NoticeType:    member.GetNoticeType(),
		NoticeGroup:   ToTeamNoticeGroup(ctx, member.GetNoticeGroup()),
		DutyCycle: slices.MapFilter(member.GetDutyCycle(), func(v do.TimeEngine) (*team.TimeEngine, bool) {
			if validate.IsNil(v) {
				return nil, false
			}
			return ToTimeEngine(ctx, v), true
		}),
	}
	memberDo.WithContext(ctx)
	return memberDo
}

func ToTeamNoticeMembers(ctx context.Context, members []do.NoticeMember) []*team.NoticeMember {
	return slices.MapFilter(members, func(v do.NoticeMember) (*team.NoticeMember, bool) {
		if validate.IsNil(v) {
			return nil, false
		}
		return ToTeamNoticeMember(ctx, v), true
	})
}
