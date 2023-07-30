package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"gorm.io/gen"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
)

type (
	NodeRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *NodeRepo) CreateNode(ctx context.Context, m *model.PromNode) error {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.CreateNode")
	defer span.End()
	return query.Use(l.data.DB()).WithContext(ctx).PromNode.Create(m)
}

func (l *NodeRepo) UpdateNodeById(ctx context.Context, id uint32, m *model.PromNode) error {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.UpdateNodeById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNode.Where(query.PromNode.ID.Eq(int32(id))).Updates(m); err != nil {
		return err
	}

	return nil
}

func (l *NodeRepo) DeleteNodeById(ctx context.Context, id uint32) error {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.DeleteNodeById")
	defer span.End()

	if _, err := query.Use(l.data.DB()).WithContext(ctx).PromNode.Where(query.PromNode.ID.Eq(int32(id))).Delete(); err != nil {
		return err
	}

	return nil
}

func (l *NodeRepo) GetNodeById(ctx context.Context, id uint32) (*model.PromNode, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.GetNodeById")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNode.FindById(ctx, int32(id))
}

func (l *NodeRepo) ListNode(ctx context.Context, q *promBizV1.NodeListQueryParams) ([]*model.PromNode, int64, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.ListNode")
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNode.Scopes(
		func(dao gen.Dao) gen.Dao {
			if q.Keyword != "" {
				dao = dao.Or(
					query.PromNode.EnName.Like("%"+q.Keyword),
					query.PromNode.ChName.Like("%"+q.Keyword),
					query.PromNode.Remark.Like("%"+q.Keyword+"%"),
				)
			}
			return dao
		},
	).FindByPage(q.Offset, q.Limit)
}

func (l *NodeRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.V1")
	defer span.End()
	return "node v1"
}

var _ promBizV1.INodeRepo = (*NodeRepo)(nil)

func NewNodeRepo(data *Data, logger log.Logger) *NodeRepo {
	return &NodeRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Node"))}
}
