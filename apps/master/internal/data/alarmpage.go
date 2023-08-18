package data

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"

	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"

	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
	buildQuery "prometheus-manager/pkg/build_query"
	"prometheus-manager/pkg/util/stringer"

	"prometheus-manager/apps/master/internal/biz"
)

type (
	AlarmPageV1Repo struct {
		logger *log.Helper
		data   *Data
		db     *query.Query
	}
)

var _ biz.IAlarmPageV1Repo = (*AlarmPageV1Repo)(nil)

func NewAlarmPageV1Repo(data *Data, logger log.Logger) *AlarmPageV1Repo {
	return &AlarmPageV1Repo{data: data, db: query.Use(data.DB()), logger: log.NewHelper(log.With(logger, "module", alarmPageModuleName))}
}

func (l *AlarmPageV1Repo) V1(_ context.Context) string {
	return "AlarmPageV1Repo.V1"
}

func (l *AlarmPageV1Repo) CreateAlarmPage(ctx context.Context, m *model.PromAlarmPage) error {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageV1Repo.CreateAlarmPage")
	defer span.End()
	if m == nil {
		return perrors.ErrorServerDataNotFound("PromAlarmPage is nil")
	}

	promAlarmPage := l.db.PromAlarmPage

	return promAlarmPage.WithContext(ctx).Create(m)
}

func (l *AlarmPageV1Repo) UpdateAlarmPageById(ctx context.Context, id int32, m *model.PromAlarmPage) error {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageV1Repo.UpdateAlarmPageById")
	defer span.End()
	if m == nil {
		return perrors.ErrorServerDataNotFound("PromAlarmPage is nil")
	}

	promAlarmPage := l.db.PromAlarmPage

	inf, err := promAlarmPage.WithContext(ctx).Where(promAlarmPage.ID.Eq(id)).UpdateColumnSimple(
		promAlarmPage.Name.Value(m.Name),
		promAlarmPage.Remark.Value(m.Remark),
		promAlarmPage.Icon.Value(m.Icon),
		promAlarmPage.Color.Value(m.Color),
	)
	if err != nil {
		l.logger.WithContext(ctx).Errorw("UpdateAlarmPageById", id, "err", err)
		return perrors.ErrorServerDatabaseError("UpdateAlarmPageById err").WithCause(err).WithMetadata(map[string]string{
			"model": stringer.New(m).String(),
			"id":    strconv.Itoa(int(id)),
		})
	}

	if inf.RowsAffected != 1 {
		l.logger.WithContext(ctx).Warnw("UpdateAlarmPageById", id, "err", err)
		return perrors.ErrorClientNotFound("alarmPage not found").WithMetadata(map[string]string{
			"model": stringer.New(m).String(),
			"id":    strconv.Itoa(int(id)),
		})
	}

	return nil
}

func (l *AlarmPageV1Repo) UpdateAlarmPagesStatusByIds(ctx context.Context, ids []int32, status prom.Status) error {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageV1Repo.UpdateAlarmPagesStatusById")
	defer span.End()

	promAlarmPage := l.db.PromAlarmPage

	_, err := promAlarmPage.WithContext(ctx).Where(promAlarmPage.ID.In(ids...)).UpdateColumnSimple(
		promAlarmPage.Status.Value(int32(status)),
	)
	if err != nil {
		l.logger.WithContext(ctx).Errorw("UpdateAlarmPagesStatusById", ids, "err", err)
		return perrors.ErrorServerDatabaseError("UpdateAlarmPagesStatusById err").WithCause(err).WithMetadata(map[string]string{
			"ids":    stringer.New(ids).String(),
			"status": status.String(),
		})
	}

	return nil
}

func (l *AlarmPageV1Repo) DeleteAlarmPageById(ctx context.Context, id int32) error {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageV1Repo.DeleteAlarmPageById")
	defer span.End()

	promAlarmPage := l.db.PromAlarmPage

	inf, err := promAlarmPage.WithContext(ctx).Where(promAlarmPage.ID.Eq(id)).Delete()
	if err != nil {
		l.logger.WithContext(ctx).Errorw("DeleteAlarmPageById", id, "err", err)
		return perrors.ErrorServerDatabaseError("DeleteAlarmPageById", err).WithCause(err).WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	if inf.RowsAffected != 1 {
		l.logger.WithContext(ctx).Warnw("DeleteAlarmPageById", id, "err", err)
		return perrors.ErrorClientNotFound("alarmPage not found").WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	return nil
}

func (l *AlarmPageV1Repo) GetAlarmPageById(ctx context.Context, req *pb.GetAlarmPageRequest) (*model.PromAlarmPage, error) {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageV1Repo.GetAlarmPageById")
	defer span.End()

	if req == nil {
		return nil, perrors.ErrorServerUnknown("GetAlarmPageById req is nil")
	}

	promAlarmPage := l.db.PromAlarmPage
	promStrategy := l.db.PromStrategy

	queryPrams := req.GetQuery()
	id := req.GetId()

	offset, limit := buildQuery.GetPage(queryPrams.GetPage())
	promStrategyDB := promAlarmPage.PromStrategies
	if queryPrams != nil {
		sorts := queryPrams.GetSort()
		iSorts := make([]buildQuery.ISort, 0, len(sorts))
		for _, sort := range sorts {
			iSorts = append(iSorts, sort)
		}
		promStrategyDB.Order(buildQuery.GetSorts(&promStrategy, iSorts...)...)
		promStrategyDB.Select(buildQuery.GetSlectExprs(&promStrategy, queryPrams)...)
		keyword := queryPrams.GetKeyword()
		if keyword != "" {
			key := "%" + keyword + "%"
			promStrategyDB.Where(buildQuery.GetExprKeywords(key, promStrategy.Alert)...)
		}
		if queryPrams.GetStartAt() > 0 && queryPrams.GetEndAt() > 0 {
			promStrategyDB.Where(promStrategy.CreatedAt.Between(
				time.Unix(queryPrams.GetStartAt(), 0),
				time.Unix(queryPrams.GetEndAt(), 0),
			))
		}
	}

	strategyQuery := req.GetStrategy()
	if strategyQuery != nil {
		if strategyQuery.GetId() != 0 {
			promStrategyDB.Where(promStrategy.ID.Eq(strategyQuery.GetId()))
		}
		if strategyQuery.GetAlert() != "" {
			promStrategyDB.Where(promStrategy.Alert.Eq(strategyQuery.GetAlert()))
		}
		if strategyQuery.GetGroupId() != 0 {
			promStrategyDB.Where(promStrategy.GroupID.Eq(strategyQuery.GetGroupId()))
		}
	}

	detail, err := promAlarmPage.WithContext(ctx).Preload(promStrategyDB.Offset(int(offset)).Limit(int(limit))).Where(promAlarmPage.ID.Eq(id)).First()
	if err != nil {
		l.logger.WithContext(ctx).Errorw("GetAlarmPageById", id, "err", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perrors.ErrorLogicDataNotFound("alarmPage not found").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
			})
		}
		return nil, perrors.ErrorServerDatabaseError("GetAlarmPageById", err).WithCause(err).WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	return detail, nil
}

func (l *AlarmPageV1Repo) ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) ([]*model.PromAlarmPage, int64, error) {
	ctx, span := otel.Tracer(alarmPageModuleName).Start(ctx, "AlarmPageV1Repo.ListAlarmPage")
	defer span.End()

	promAlarmPage := l.db.PromAlarmPage
	offset, limit := buildQuery.GetPage(req.GetQuery().GetPage())
	promAlarmPageDB := promAlarmPage.WithContext(ctx)
	if req != nil {
		queryPrams := req.GetQuery()
		if queryPrams != nil {
			sorts := queryPrams.GetSort()
			iSorts := make([]buildQuery.ISort, 0, len(sorts))
			for _, sort := range sorts {
				iSorts = append(iSorts, sort)
			}
			promAlarmPageDB = promAlarmPageDB.Order(buildQuery.GetSorts(&promAlarmPage, iSorts...)...)
			promAlarmPageDB = promAlarmPageDB.Select(buildQuery.GetSlectExprs(&promAlarmPage, queryPrams)...)
			keyword := queryPrams.GetKeyword()
			if keyword != "" {
				key := "%" + keyword + "%"
				promAlarmPageDB = promAlarmPageDB.Or(buildQuery.GetConditionKeywords(
					key, promAlarmPage.Name,
					promAlarmPage.Color,
					promAlarmPage.Remark,
				)...)
			}
			if queryPrams.GetStartAt() > 0 && queryPrams.GetEndAt() > 0 {
				promAlarmPageDB = promAlarmPageDB.Where(promAlarmPage.CreatedAt.Between(
					time.Unix(queryPrams.GetStartAt(), 0),
					time.Unix(queryPrams.GetEndAt(), 0),
				))
			}
		}

		promAlarmQuery := req.GetAlarmPage()
		if promAlarmQuery != nil {
			if promAlarmQuery.GetId() > 0 {
				promAlarmPageDB = promAlarmPageDB.Where(promAlarmPage.ID.Eq(promAlarmQuery.GetId()))
			}
			if promAlarmQuery.GetName() != "" {
				promAlarmPageDB = promAlarmPageDB.Where(promAlarmPage.Name.Eq(promAlarmQuery.GetName()))
			}
			if promAlarmQuery.GetRemark() != "" {
				promAlarmPageDB = promAlarmPageDB.Where(promAlarmPage.Remark.Eq(promAlarmQuery.GetRemark()))
			}
			if promAlarmQuery.GetIcon() != "" {
				promAlarmPageDB = promAlarmPageDB.Where(promAlarmPage.Icon.Eq(promAlarmQuery.GetIcon()))
			}
			if promAlarmQuery.GetColor() != "" {
				promAlarmPageDB = promAlarmPageDB.Where(promAlarmPage.Color.Eq(promAlarmQuery.GetColor()))
			}
			if promAlarmQuery.GetStatus() != prom.Status_Status_NONE {
				promAlarmPageDB = promAlarmPageDB.Where(promAlarmPage.Status.Eq(int32(promAlarmQuery.GetStatus())))
			}
		}
	}

	return promAlarmPageDB.FindByPage(int(offset), int(limit))
}
