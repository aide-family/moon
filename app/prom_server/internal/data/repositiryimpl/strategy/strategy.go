package strategy

import (
	"context"

	"github.com/aide-family/moon/pkg/after"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ repository.StrategyRepo = (*strategyRepoImpl)(nil)

type (
	strategyRepoImpl struct {
		repository.UnimplementedStrategyRepo

		changeGroupChannel chan<- uint32

		data *data.Data
		log  *log.Helper

		strategyGroupRepo repository.StrategyGroupRepo
	}
)

func (l *strategyRepoImpl) BindStrategyNotifyObject(ctx context.Context, strategyBo *bo.StrategyBO, notifyBo []*bo.NotifyBO) error {
	strategyModel := strategyBo.ToModel()
	notifyModels := slices.To(notifyBo, func(item *bo.NotifyBO) *do.PromAlarmNotify {
		return item.ToModel()
	})
	return l.data.DB().WithContext(ctx).Model(strategyModel).Association(do.PromStrategyPreloadFieldPromNotifies).Replace(notifyModels)
}

func (l *strategyRepoImpl) List(ctx context.Context, wheres ...basescopes.ScopeMethod) ([]*bo.StrategyBO, error) {
	var modelList []*do.PromStrategy
	whereList := append(wheres, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Scopes(whereList...).Find(&modelList).Error; err != nil {
		return nil, err
	}
	list := slices.To(modelList, func(item *do.PromStrategy) *bo.StrategyBO {
		return bo.StrategyModelToBO(item)
	})
	return list, nil
}

func (l *strategyRepoImpl) ListStrategyByIds(ctx context.Context, ids []uint32) ([]*bo.StrategyBO, error) {
	modelList := make([]*do.PromStrategy, 0, len(ids))
	whereList := []basescopes.ScopeMethod{
		basescopes.InIds(ids...),
		basescopes.WithCreateBy(ctx),
	}
	if err := l.data.DB().WithContext(ctx).Scopes(whereList...).Find(&modelList).Error; err != nil {
		return nil, err
	}

	list := slices.To(modelList, func(item *do.PromStrategy) *bo.StrategyBO {
		return bo.StrategyModelToBO(item)
	})
	return list, nil
}

// CreateStrategy 创建策略
func (l *strategyRepoImpl) CreateStrategy(ctx context.Context, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
	newStrategy := strategyBO.ToModel()
	// 替换报警页面和分类
	alarmPages := slices.To(strategyBO.AlarmPageIds, func(pageId uint32) *do.SysDict {
		return &do.SysDict{
			BaseModel: do.BaseModel{ID: pageId},
		}
	})
	categories := slices.To(strategyBO.CategoryIds, func(categoryId uint32) *do.SysDict {
		return &do.SysDict{
			BaseModel: do.BaseModel{ID: categoryId},
		}
	})
	newStrategy.CreateBy = middler.GetUserId(ctx)
	// 默认不开启
	newStrategy.Status = vobj.StatusDisabled

	err := l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := basescopes.WithTx(ctx, tx)
		if err := tx.WithContext(txCtx).Create(newStrategy).Error; err != nil {
			return err
		}

		if err := tx.WithContext(txCtx).Model(newStrategy).Association(do.PromStrategyPreloadFieldAlarmPages).Replace(alarmPages); err != nil {
			return err
		}
		if err := tx.WithContext(txCtx).Model(newStrategy).Association(do.PromStrategyPreloadFieldCategories).Replace(categories); err != nil {
			return err
		}

		// 更新策略组的策略数量
		if err := l.strategyGroupRepo.UpdateStrategyCount(txCtx, strategyBO.GroupId); err != nil {
			return err
		}

		// 更新策略组的启用策略数量
		if err := l.strategyGroupRepo.UpdateEnableStrategyCount(txCtx, strategyBO.GroupId); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	go func() {
		defer after.Recover(l.log)
		l.changeGroupChannel <- strategyBO.GroupId
	}()

	return bo.StrategyModelToBO(newStrategy), nil
}

// UpdateStrategyById 更新策略 TODO 需要增加事物, 保证数据一致性
func (l *strategyRepoImpl) UpdateStrategyById(ctx context.Context, id uint32, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
	detail, err := l.getStrategyById(ctx, id)
	if err != nil {
		return nil, err
	}

	newStrategy := strategyBO.ToModel()
	newStrategy.ID = detail.ID
	newStrategy.Status = detail.Status
	// 替换报警页面和分类
	alarmPages := slices.To(strategyBO.AlarmPageIds, func(pageId uint32) *do.SysDict {
		return &do.SysDict{
			BaseModel: do.BaseModel{ID: pageId},
		}
	})
	categories := slices.To(strategyBO.CategoryIds, func(categoryId uint32) *do.SysDict {
		return &do.SysDict{
			BaseModel: do.BaseModel{ID: categoryId},
		}
	})

	var groupIds []uint32
	if detail.GroupID != strategyBO.GroupId {
		groupIds = []uint32{detail.GroupID, newStrategy.GroupID}
	}

	newStrategyMap := newStrategy.ToMap()
	err = l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := basescopes.WithTx(ctx, tx)
		if err = tx.WithContext(txCtx).Model(detail).Scopes(basescopes.InIds(id)).Updates(newStrategyMap).Error; err != nil {
			return err
		}

		if err = tx.WithContext(txCtx).Model(detail).Association(do.PromStrategyPreloadFieldAlarmPages).Clear(); err != nil {
			return err
		}
		if err = tx.WithContext(txCtx).Model(detail).Association(do.PromStrategyPreloadFieldAlarmPages).Replace(&alarmPages); err != nil {
			return err
		}
		if err = tx.WithContext(txCtx).Model(detail).Association(do.PromStrategyPreloadFieldCategories).Clear(); err != nil {
			return err
		}
		if err = tx.WithContext(txCtx).Model(detail).Association(do.PromStrategyPreloadFieldCategories).Replace(&categories); err != nil {
			return err
		}

		// 如果策略组发生变更
		if len(groupIds) > 0 {
			// 更新策略组的启用策略数量
			if err = l.strategyGroupRepo.UpdateStrategyCount(txCtx, groupIds...); err != nil {
				return err
			}

			if err = l.strategyGroupRepo.UpdateEnableStrategyCount(txCtx, groupIds...); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	go func() {
		defer after.Recover(l.log)
		l.changeGroupChannel <- detail.GroupID
		l.changeGroupChannel <- strategyBO.GroupId
	}()

	return bo.StrategyModelToBO(newStrategy), nil
}

func (l *strategyRepoImpl) BatchUpdateStrategyStatusByIds(ctx context.Context, status vobj.Status, ids []uint32) error {
	// 查询规则组ID列表
	groupIds, err := l.getStrategyGroupIdsByStrategyIds(ctx, ids)
	if err != nil {
		return err
	}
	err = l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := basescopes.WithTx(ctx, tx)
		if err = tx.WithContext(txCtx).Scopes(basescopes.InIds(ids...)).Updates(&do.PromStrategy{Status: status}).Error; err != nil {
			return err
		}
		// 更新策略组的启用策略数量
		if err = l.strategyGroupRepo.UpdateEnableStrategyCount(txCtx, groupIds...); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	go func() {
		defer after.Recover(l.log)
		for _, groupId := range groupIds {
			l.changeGroupChannel <- groupId
		}
	}()
	return nil
}

// getStrategyGroupIds 获取策略组ID列表
func (l *strategyRepoImpl) getStrategyGroupIdsByStrategyIds(ctx context.Context, ids []uint32) ([]uint32, error) {
	// 查询规则组ID列表
	var groupIds []uint32
	field := do.PromStrategyFieldGroupID
	whereList := []basescopes.ScopeMethod{
		basescopes.InIds(ids...),
		basescopes.WithCreateBy(ctx),
	}
	if err := l.data.DB().WithContext(ctx).
		Model(&do.PromStrategy{}).
		Scopes(whereList...).
		Select(field).
		Pluck(field, &groupIds).Error; err != nil {
		return nil, err
	}
	return groupIds, nil
}

func (l *strategyRepoImpl) DeleteStrategyByIds(ctx context.Context, ids ...uint32) error {
	// 查询规则组ID列表
	groupIds, err := l.getStrategyGroupIdsByStrategyIds(ctx, ids)
	if err != nil {
		return err
	}
	whereList := []basescopes.ScopeMethod{
		basescopes.InIds(ids...),
		basescopes.WithCreateBy(ctx),
	}

	err = l.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := basescopes.WithTx(ctx, tx)
		if err = tx.WithContext(txCtx).Scopes(whereList...).Delete(&do.PromStrategy{}).Error; err != nil {
			return err
		}

		// 更新策略组的策略数量
		if err = l.strategyGroupRepo.UpdateStrategyCount(txCtx, groupIds...); err != nil {
			return err
		}

		// 更新策略组的启用策略数量
		if err = l.strategyGroupRepo.UpdateEnableStrategyCount(txCtx, groupIds...); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	go func() {
		defer after.Recover(l.log)
		for _, groupId := range groupIds {
			l.changeGroupChannel <- groupId
		}
	}()

	return nil
}

func (l *strategyRepoImpl) GetStrategyById(ctx context.Context, id uint32, wheres ...basescopes.ScopeMethod) (*bo.StrategyBO, error) {
	whereList := append(wheres, basescopes.WithCreateBy(ctx))
	firstStrategy, err := l.getStrategyById(ctx, id, whereList...)
	if err != nil {
		return nil, err
	}
	return bo.StrategyModelToBO(firstStrategy), nil
}

func (l *strategyRepoImpl) getStrategyById(ctx context.Context, id uint32, wheres ...basescopes.ScopeMethod) (*do.PromStrategy, error) {
	var first do.PromStrategy
	whereList := append(wheres, basescopes.WithCreateBy(ctx), basescopes.InIds(id))
	if err := l.data.DB().WithContext(ctx).Scopes(whereList...).First(&first).Error; err != nil {
		return nil, err
	}
	return &first, nil
}

func (l *strategyRepoImpl) ListStrategy(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyBO, error) {
	var listStrategy []*do.PromStrategy
	whereList := append(scopes, basescopes.WithCreateBy(ctx))
	if err := l.data.DB().WithContext(ctx).Scopes(append(whereList, bo.Page(pgInfo))...).Find(&listStrategy).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.PromStrategy{}).Scopes(whereList...).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}
	list := slices.To(listStrategy, func(i *do.PromStrategy) *bo.StrategyBO {
		return bo.StrategyModelToBO(i)
	})
	return list, nil
}

func NewStrategyRepo(
	data *data.Data,
	changeGroupChannel chan<- uint32,
	strategyGroupRepo repository.StrategyGroupRepo,
	logger log.Logger,
) repository.StrategyRepo {
	return &strategyRepoImpl{
		data:               data,
		log:                log.NewHelper(logger),
		strategyGroupRepo:  strategyGroupRepo,
		changeGroupChannel: changeGroupChannel,
	}
}
