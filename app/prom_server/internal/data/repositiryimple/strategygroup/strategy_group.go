package strategygroup

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.StrategyGroupRepo = (*strategyGroupRepoImpl)(nil)

type (
	strategyGroupRepoImpl struct {
		repository.UnimplementedStrategyGroupRepo

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

	tx := basescopes.GetTx(ctx, l.data.DB())
	strategyDB := tx.Model(&do.PromStrategy{}).WithContext(ctx)
	var strategyCountList []StrategyCount
	if err := strategyDB.Scopes(basescopes.StrategyTableGroupIdsEQ(ids...)).Select("count(id) as count, group_id").Group("group_id").Scan(&strategyCountList).Error; err != nil {
		return err
	}

	db := tx.Model(&do.PromStrategyGroup{}).WithContext(ctx)
	switch len(strategyCountList) {
	case 0:
		return db.Scopes(basescopes.InIds(ids...)).Update("strategy_count", 0).Error
	case 1:
		return db.Scopes(basescopes.InIds(ids...)).Update("strategy_count", strategyCountList[0].Count).Error
	default:
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

		return db.Scopes(basescopes.InIds(ids...)).Update("strategy_count", gorm.Expr(caseSet.String())).Error
	}
}

func (l *strategyGroupRepoImpl) UpdateEnableStrategyCount(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}

	tx := basescopes.GetTx(ctx, l.data.DB())
	strategyDB := tx.Model(&do.PromStrategy{}).WithContext(ctx)
	var strategyCountList []StrategyCount
	wheres := []basescopes.ScopeMethod{
		basescopes.StrategyTableGroupIdsEQ(ids...),
		basescopes.StatusEQ(vo.StatusEnabled),
	}
	if err := strategyDB.Scopes(wheres...).Select("count(id) as count, group_id").Group("group_id").Scan(&strategyCountList).Error; err != nil {
		return err
	}

	db := tx.Model(&do.PromStrategyGroup{}).WithContext(ctx)
	switch len(strategyCountList) {
	case 0:
		return db.Scopes(basescopes.InIds(ids...)).Update("enable_strategy_count", 0).Error
	case 1:
		return db.Scopes(basescopes.InIds(ids...)).Update("enable_strategy_count", strategyCountList[0].Count).Error
	default:
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

		return db.Scopes(basescopes.InIds(ids...)).Update("enable_strategy_count", gorm.Expr(caseSet.String())).Error
	}
}

func (l *strategyGroupRepoImpl) Create(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	strategyGroupModel := strategyGroup.ToModel()
	if err := l.data.DB().WithContext(ctx).Create(strategyGroupModel).Error; err != nil {
		return nil, err
	}
	return bo.StrategyGroupModelToBO(strategyGroupModel), nil
}

func (l *strategyGroupRepoImpl) UpdateById(ctx context.Context, id uint32, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	strategyGroupModel := strategyGroup.ToModel()
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(id)).Updates(strategyGroupModel).Error; err != nil {
		return nil, err
	}
	return bo.StrategyGroupModelToBO(strategyGroupModel), nil
}

func (l *strategyGroupRepoImpl) BatchUpdateStatus(ctx context.Context, status vo.Status, ids []uint32) error {
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(ids...)).Updates(&do.PromStrategyGroup{Status: status}).Error; err != nil {
		return err
	}
	return nil
}

func (l *strategyGroupRepoImpl) DeleteByIds(ctx context.Context, ids ...uint32) error {
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(ids...)).Delete(&do.PromStrategyGroup{}).Error; err != nil {
		return err
	}
	return nil
}

func (l *strategyGroupRepoImpl) GetById(ctx context.Context, id uint32) (*bo.StrategyGroupBO, error) {
	var first do.PromStrategyGroup
	if err := l.data.DB().WithContext(ctx).First(&first, id).Error; err != nil {
		return nil, err
	}

	return bo.StrategyGroupModelToBO(&first), nil
}

func (l *strategyGroupRepoImpl) List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	var strategyModelList []*do.PromStrategyGroup
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, basescopes.Page(pgInfo))...).Find(&strategyModelList).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.PromStrategyGroup{}).WithContext(ctx).Scopes(scopes...).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}
	list := slices.To(strategyModelList, func(m *do.PromStrategyGroup) *bo.StrategyGroupBO {
		return bo.StrategyGroupModelToBO(m)
	})
	return list, nil
}

func (l *strategyGroupRepoImpl) ListAllLimit(ctx context.Context, limit int, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	var strategyModelList []*do.PromStrategyGroup
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Limit(limit).Find(&strategyModelList).Error; err != nil {
		return nil, err
	}
	list := slices.To(strategyModelList, func(m *do.PromStrategyGroup) *bo.StrategyGroupBO {
		return bo.StrategyGroupModelToBO(m)
	})
	return list, nil
}

func NewStrategyGroupRepo(data *data.Data, logger log.Logger) repository.StrategyGroupRepo {
	return &strategyGroupRepoImpl{
		data: data,
		log:  log.NewHelper(logger),
	}
}
