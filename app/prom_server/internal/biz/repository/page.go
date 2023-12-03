package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/pkg/helper/valueobj"
)

var _ PageRepo = (*UnimplementedPageRepo)(nil)

type (
	PageRepo interface {
		mustEmbedUnimplemented()
		// CreatePage 创建页面
		CreatePage(ctx context.Context, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error)
		// UpdatePageById 通过id更新页面
		UpdatePageById(ctx context.Context, id uint, pageBO *bo.AlarmPageBO) (*bo.AlarmPageBO, error)
		// BatchUpdatePageStatusByIds 通过id批量更新页面状态
		BatchUpdatePageStatusByIds(ctx context.Context, status valueobj.Status, ids []uint) error
		// DeletePageByIds 通过id删除页面
		DeletePageByIds(ctx context.Context, id ...uint) error
		// GetPageById 通过id获取页面详情
		GetPageById(ctx context.Context, id uint) (*bo.AlarmPageBO, error)
		// ListPage 获取页面列表
		ListPage(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.AlarmPageBO, error)
	}

	UnimplementedPageRepo struct{}
)

func (UnimplementedPageRepo) mustEmbedUnimplemented() {}

func (UnimplementedPageRepo) CreatePage(_ context.Context, _ *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePage not implemented")
}

func (UnimplementedPageRepo) UpdatePageById(_ context.Context, _ uint, _ *bo.AlarmPageBO) (*bo.AlarmPageBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePageById not implemented")
}

func (UnimplementedPageRepo) BatchUpdatePageStatusByIds(_ context.Context, _ valueobj.Status, _ []uint) error {
	return status.Errorf(codes.Unimplemented, "method BatchUpdatePageStatusByIds not implemented")
}

func (UnimplementedPageRepo) DeletePageByIds(_ context.Context, _ ...uint) error {
	return status.Errorf(codes.Unimplemented, "method DeletePageByIds not implemented")
}

func (UnimplementedPageRepo) GetPageById(_ context.Context, _ uint) (*bo.AlarmPageBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPageById not implemented")
}

func (UnimplementedPageRepo) ListPage(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*bo.AlarmPageBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPage not implemented")
}
