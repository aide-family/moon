package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

func NewMenuBiz(teamMenuRepo repository.TeamMenu, msgRepo repository.Msg) *MenuBiz {
	return &MenuBiz{
		teamMenuRepo: teamMenuRepo,
		msgRepo:      msgRepo,
	}
}

type MenuBiz struct {
	teamMenuRepo repository.TeamMenu
	msgRepo      repository.Msg
}

// MenuList 菜单列表
func (b *MenuBiz) MenuList(ctx context.Context) ([]*bizmodel.SysTeamMenu, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	return b.teamMenuRepo.GetTeamMenuList(ctx, &bo.QueryTeamMenuListParams{TeamID: claims.GetTeam()})
}
