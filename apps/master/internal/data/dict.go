package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"
	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
	buildQuery "prometheus-manager/pkg/build_query"
	"strconv"
	"time"
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
	return &DictV1Repo{data: data, db: query.Use(data.DB()), logger: log.NewHelper(log.With(logger, "module", "data/Dict"))}
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
		}
	}

	return promDictDB.FindByPage(int(offset), int(limit))
}
