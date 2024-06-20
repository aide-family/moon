package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type TeamMenu interface {
	GetTeamMenuList(ctx context.Context, params *bo.QueryTeamMenuListParams) ([]*bizmodel.SysTeamMenu, error)
}
