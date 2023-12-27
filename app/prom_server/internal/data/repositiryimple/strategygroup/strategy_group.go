package strategygroup

import (
	"bytes"
	"context"
	"fmt"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"prometheus-manager/pkg/helper/model/strategyscopes"

	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/strategygroupscopes"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.StrategyGroupRepo = (*strategyGroupRepoImpl)(nil)

type (
	strategyGroupRepoImpl struct {
		repository.UnimplementedStrategyGroupRepo
		query.IAction[model.PromStrategyGroup]

		data *data.Data
		log  *log.Helper
	}
)

type StrategyCount struct {
	Count   int64  `json:"count"`
	GroupId uint32 `json:"group_id"`
}

func (l *strategyGroupRepoImpl) UpdateStrategyCount(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}

	strategyDB := l.data.DB().Model(&model.PromStrategy{}).WithContext(ctx)
	var strategyCountList []StrategyCount
	if err := strategyDB.Scopes(strategyscopes.GroupIdsEQ(ids...)).Select("count(id) as count, group_id").Group("group_id").Scan(&strategyCountList).Error; err != nil {
		return err
	}
	var caseSet bytes.Buffer
	caseSet.WriteString("CASE id ")
	for _, strategyCount := range strategyCountList {
		caseSet.WriteString("WHEN ")
		caseSet.WriteString(fmt.Sprintf("%d", strategyCount.GroupId))
		caseSet.WriteString(" THEN ")
		caseSet.WriteString(fmt.Sprintf("%d", strategyCount.Count))
		caseSet.WriteString(" ")
	}
	caseSet.WriteString("END")
	db := l.data.DB().Model(&model.PromStrategyGroup{}).WithContext(ctx)
	return db.Scopes(strategygroupscopes.InIds(ids)).Update("strategy_count", gorm.Expr(caseSet.String())).Error
}

func (l *strategyGroupRepoImpl) Create(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	strategyGroupModel := strategyGroup.ToModel()
	if err := l.WithContext(ctx).Create(strategyGroupModel); err != nil {
		return nil, err
	}
	return bo.StrategyGroupModelToBO(strategyGroupModel), nil
}

func (l *strategyGroupRepoImpl) UpdateById(ctx context.Context, id uint32, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	strategyGroupModel := strategyGroup.ToModel()
	if err := l.WithContext(ctx).UpdateByID(id, strategyGroupModel); err != nil {
		return nil, err
	}
	return bo.StrategyGroupModelToBO(strategyGroupModel), nil
}

func (l *strategyGroupRepoImpl) BatchUpdateStatus(ctx context.Context, status valueobj.Status, ids []uint32) error {
	if err := l.WithContext(ctx).Update(&model.PromStrategyGroup{Status: status}, strategygroupscopes.InIds(ids)); err != nil {
		return err
	}
	return nil
}

func (l *strategyGroupRepoImpl) DeleteByIds(ctx context.Context, ids ...uint32) error {
	if err := l.WithContext(ctx).Delete(strategygroupscopes.InIds(ids)); err != nil {
		return err
	}
	return nil
}

func (l *strategyGroupRepoImpl) GetById(ctx context.Context, id uint32) (*bo.StrategyGroupBO, error) {
	first, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}
	return bo.StrategyGroupModelToBO(first), nil
}

func (l *strategyGroupRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	strategyModelList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	list := slices.To(strategyModelList, func(m *model.PromStrategyGroup) *bo.StrategyGroupBO {
		return bo.StrategyGroupModelToBO(m)
	})
	return list, nil
}

func NewStrategyGroupRepo(data *data.Data, logger log.Logger) repository.StrategyGroupRepo {
	return &strategyGroupRepoImpl{
		IAction: query.NewAction[model.PromStrategyGroup](
			query.WithDB[model.PromStrategyGroup](data.DB()),
		),
		data: data,
		log:  log.NewHelper(logger),
	}
}
