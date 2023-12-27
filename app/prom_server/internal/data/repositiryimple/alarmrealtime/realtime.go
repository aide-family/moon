package alarmrealtime

import (
	"context"
	"strconv"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/pkg/helper/model/basescopes"

	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/alarmscopes"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.AlarmRealtimeRepo = (*alarmRealtimeImpl)(nil)

type alarmRealtimeImpl struct {
	repository.UnimplementedAlarmRealtimeRepo
	log *log.Helper
	d   *data.Data

	query.IAction[model.PromAlarmRealtime]
}

func (l *alarmRealtimeImpl) Create(ctx context.Context, req ...*bo.AlarmRealtimeBO) ([]*bo.AlarmRealtimeBO, error) {
	historyIds := make([]uint32, 0, len(req))
	newRealtimeModels := slices.To(req, func(item *bo.AlarmRealtimeBO) *model.PromAlarmRealtime {
		historyIds = append(historyIds, item.HistoryID)
		return item.ToModel()
	})
	if err := l.DB().WithContext(ctx).Scopes(alarmscopes.ClauseOnConflict()).CreateInBatches(newRealtimeModels, 50).Error; err != nil {
		return nil, err
	}

	var realtimeAlarmList []*model.PromAlarmRealtime
	// 查询插入的新数据
	if err := l.DB().WithContext(ctx).Scopes(alarmscopes.InHistoryIds(historyIds...)).Find(&realtimeAlarmList).Error; err != nil {
		return nil, err
	}

	return slices.To(realtimeAlarmList, func(m *model.PromAlarmRealtime) *bo.AlarmRealtimeBO { return bo.AlarmRealtimeModelToBO(m) }), nil
}

func (l *alarmRealtimeImpl) CacheByHistoryId(ctx context.Context, req ...*bo.AlarmRealtimeBO) error {
	// 写入redis hash表中
	saveCacheArgs := make([]any, 0, len(req))
	removeCacheArgs := make([]string, 0, len(req))
	for _, alarmRealtimeBO := range req {
		realtimeIdKey := strconv.Itoa(int(alarmRealtimeBO.ID))
		if alarmRealtimeBO.Status.IsResolved() {
			removeCacheArgs = append(removeCacheArgs, realtimeIdKey)
			continue
		}
		saveCacheArgs = append(saveCacheArgs, realtimeIdKey, alarmRealtimeBO)
	}

	key := consts.AlarmRealtimeCacheById.String()
	eg := errgroup.Group{}
	eg.Go(func() error {
		// 插入新数据
		if len(saveCacheArgs) == 0 {
			return nil
		}
		return l.d.Client().HMSet(ctx, key, saveCacheArgs...).Err()
	})
	eg.Go(func() error {
		// 删除已经恢复的告警数据
		if len(removeCacheArgs) == 0 {
			return nil
		}
		return l.d.Client().HDel(ctx, key, removeCacheArgs...).Err()
	})

	return eg.Wait()
}

func (l *alarmRealtimeImpl) DeleteCacheByHistoryId(ctx context.Context, historyId ...uint32) error {
	key := consts.AlarmRealtimeCacheById.String()
	fields := slices.To(historyId, func(id uint32) string { return strconv.Itoa(int(id)) })
	return l.d.Client().HDel(ctx, key, fields...).Err()
}

func (l *alarmRealtimeImpl) AppendAlarmBeenNotifyMembers(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmBeenNotifyMemberBO) error {
	first, err := l.WithContext(ctx).FirstByID(realtimeAlarmID)
	if err != nil {
		return err
	}

	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationBeenNotifyMembers).Append(req.ToModel())
}

func (l *alarmRealtimeImpl) AppendAlarmBeenNotifyChatGroups(ctx context.Context, realtimeAlarmID uint32, req *bo.PromAlarmBeenNotifyChatGroupBO) error {
	first, err := l.WithContext(ctx).FirstByID(realtimeAlarmID)
	if err != nil {
		return err
	}

	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationBeenChatGroups).Append(req.ToModel())
}

func (l *alarmRealtimeImpl) GetRealtimeDetailById(ctx context.Context, id uint32, scopes ...query.ScopeMethod) (*bo.AlarmRealtimeBO, error) {
	first, err := l.WithContext(ctx).First(append(scopes, basescopes.InIds(id))...)
	if err != nil {
		return nil, err
	}

	return bo.AlarmRealtimeModelToBO(first), nil
}

func (l *alarmRealtimeImpl) GetRealtimeList(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.AlarmRealtimeBO, error) {
	list, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	return slices.To(list, func(info *model.PromAlarmRealtime) *bo.AlarmRealtimeBO {
		return bo.AlarmRealtimeModelToBO(info)
	}), nil
}

func (l *alarmRealtimeImpl) AlarmIntervene(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmInterveneBO) error {
	first, err := l.WithContext(ctx).FirstByID(realtimeAlarmID)
	if err != nil {
		return err
	}
	newAlarmIntervene := req.ToModel()
	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationReplaceIntervenes).Append(newAlarmIntervene)
}

func (l *alarmRealtimeImpl) AlarmUpgrade(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmUpgradeBO) error {
	first, err := l.WithContext(ctx).FirstByID(realtimeAlarmID)
	if err != nil {
		return err
	}
	newAlarmUpgrade := req.ToModel()
	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationUpgradeInfo).Append(newAlarmUpgrade)
}

func (l *alarmRealtimeImpl) AlarmSuppress(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmSuppressBO) error {
	first, err := l.WithContext(ctx).FirstByID(realtimeAlarmID)
	if err != nil {
		return err
	}
	newAlarmSuppress := req.ToModel()
	return l.DB().WithContext(ctx).Model(first).Association(alarmscopes.RealtimeAssociationSuppressInfo).Append(newAlarmSuppress)
}

func (l *alarmRealtimeImpl) GetRealtimeCount(ctx context.Context, scopes ...query.ScopeMethod) (int64, error) {
	return l.WithContext(ctx).Count(scopes...)
}

func NewAlarmRealtime(d *data.Data, logger log.Logger) repository.AlarmRealtimeRepo {
	return &alarmRealtimeImpl{
		log: log.NewHelper(log.With(logger, "module", "repository.alarm.realtime")),
		d:   d,
		IAction: query.NewAction[model.PromAlarmRealtime](
			query.WithDB[model.PromAlarmRealtime](d.DB()),
		),
	}
}
