package bo

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
)

type SaveMenuRequest struct {
	MenuId        uint32
	Name          string
	MenuPath      string
	MenuIcon      string
	MenuType      vobj.MenuType
	MenuCategory  vobj.MenuCategory
	ApiPath       string
	Status        vobj.GlobalStatus
	ProcessType   vobj.MenuProcessType
	ParentID      uint32
	RelyOnBrother bool
}

type ListMenuParams struct {
	*PaginationRequest
	TeamID       uint32
	Status       vobj.GlobalStatus
	MenuType     vobj.MenuType
	MenuCategory vobj.MenuCategory
	ProcessType  vobj.MenuProcessType
	Keyword      string
}

func (l *ListMenuParams) WithTeamID(ctx context.Context) (*ListMenuParams, error) {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorParams("team id is not found")
	}
	l.TeamID = teamID
	return l, nil
}

func (l *ListMenuParams) ToListReply(items []do.Menu) *ListMenuReply {
	return &ListMenuReply{
		Items:           items,
		PaginationReply: l.ToReply(),
	}
}

type ListMenuReply = ListReply[do.Menu]
