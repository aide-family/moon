package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ PromDictRepo = (*UnimplementedPromDictRepo)(nil)

type (
	PromDictRepo interface {
		mustEmbedUnimplemented()
		// CreateDict 创建字典
		CreateDict(ctx context.Context, dict *dobo.DictDO) (*dobo.DictDO, error)
		// UpdateDictById 通过id更新字典
		UpdateDictById(ctx context.Context, id uint, dict *dobo.DictDO) (*dobo.DictDO, error)
		// BatchUpdateDictStatusByIds 通过id批量更新字典状态
		BatchUpdateDictStatusByIds(ctx context.Context, status int32, ids []uint) error
		// DeleteDictByIds 通过id删除字典
		DeleteDictByIds(ctx context.Context, id ...uint) error
		// GetDictById 通过id获取字典详情
		GetDictById(ctx context.Context, id uint) (*dobo.DictDO, error)
		// ListDict 获取字典列表
		ListDict(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.DictDO, error)
	}

	UnimplementedPromDictRepo struct{}
)

func (UnimplementedPromDictRepo) mustEmbedUnimplemented() {}

func (UnimplementedPromDictRepo) CreateDict(_ context.Context, _ *dobo.DictDO) (*dobo.DictDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDict not implemented")
}

func (UnimplementedPromDictRepo) UpdateDictById(_ context.Context, _ uint, _ *dobo.DictDO) (*dobo.DictDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDictById not implemented")
}

func (UnimplementedPromDictRepo) BatchUpdateDictStatusByIds(_ context.Context, _ int32, _ []uint) error {
	return status.Errorf(codes.Unimplemented, "method BatchUpdateDictStatusByIds not implemented")
}

func (UnimplementedPromDictRepo) DeleteDictByIds(_ context.Context, _ ...uint) error {
	return status.Errorf(codes.Unimplemented, "method DeleteDictByIds not implemented")
}

func (UnimplementedPromDictRepo) GetDictById(_ context.Context, _ uint) (*dobo.DictDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDictById not implemented")
}

func (UnimplementedPromDictRepo) ListDict(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*dobo.DictDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDict not implemented")
}
