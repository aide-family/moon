package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
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
	return bizQuery.TimeEngineRule.WithContext(ctx).Find()
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
