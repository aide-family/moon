package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToTeam(ctx context.Context, teamDo do.Team) *system.Team {
	if validate.IsNil(teamDo) {
		return nil
	}
	team, ok := teamDo.(*system.Team)
	if ok {
		team.WithContext(ctx)
		return team
	}
	return &system.Team{
		CreatorModel:  ToCreatorModel(ctx, teamDo),
		Name:          teamDo.GetName(),
		Status:        teamDo.GetStatus(),
		Remark:        teamDo.GetRemark(),
		Logo:          teamDo.GetLogo(),
		LeaderID:      teamDo.GetLeaderID(),
		UUID:          teamDo.GetUUID(),
		Capacity:      teamDo.GetCapacity(),
		Leader:        ToUser(ctx, teamDo.GetLeader()),
		Admins:        ToUsers(ctx, teamDo.GetAdmins()),
		Resources:     nil,
		BizDBConfig:   crypto.NewObject(teamDo.GetBizDBConfig()),
		AlarmDBConfig: crypto.NewObject(teamDo.GetAlarmDBConfig()),
	}
}

func ToTeams(ctx context.Context, teamDos []do.Team) []*system.Team {
	return slices.MapFilter(teamDos, func(teamDo do.Team) (*system.Team, bool) {
		if validate.IsNil(teamDo) {
			return nil, false
		}
		return ToTeam(ctx, teamDo), true
	})
}

func ToTeamMember(ctx context.Context, memberDo do.TeamMember) *system.TeamMember {
	if validate.IsNil(memberDo) {
		return nil
	}
	member, ok := memberDo.(*system.TeamMember)
	if ok {
		member.WithContext(ctx)
		return member
	}
	return &system.TeamMember{
		TeamModel:  ToTeamModel(ctx, memberDo),
		MemberName: memberDo.GetMemberName(),
		Remark:     memberDo.GetRemark(),
		UserID:     memberDo.GetUserID(),
		InviterID:  memberDo.GetInviterID(),
		Position:   memberDo.GetPosition(),
		Status:     memberDo.GetStatus(),
		Roles:      ToTeamRoles(ctx, memberDo.GetRoles()),
		User:       ToUser(ctx, memberDo.GetUser()),
		Inviter:    ToUser(ctx, memberDo.GetInviter()),
	}
}

func ToTeamMembers(ctx context.Context, memberDos []do.TeamMember) []*system.TeamMember {
	return slices.MapFilter(memberDos, func(memberDo do.TeamMember) (*system.TeamMember, bool) {
		if validate.IsNil(memberDo) {
			return nil, false
		}
		return ToTeamMember(ctx, memberDo), true
	})
}
