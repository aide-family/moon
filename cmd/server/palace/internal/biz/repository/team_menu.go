package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// TeamMenu 团队菜单接口
type TeamMenu interface {
	GetTeamMenuList(context.Context, *bo.QueryTeamMenuListParams) ([]*bizmodel.SysTeamMenu, error)
}
