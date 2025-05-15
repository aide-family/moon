package impl

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewMenuRepo(d *data.Data) repository.Menu {
	return &menuImpl{
		Data: d,
	}
}

type menuImpl struct {
	*data.Data
}

func (m *menuImpl) Find(ctx context.Context, ids []uint32) ([]do.Menu, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	mainQuery := getMainQuery(ctx, m)
	menu := mainQuery.Menu
	menuDo, err := menu.WithContext(ctx).Where(menu.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	menus := slices.Map(menuDo, func(menu *system.Menu) do.Menu { return menu })
	return menus, nil
}
