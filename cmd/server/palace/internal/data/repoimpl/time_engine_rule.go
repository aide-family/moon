package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen/field"
)

// NewTimeEngineRuleRepository 创建时间引擎规则仓库
func NewTimeEngineRuleRepository(data *data.Data) repository.TimeEngineRule {
	return &timeEngineRuleRepositoryImpl{data: data}
}

type timeEngineRuleRepositoryImpl struct {
	data *data.Data
}

// CreateTimeEngineRule 创建时间引擎规则
func (t *timeEngineRuleRepositoryImpl) CreateTimeEngineRule(ctx context.Context, req *bizmodel.TimeEngineRule) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	return bizQuery.TimeEngineRule.WithContext(ctx).Create(req)
}

// DeleteTimeEngineRule 删除时间引擎规则
func (t *timeEngineRuleRepositoryImpl) DeleteTimeEngineRule(ctx context.Context, id uint32) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.TimeEngineRule.WithContext(ctx).Where(bizQuery.TimeEngineRule.ID.Eq(id)).Delete()
	return err
}

// GetTimeEngineRule 获取时间引擎规则
func (t *timeEngineRuleRepositoryImpl) GetTimeEngineRule(ctx context.Context, id uint32) (*bizmodel.TimeEngineRule, error) {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return nil, err
	}
	return bizQuery.TimeEngineRule.WithContext(ctx).Where(bizQuery.TimeEngineRule.ID.Eq(id)).First()
}

// ListTimeEngineRule 获取时间引擎规则列表
func (t *timeEngineRuleRepositoryImpl) ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) ([]*bizmodel.TimeEngineRule, error) {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return nil, err
	}
	wrapper := bizQuery.TimeEngineRule.WithContext(ctx)
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(bizQuery.TimeEngineRule.Status.Eq(req.Status.GetValue()))
	}
	if req.Keyword != "" {
		wrapper = wrapper.Where(bizQuery.TimeEngineRule.Name.Like(req.Keyword))
	}
	if !req.Category.IsUnknown() {
		wrapper = wrapper.Where(bizQuery.TimeEngineRule.Category.Eq(req.Category.GetValue()))
	}
	if wrapper, err = types.WithPageQuery(wrapper, req.Page); err != nil {
		return nil, err
	}
	return wrapper.Order(bizQuery.TimeEngineRule.ID.Desc()).Find()
}

// UpdateTimeEngineRule 更新时间引擎规则
func (t *timeEngineRuleRepositoryImpl) UpdateTimeEngineRule(ctx context.Context, req *bizmodel.TimeEngineRule) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.TimeEngineRule.WithContext(ctx).Where(bizQuery.TimeEngineRule.ID.Eq(req.ID)).Updates(req)
	return err
}

// BatchUpdateTimeEngineRuleStatus 批量更新时间引擎规则状态
func (t *timeEngineRuleRepositoryImpl) BatchUpdateTimeEngineRuleStatus(ctx context.Context, req *bo.BatchUpdateTimeEngineRuleStatusRequest) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.TimeEngineRule.WithContext(ctx).Where(
		bizQuery.TimeEngineRule.ID.In(req.IDs...),
		bizQuery.TimeEngineRule.Status.Neq(req.Status.GetValue()),
	).Update(bizQuery.TimeEngineRule.Status, req.Status.GetValue())
	return err
}

// CreateTimeEngine 创建时间引擎
func (t *timeEngineRuleRepositoryImpl) CreateTimeEngine(ctx context.Context, req *bizmodel.TimeEngine) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}

	if err := bizQuery.TimeEngine.WithContext(ctx).Omit(field.AssociationFields).Create(req); err != nil {
		return err
	}

	return bizQuery.TimeEngine.Rules.WithContext(ctx).Model(req).Append(req.Rules...)
}

// UpdateTimeEngine 更新时间引擎
func (t *timeEngineRuleRepositoryImpl) UpdateTimeEngine(ctx context.Context, req *bizmodel.TimeEngine) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	if err := bizQuery.TimeEngine.WithContext(ctx).Omit(field.AssociationFields).Where(bizQuery.TimeEngine.ID.Eq(req.ID)).Save(req); err != nil {
		return err
	}
	return bizQuery.TimeEngine.Rules.Model(req).Replace(req.Rules...)
}

// BatchUpdateTimeEngineStatus 批量更新时间引擎状态
func (t *timeEngineRuleRepositoryImpl) BatchUpdateTimeEngineStatus(ctx context.Context, req *bo.BatchUpdateTimeEngineStatusRequest) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.TimeEngine.WithContext(ctx).Where(
		bizQuery.TimeEngine.ID.In(req.IDs...),
		bizQuery.TimeEngine.Status.Neq(req.Status.GetValue()),
	).Update(bizQuery.TimeEngine.Status, req.Status.GetValue())
	return err
}

// DeleteTimeEngine 删除时间引擎
func (t *timeEngineRuleRepositoryImpl) DeleteTimeEngine(ctx context.Context, id uint32) error {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return err
	}
	_, err = bizQuery.TimeEngine.WithContext(ctx).Where(bizQuery.TimeEngine.ID.Eq(id)).Delete()
	return err
}

// GetTimeEngine 获取时间引擎
func (t *timeEngineRuleRepositoryImpl) GetTimeEngine(ctx context.Context, id uint32) (*bizmodel.TimeEngine, error) {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return nil, err
	}
	return bizQuery.TimeEngine.WithContext(ctx).Where(bizQuery.TimeEngine.ID.Eq(id)).Preload(field.Associations).First()
}

// ListTimeEngine 获取时间引擎列表
func (t *timeEngineRuleRepositoryImpl) ListTimeEngine(ctx context.Context, req *bo.ListTimeEngineRequest) ([]*bizmodel.TimeEngine, error) {
	bizQuery, err := getBizQuery(ctx, t.data)
	if err != nil {
		return nil, err
	}
	wrapper := bizQuery.TimeEngine.WithContext(ctx)
	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(bizQuery.TimeEngine.Status.Eq(req.Status.GetValue()))
	}
	if req.Keyword != "" {
		wrapper = wrapper.Where(bizQuery.TimeEngine.Name.Like(req.Keyword))
	}
	if wrapper, err = types.WithPageQuery(wrapper, req.Page); err != nil {
		return nil, err
	}
	return wrapper.Order(bizQuery.TimeEngine.ID.Desc()).Find()
}
