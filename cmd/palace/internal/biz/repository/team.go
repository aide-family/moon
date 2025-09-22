package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type Team interface {
	FindByID(ctx context.Context, id uint32) (do.Team, error)
	Create(ctx context.Context, team bo.CreateTeamRequest) error
	Update(ctx context.Context, team bo.UpdateTeamRequest) error
	UpdateStatus(ctx context.Context, teamID uint32, status vobj.TeamStatus) error
	Delete(ctx context.Context, id uint32) error
	List(ctx context.Context, req *bo.TeamListRequest) (*bo.TeamListReply, error)
	CheckNameUnique(ctx context.Context, name string, teamID uint32) error
	FindByName(ctx context.Context, name string) (do.Team, error)
}
