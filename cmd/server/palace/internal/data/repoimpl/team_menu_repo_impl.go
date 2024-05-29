package repoimpl

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel/bizquery"
)

func NewTeamMenuRepo(data *data.Data) repo.TeamMenuRepo {
	return &teamMenuRepoImpl{
		data: data,
	}
}

type teamMenuRepoImpl struct {
	data *data.Data
}

func (l *teamMenuRepoImpl) GetTeamMenuList(ctx context.Context, params *bo.QueryTeamMenuListParams) ([]*bizmodel.SysTeamMenu, error) {
	bizDB, err := l.data.GetBizGormDB(params.TeamID)
	if err != nil {
		return nil, err
	}
	q := bizquery.Use(bizDB)
	return q.SysTeamMenu.WithContext(ctx).Find()
}
