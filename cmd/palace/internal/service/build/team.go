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

func ToSaveOneTeamRequest(req *palace.SaveTeamRequest) *bo.SaveOneTeamRequest {
	if validate.IsNil(req) {
		panic("SaveOneTeamRequest is nil")
	}
	return &bo.SaveOneTeamRequest{
		TeamID:     req.GetTeamId(),
		TeamName:   req.GetName(),
		TeamRemark: req.GetRemark(),
		TeamLogo:   req.GetLogo(),
	}
}

func ToSaveOneTeamRequestByCreate(req *palace.CreateTeamRequest) *bo.SaveOneTeamRequest {
	if validate.IsNil(req) {
		panic("CreateTeamRequest is nil")
	}
	return &bo.SaveOneTeamRequest{
		TeamName:   req.GetName(),
		TeamRemark: req.GetRemark(),
		TeamLogo:   req.GetLogo(),
	}
}

func ToTeamItem(team do.Team) *common.TeamItem {
	if validate.IsNil(team) {
		return nil
	}

	return &common.TeamItem{
		TeamId:          team.GetID(),
		Uuid:            team.GetUUID().String(),
		Name:            team.GetName(),
		Remark:          team.GetRemark(),
		Logo:            team.GetLogo(),
		Status:          common.TeamStatus(team.GetStatus().GetValue()),
		Creator:         ToUserBaseItem(team.GetCreator()),
		Leader:          ToUserBaseItem(team.GetLeader()),
		Admins:          ToUserBaseItems(team.GetAdmins()),
		CreatedAt:       timex.Format(team.GetCreatedAt()),
		UpdatedAt:       timex.Format(team.GetUpdatedAt()),
		MemberCount:     0,
		StrategyCount:   0,
		DatasourceCount: 0,
	}
}

func ToTeamBaseItem(team do.Team) *common.TeamBaseItem {
	if validate.IsNil(team) {
		return nil
	}

	return &common.TeamBaseItem{
		TeamId: team.GetID(),
		Name:   team.GetName(),
		Remark: team.GetRemark(),
		Logo:   team.GetLogo(),
	}
}

// ToTeamItems converts a slice of system Team objects to a slice of TeamItem proto objects
func ToTeamItems(teams []do.Team) []*common.TeamItem {
	return slices.Map(teams, ToTeamItem)
}

func ToTeamBaseItems(teams []do.Team) []*common.TeamBaseItem {
	return slices.Map(teams, ToTeamBaseItem)
}

func ToTeamListRequest(req *palace.GetTeamListRequest) *bo.TeamListRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.TeamListRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Keyword:           req.GetKeyword(),
		Status: slices.MapFilter(req.GetStatus(), func(status common.TeamStatus) (vobj.TeamStatus, bool) {
			vobjStatus := vobj.TeamStatus(status)
			return vobjStatus, !vobjStatus.IsUnknown() && vobjStatus.Exist()
		}),
		UserIds:   nil,
		LeaderId:  req.GetLeaderId(),
		CreatorId: req.GetCreatorId(),
	}
}

func ToTeamMemberListRequest(req *palace.GetTeamMembersRequest, teamId uint32) *bo.TeamMemberListRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.TeamMemberListRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Keyword:           req.GetKeyword(),
		Status: slices.MapFilter(req.GetStatus(), func(status common.MemberStatus) (vobj.MemberStatus, bool) {
			vobjStatus := vobj.MemberStatus(status)
			return vobjStatus, !vobjStatus.IsUnknown() && vobjStatus.Exist()
		}),
		Positions: slices.MapFilter(req.GetPositions(), func(position common.MemberPosition) (vobj.Position, bool) {
			vobjPosition := vobj.Position(position)
			return vobjPosition, !vobjPosition.IsUnknown() && vobjPosition.Exist()
		}),
		TeamId: teamId,
	}
}

func ToTeamMemberSelectRequest(req *palace.SelectTeamMembersRequest, teamId uint32) *bo.SelectTeamMembersRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.SelectTeamMembersRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Keyword:           req.GetKeyword(),
		Status: slices.MapFilter(req.GetStatus(), func(status common.MemberStatus) (vobj.MemberStatus, bool) {
			vobjStatus := vobj.MemberStatus(status)
			return vobjStatus, !vobjStatus.IsUnknown() && vobjStatus.Exist()
		}),
		TeamId: teamId,
	}
}

func ToTeamMemberItem(member do.TeamMember) *common.TeamMemberItem {
	if validate.IsNil(member) {
		return nil
	}
	return &common.TeamMemberItem{
		TeamMemberId: member.GetTeamMemberID(),
		User:         ToUserBaseItem(member.GetUser()),
		Position:     common.MemberPosition(member.GetPosition().GetValue()),
		Status:       common.MemberStatus(member.GetStatus().GetValue()),
		Inviter:      ToUserBaseItem(member.GetInviter()),
		Roles:        ToTeamRoleItems(member.GetRoles()),
		CreatedAt:    timex.Format(member.GetCreatedAt()),
		UpdatedAt:    timex.Format(member.GetUpdatedAt()),
	}
}

func ToTeamMemberItems(members []do.TeamMember) []*common.TeamMemberItem {
	return slices.Map(members, ToTeamMemberItem)
}

func ToTeamMemberBaseItem(member do.TeamMember) *common.TeamMemberBaseItem {
	if validate.IsNil(member) {
		return nil
	}
	return &common.TeamMemberBaseItem{
		TeamMemberId: member.GetID(),
		MemberName:   member.GetMemberName(),
		Remark:       member.GetRemark(),
		Position:     common.MemberPosition(member.GetPosition().GetValue()),
		Status:       common.MemberStatus(member.GetStatus().GetValue()),
		CreatedAt:    timex.Format(member.GetCreatedAt()),
		UpdatedAt:    timex.Format(member.GetUpdatedAt()),
		User:         ToUserBaseItem(member.GetUser()),
	}
}

func ToTeamMemberBaseItems(members []do.TeamMember) []*common.TeamMemberBaseItem {
	return slices.Map(members, ToTeamMemberBaseItem)
}
