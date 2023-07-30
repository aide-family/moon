package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gen"
	"prometheus-manager/api/perrors"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
	"prometheus-manager/pkg/util/stringer"
)

type (
	NodeRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *NodeRepo) CreateNode(ctx context.Context, m *model.PromNode) error {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.CreateNode")
	span.SetAttributes(attribute.Stringer("model", stringer.New(m)))
	defer span.End()
	return query.Use(l.data.DB()).WithContext(ctx).PromNode.Create(m)
}

func (l *NodeRepo) UpdateNodeById(ctx context.Context, id uint32, m *model.PromNode) error {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.UpdateNodeById")
	span.SetAttributes(attribute.Stringer("model", stringer.New(m)))
	defer span.End()

	if id == 0 || m == nil {
		return perrors.ErrorServerDataNotFound("node id is not found")
	}

	modelInstance := query.Use(l.data.DB()).PromNode
	db := modelInstance.WithContext(ctx)
	if _, err := db.Where(modelInstance.ID.Eq(int32(id))).Updates(m); err != nil {
		return perrors.ErrorServerDatabaseError("update node error: %v", err)
	}

	return nil
}

func (l *NodeRepo) DeleteNodeById(ctx context.Context, id uint32) error {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.DeleteNodeById")
	span.SetAttributes(attribute.Int64("id", int64(id)))
	defer span.End()

	modelInstance := query.Use(l.data.DB()).PromNode
	db := modelInstance.WithContext(ctx)

	if _, err := db.Where(modelInstance.ID.Eq(int32(id))).Delete(); err != nil {
		return err
	}

	return nil
}

func (l *NodeRepo) GetNodeById(ctx context.Context, id uint32) (*model.PromNode, error) {
	ctx, span := otel.Tracer("data").Start(ctx, "NodeRepo.GetNodeById")
	span.SetAttributes(attribute.Int64("id", int64(id)))
	defer span.End()

	return query.Use(l.data.DB()).WithContext(ctx).PromNode.FindById(ctx, int32(id))
}

func (l *NodeRepo) ListNode(ctx context.Context, q *promBizV1.NodeListQueryParams) ([]*model.PromNode, int64, error) {
	ctx, span := otel.Tracer("query").Start(ctx, "NodeRepo.ListNode")
	span.SetAttributes(attribute.Stringer("id", stringer.New(q)))
	defer span.End()

	modelInstance := query.Use(l.data.DB()).PromNode
	db := modelInstance.WithContext(ctx)

	return db.Scopes(
		func(dao gen.Dao) gen.Dao {
			if q.Keyword != "" {
				dao = dao.Or(
					modelInstance.EnName.Like("%"+q.Keyword),
					modelInstance.ChName.Like("%"+q.Keyword),
					modelInstance.Remark.Like("%"+q.Keyword+"%"),
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
