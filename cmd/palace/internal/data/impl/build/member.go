package build

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func ToStrategyMember(ctx context.Context, member do.NoticeMember) *team.NoticeMember {
	if validate.IsNil(member) {
		return nil
	}
	if member, ok := member.(*team.NoticeMember); ok {
		member.WithContext(ctx)
		return member
	}
	return &team.NoticeMember{
		TeamModel:     ToTeamModel(ctx, member),
		NoticeGroupID: member.GetNoticeGroupID(),
		UserID:        member.GetUserID(),
		NoticeType:    member.GetNoticeType(),
		NoticeGroup:   ToStrategyNotice(ctx, member.GetNoticeGroup()),
	}
}

func ToStrategyMembers(ctx context.Context, members []do.NoticeMember) []*team.NoticeMember {
	return slices.MapFilter(members, func(v do.NoticeMember) (*team.NoticeMember, bool) {
		if validate.IsNil(v) {
			return nil, false
		}
		return ToStrategyMember(ctx, v), true
	})
}
