package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type Team interface {
	FindByID(ctx context.Context, id uint32) (do.Team, error)
	Create(ctx context.Context, team bo.CreateTeamRequest) (do.Team, error)
	Update(ctx context.Context, team bo.UpdateTeamRequest) (do.Team, error)
	Delete(ctx context.Context, id uint32) error
	List(ctx context.Context, req *bo.TeamListRequest) (*bo.TeamListReply, error)
}
