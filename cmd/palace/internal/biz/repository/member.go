package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type Member interface {
	FindByUserID(ctx context.Context, userID uint32) (do.TeamMember, error)
	List(ctx context.Context, req *bo.TeamMemberListRequest) (*bo.TeamMemberListReply, error)
	Select(ctx context.Context, req *bo.SelectTeamMembersRequest) (*bo.SelectTeamMembersReply, error)
	UpdateStatus(ctx context.Context, req bo.UpdateMemberStatus) error
	UpdatePosition(ctx context.Context, req bo.UpdateMemberPosition) error
	UpdateRoles(ctx context.Context, req bo.UpdateMemberRoles) error
	Get(ctx context.Context, id uint32) (do.TeamMember, error)
	Find(ctx context.Context, ids []uint32) ([]do.TeamMember, error)
	Create(ctx context.Context, req *bo.CreateTeamMemberReq) error
}
