package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api/perrors"
	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/pkg/alert"
	"prometheus-manager/pkg/dal/query"
)

type (
	AlertStotageRepo struct {
		logger *log.Helper
		data   *Data
	}

	AlertCacheRepo struct {
		logger *log.Helper
		data   *Data
	}

	AlertNotifyRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IWatchRepo = (*AlertStotageRepo)(nil)
var _ biz.IWatchRepo = (*AlertCacheRepo)(nil)
var _ biz.IWatchRepo = (*AlertNotifyRepo)(nil)

func NewAlertStorageRepo(data *Data, logger log.Logger) *AlertStotageRepo {
	return &AlertStotageRepo{data: data, logger: log.NewHelper(log.With(logger, "module", watchModuleName))}
}

func NewAlertCacheRepo(data *Data, logger log.Logger) *AlertCacheRepo {
	return &AlertCacheRepo{data: data, logger: log.NewHelper(log.With(logger, "module", watchModuleName))}
}

func NewAlertNotifyRepo(data *Data, logger log.Logger) *AlertNotifyRepo {
	return &AlertNotifyRepo{data: data, logger: log.NewHelper(log.With(logger, "module", watchModuleName))}
}

func NewWatchRepoes(data *Data, logger log.Logger) []biz.IWatchRepo {
	return []biz.IWatchRepo{
		NewAlertStorageRepo(data, logger),
		NewAlertCacheRepo(data, logger),
		NewAlertNotifyRepo(data, logger),
	}
}

func (l *AlertStotageRepo) V1(_ context.Context) string {
	return "AlertStotageRepo.V1"
}

func (l *AlertStotageRepo) SyncAlert(ctx context.Context, req *alert.Data) error {
	ctx, span := otel.Tracer(watchModuleName).Start(ctx, "AlertStotageRepo.WatchAlert")
	defer span.End()

	historyModel := alertDataToModel(req)

	return query.Use(l.data.DB()).Transaction(func(tx *query.Query) error {
		if historyModel != nil {
			history := tx.PromAlarmHistory
			if err := history.WithContext(ctx).Create(historyModel); err != nil {
				l.logger.WithContext(ctx).Errorf("alarm history storage err: %v", err)
				return perrors.ErrorServerDatabaseError("alarm history storage err").WithCause(err)
			}
		}
		return nil
	})
}

/* ------------------------------AlertCacheRepo------------------------------ */

func (l *AlertCacheRepo) V1(_ context.Context) string {
	return "AlertCacheRepo.V1"
}

func (l *AlertCacheRepo) SyncAlert(ctx context.Context, req *alert.Data) error {
	ctx, span := otel.Tracer(watchModuleName).Start(ctx, "AlertCacheRepo.WatchAlert")
	defer span.End()
	marshal, err := json.Marshal(req)
	if err != nil {
		return err
	}
	// TODO 落库
	fmt.Println(string(marshal))
	return nil
}

/* ------------------------------AlertNotifyRepo------------------------------ */

func (l *AlertNotifyRepo) V1(_ context.Context) string {
	return "AlertNotifyRepo.V1"
}

func (l *AlertNotifyRepo) SyncAlert(ctx context.Context, req *alert.Data) error {
	ctx, span := otel.Tracer(watchModuleName).Start(ctx, "AlertNotifyRepo.WatchAlert")
	defer span.End()
	marshal, err := json.Marshal(req)
	if err != nil {
		return err
	}
	// TODO 落库
	fmt.Println(string(marshal))
	return nil
}
