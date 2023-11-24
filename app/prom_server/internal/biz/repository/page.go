package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ PageRepo = (*UnimplementedPageRepo)(nil)

type (
	PageRepo interface {
		mustEmbedUnimplemented()
		// CreatePage 创建页面
		CreatePage(ctx context.Context, pageDo *dobo.AlarmPageDO) (*dobo.AlarmPageDO, error)
		// UpdatePageById 通过id更新页面
		UpdatePageById(ctx context.Context, id uint, pageDo *dobo.AlarmPageDO) (*dobo.AlarmPageDO, error)
		// BatchUpdatePageStatusByIds 通过id批量更新页面状态
		BatchUpdatePageStatusByIds(ctx context.Context, status int32, ids []uint) error
		// DeletePageByIds 通过id删除页面
		DeletePageByIds(ctx context.Context, id ...uint) error
		// GetPageById 通过id获取页面详情
		GetPageById(ctx context.Context, id uint) (*dobo.AlarmPageDO, error)
		// ListPage 获取页面列表
		ListPage(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.AlarmPageDO, error)
	}

	UnimplementedPageRepo struct{}
)

func (UnimplementedPageRepo) mustEmbedUnimplemented() {}

func (UnimplementedPageRepo) CreatePage(_ context.Context, _ *dobo.AlarmPageDO) (*dobo.AlarmPageDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePage not implemented")
}

func (UnimplementedPageRepo) UpdatePageById(_ context.Context, _ uint, _ *dobo.AlarmPageDO) (*dobo.AlarmPageDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePageById not implemented")
}

func (UnimplementedPageRepo) BatchUpdatePageStatusByIds(_ context.Context, _ int32, _ []uint) error {
	return status.Errorf(codes.Unimplemented, "method BatchUpdatePageStatusByIds not implemented")
}

func (UnimplementedPageRepo) DeletePageByIds(_ context.Context, _ ...uint) error {
	return status.Errorf(codes.Unimplemented, "method DeletePageByIds not implemented")
}

func (UnimplementedPageRepo) GetPageById(_ context.Context, _ uint) (*dobo.AlarmPageDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPageById not implemented")
}

func (UnimplementedPageRepo) ListPage(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*dobo.AlarmPageDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPage not implemented")
}
