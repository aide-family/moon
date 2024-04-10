package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
)

var _ PromDictRepo = (*UnimplementedPromDictRepo)(nil)

type (
	PromDictRepo interface {
		mustEmbedUnimplemented()
		// CreateDict 创建字典
		CreateDict(ctx context.Context, dict *bo.DictBO) (*bo.DictBO, error)
		// UpdateDictById 通过id更新字典
		UpdateDictById(ctx context.Context, id uint32, dict *bo.DictBO) (*bo.DictBO, error)
		// BatchUpdateDictStatusByIds 通过id批量更新字典状态
		BatchUpdateDictStatusByIds(ctx context.Context, status vobj.Status, ids []uint32) error
		// DeleteDictByIds 通过id删除字典
		DeleteDictByIds(ctx context.Context, id ...uint32) error
		// GetDictById 通过id获取字典详情
		GetDictById(ctx context.Context, id uint32) (*bo.DictBO, error)
		GetDictByIds(ctx context.Context, ids ...uint32) ([]*bo.DictBO, error)
		// ListDict 获取字典列表
		ListDict(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.DictBO, error)
	}

	UnimplementedPromDictRepo struct{}
)

func (UnimplementedPromDictRepo) GetDictByIds(_ context.Context, _ ...uint32) ([]*bo.DictBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDictByIds not implemented")
}

func (UnimplementedPromDictRepo) mustEmbedUnimplemented() {}

func (UnimplementedPromDictRepo) CreateDict(_ context.Context, _ *bo.DictBO) (*bo.DictBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDict not implemented")
}

func (UnimplementedPromDictRepo) UpdateDictById(_ context.Context, _ uint32, _ *bo.DictBO) (*bo.DictBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDictById not implemented")
}

func (UnimplementedPromDictRepo) BatchUpdateDictStatusByIds(_ context.Context, _ vobj.Status, _ []uint32) error {
	return status.Errorf(codes.Unimplemented, "method BatchUpdateDictStatusByIds not implemented")
}

func (UnimplementedPromDictRepo) DeleteDictByIds(_ context.Context, _ ...uint32) error {
	return status.Errorf(codes.Unimplemented, "method DeleteDictByIds not implemented")
}

func (UnimplementedPromDictRepo) GetDictById(_ context.Context, _ uint32) (*bo.DictBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDictById not implemented")
}

func (UnimplementedPromDictRepo) ListDict(_ context.Context, _ bo.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.DictBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDict not implemented")
}
