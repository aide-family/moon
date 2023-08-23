package data

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"

	"prometheus-manager/api"
	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"
	nodeV1Pull "prometheus-manager/api/strategy/v1/pull"

	buildQuery "prometheus-manager/pkg/build_query"
	"prometheus-manager/pkg/conn"
	"prometheus-manager/pkg/dal/model"
	"prometheus-manager/pkg/dal/query"
	"prometheus-manager/pkg/util/stringer"

	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/apps/master/internal/conf"
)

type (
	DictV1Repo struct {
		logger *log.Helper
		data   *Data
		db     *query.Query
	}
)

var _ biz.IDictV1Repo = (*DictV1Repo)(nil)

func NewDictRepo(data *Data, logger log.Logger) *DictV1Repo {
	return &DictV1Repo{data: data, db: query.Use(data.DB()), logger: log.NewHelper(log.With(logger, "module", dictModuleName))}
}

func (l *DictV1Repo) V1(_ context.Context) string {
	return "DictV1Repo.V1"
}

func (l *DictV1Repo) CreateDict(ctx context.Context, m *model.PromDict) error {
	ctx, span := otel.Tracer(dictModuleName).Start(ctx, "DictV1Repo.CreateDict")
	defer span.End()
	if m == nil {
		return perrors.ErrorServerDataNotFound("PromDict is nil")
	}

	promDict := l.db.PromDict

	return promDict.WithContext(ctx).Create(m)
}

func (l *DictV1Repo) UpdateDictById(ctx context.Context, id int32, m *model.PromDict) error {
	ctx, span := otel.Tracer(dictModuleName).Start(ctx, "DictV1Repo.UpdateDictById")
	defer span.End()
	if m == nil {
		return perrors.ErrorServerDataNotFound("PromDict is nil")
	}

	promDict := l.db.PromDict

	inf, err := promDict.WithContext(ctx).Where(promDict.ID.Eq(id)).UpdateColumnSimple(
		promDict.Name.Value(m.Name),
		promDict.Remark.Value(m.Remark),
		promDict.Color.Value(m.Color),
		promDict.Category.Value(m.Category),
	)
	if err != nil {
		l.logger.WithContext(ctx).Errorw("UpdateDictById", id, "err", err)
		return perrors.ErrorServerDatabaseError("UpdateDictById err").WithCause(err).WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}
	if inf.RowsAffected != 1 {
		return perrors.ErrorClientNotFound("PromDict not found").WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	return nil
}

func (l *DictV1Repo) UpdateDictByIds(ctx context.Context, ids []int32, status prom.Status) error {
	ctx, span := otel.Tracer(dictModuleName).Start(ctx, "DictV1Repo.UpdateDictByIds")
	defer span.End()

	promDict := l.db.PromDict

	_, err := promDict.WithContext(ctx).Where(promDict.ID.In(ids...)).UpdateColumnSimple(promDict.Status.Value(int32(status)))
	if err != nil {
		l.logger.WithContext(ctx).Errorw("UpdateDictByIds", ids, "err", err)
		return perrors.ErrorServerDatabaseError("UpdateDictByIds err").WithCause(err).WithMetadata(map[string]string{
			"ids": stringer.New(ids).String(),
		})
	}

	return nil
}

func (l *DictV1Repo) DeleteDictById(ctx context.Context, id int32) error {
	ctx, span := otel.Tracer(dictModuleName).Start(ctx, "DictV1Repo.DeleteDictById")
	defer span.End()

	promDict := l.db.PromDict

	inf, err := promDict.WithContext(ctx).Where(promDict.ID.Eq(id)).Delete()
	if err != nil {
		l.logger.WithContext(ctx).Errorw("DeleteDictById", id, "err", err)
		return perrors.ErrorServerDatabaseError("DeleteDictById err").WithCause(err).WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	if inf.RowsAffected != 1 {
		return perrors.ErrorClientNotFound("PromDict not found").WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	return nil
}

func (l *DictV1Repo) GetDictById(ctx context.Context, id int32) (*model.PromDict, error) {
	ctx, span := otel.Tracer(dictModuleName).Start(ctx, "DictV1Repo.GetDictById")
	defer span.End()

	promDict := l.db.PromDict

	detail, err := promDict.WithContext(ctx).Where(promDict.ID.Eq(id)).First()
	if err != nil {
		l.logger.WithContext(ctx).Errorw("GetDictById", id, "err", err)
		return nil, perrors.ErrorServerDatabaseError("GetDictById err").WithCause(err).WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	return detail, nil
}

func (l *DictV1Repo) ListDict(ctx context.Context, req *pb.ListDictRequest) ([]*model.PromDict, int64, error) {
	ctx, span := otel.Tracer(dictModuleName).Start(ctx, "DictV1Repo.ListDict")
	defer span.End()

	promDict := l.db.PromDict
	offset, limit := buildQuery.GetPage(req.GetQuery().GetPage())
	promDictDB := promDict.WithContext(ctx)

	if req != nil {
		queryPrams := req.GetQuery()
		if queryPrams != nil {
			sorts := queryPrams.GetSort()
			iSorts := make([]buildQuery.ISort, 0, len(sorts))
			for _, sort := range sorts {
				iSorts = append(iSorts, sort)
			}
			promDictDB = promDictDB.Order(buildQuery.GetSorts(&promDict, iSorts...)...)
			promDictDB = promDictDB.Select(buildQuery.GetSlectExprs(&promDict, queryPrams)...)
			keyword := queryPrams.GetKeyword()
			if keyword != "" {
				key := "%" + keyword + "%"
				promDictDB = promDictDB.Where(buildQuery.GetConditionKeywords(key, promDict.Name)...)
			}
			if queryPrams.GetStartAt() > 0 && queryPrams.GetEndAt() > 0 {
				promDictDB = promDictDB.Where(promDict.CreatedAt.Between(
					time.Unix(queryPrams.GetStartAt(), 0),
					time.Unix(queryPrams.GetEndAt(), 0),
				))
			}
		}

		dictQuery := req.GetDict()
		if dictQuery != nil {
			if dictQuery.GetId() > 0 {
				promDictDB = promDictDB.Where(promDict.ID.Eq(dictQuery.GetId()))
			}
			if dictQuery.GetName() != "" {
				promDictDB = promDictDB.Where(promDict.Name.Eq(dictQuery.GetName()))
			}
			if dictQuery.GetColor() != "" {
				promDictDB = promDictDB.Where(promDict.Color.Eq(dictQuery.GetColor()))
			}
			if dictQuery.GetCategory() != prom.Category_CATEGORY_NONE {
				promDictDB = promDictDB.Where(promDict.Category.Eq(int32(dictQuery.GetCategory())))
			}
			if dictQuery.GetStatus() != prom.Status_Status_NONE {
				promDictDB = promDictDB.Where(promDict.Status.Eq(int32(dictQuery.GetStatus())))
			}
		}
	}

	return promDictDB.FindByPage(int(offset), int(limit))
}

func (l *DictV1Repo) Datasources(ctx context.Context, req *pb.DatasourcesRequest) (*pb.DatasourcesReply, error) {
	ctx, span := otel.Tracer(dictModuleName).Start(ctx, "DictV1Repo.Datasources")
	defer span.End()
	nodes := conf.Get().GetPushStrategy().GetNodes()
	var err error
	var eg errgroup.Group
	datasourceList := make([]*api.Datasource, 0)
	var lock sync.Mutex
	for _, node := range nodes {
		if node.Network != conn.NetworkGrpc {
			continue
		}
		newNode := node
		eg.Go(func() error {
			rpcConn, ok := l.data.nodeGrpcClients[newNode]
			if !ok {
				rpcConn, err = conn.GetNodeGrpcClient(ctx, newNode, conn.GetDiscovery())
				if err != nil {
					l.logger.WithContext(ctx).Warnw("GRPCPushCall", newNode, "err", err)
					return nil
				}
			}

			resp, err := nodeV1Pull.NewPullClient(rpcConn).Datasources(ctx, &nodeV1Pull.DatasourcesRequest{Node: newNode.GetServerName()})
			if err != nil {
				l.logger.WithContext(ctx).Warnw("GRPCPushCall", newNode, "err", err)
				return nil
			}
			lock.Lock()
			defer lock.Unlock()
			datasourceList = append(datasourceList, resp.GetDatasource()...)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		l.logger.WithContext(ctx).Errorw("Datasource", req, "err", err)
		return nil, perrors.ErrorServerUnknown("Datasources err").WithCause(err).WithMetadata(map[string]string{
			"err": err.Error(),
		})
	}

	return &pb.DatasourcesReply{Response: &api.Response{Message: l.V1(ctx)}, Datasources: datasourceList}, nil
}
