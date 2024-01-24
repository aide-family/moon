package alarmrealtime

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.AlarmRealtimeRepo = (*alarmRealtimeImpl)(nil)

type alarmRealtimeImpl struct {
	repository.UnimplementedAlarmRealtimeRepo
	log  *log.Helper
	data *data.Data
}

type CountRealtime struct {
	StrategyID uint32 `gorm:"column:strategy_id;type:bigint(20);not null;comment:策略id"`
	Count      int64  `gorm:"column:count;type:bigint(20);not null;comment:数量"`
}

func (l *alarmRealtimeImpl) CountRealtimeAlarmByStrategyIds(ctx context.Context, strategyIds ...uint32) (map[uint32]int64, error) {
	var countRealtimeList []CountRealtime
	if err := l.data.DB().WithContext(ctx).Model(&do.PromAlarmRealtime{}).
		Where("strategy_id in (?)", strategyIds).
		Group("strategy_id").
		Select("strategy_id, count(*) as count").Scan(&countRealtimeList).Error; err != nil {
		return nil, err
	}
	resMap := make(map[uint32]int64, len(countRealtimeList))
	for _, item := range countRealtimeList {
		resMap[item.StrategyID] = item.Count
	}
	return resMap, nil
}

func (l *alarmRealtimeImpl) Create(ctx context.Context, req ...*bo.AlarmRealtimeBO) ([]*bo.AlarmRealtimeBO, error) {
	historyIds := make([]uint32, 0, len(req))
	newRealtimeModels := slices.To(req, func(item *bo.AlarmRealtimeBO) *do.PromAlarmRealtime {
		historyIds = append(historyIds, item.HistoryID)
		return item.ToModel()
	})
	if err := l.data.DB().WithContext(ctx).Clauses(basescopes.RealtimeAlarmClauseOnConflict()).CreateInBatches(newRealtimeModels, 50).Error; err != nil {
		return nil, err
	}

	var realtimeAlarmList []*do.PromAlarmRealtime
	// 查询插入的新数据
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InHistoryIds(historyIds...)).Find(&realtimeAlarmList).Error; err != nil {
		return nil, err
	}

	return slices.To(realtimeAlarmList, func(m *do.PromAlarmRealtime) *bo.AlarmRealtimeBO { return bo.AlarmRealtimeModelToBO(m) }), nil
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
		return l.data.Client().HMSet(ctx, key, saveCacheArgs...).Err()
	})
	eg.Go(func() error {
		// 删除已经恢复的告警数据
		if len(removeCacheArgs) == 0 {
			return nil
		}
		return l.data.Client().HDel(ctx, key, removeCacheArgs...).Err()
	})

	return eg.Wait()
}

func (l *alarmRealtimeImpl) DeleteCacheByHistoryId(ctx context.Context, historyId ...uint32) error {
	key := consts.AlarmRealtimeCacheById.String()
	fields := slices.To(historyId, func(id uint32) string { return strconv.Itoa(int(id)) })
	return l.data.Client().HDel(ctx, key, fields...).Err()
}

func (l *alarmRealtimeImpl) AppendAlarmBeenNotifyMembers(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmBeenNotifyMemberBO) error {
	var first do.PromAlarmRealtime

	if err := l.data.DB().WithContext(ctx).First(&first, realtimeAlarmID).Error; err != nil {
		return err
	}

	return l.data.DB().WithContext(ctx).Model(&first).Association(basescopes.RealtimeAssociationBeenNotifyMembers).Append(req.ToModel())
}

func (l *alarmRealtimeImpl) AppendAlarmBeenNotifyChatGroups(ctx context.Context, realtimeAlarmID uint32, req *bo.PromAlarmBeenNotifyChatGroupBO) error {
	var first do.PromAlarmHistory
	if err := l.data.DB().WithContext(ctx).First(&first, realtimeAlarmID).Error; err != nil {
		return err
	}

	return l.data.DB().WithContext(ctx).Model(&first).Association(basescopes.RealtimeAssociationBeenChatGroups).Append(req.ToModel())
}

func (l *alarmRealtimeImpl) GetRealtimeDetailById(ctx context.Context, id uint32, scopes ...basescopes.ScopeMethod) (*bo.AlarmRealtimeBO, error) {
	var first do.PromAlarmRealtime
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&first, id).Error; err != nil {
		return nil, err
	}

	return bo.AlarmRealtimeModelToBO(&first), nil
}

func (l *alarmRealtimeImpl) GetRealtimeList(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.AlarmRealtimeBO, error) {
	var list []*do.PromAlarmRealtime
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, basescopes.Page(pgInfo))...).Find(&list).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Model(&do.PromAlarmRealtime{}).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}

	return slices.To(list, func(info *do.PromAlarmRealtime) *bo.AlarmRealtimeBO {
		return bo.AlarmRealtimeModelToBO(info)
	}), nil
}

func (l *alarmRealtimeImpl) AlarmIntervene(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmInterveneBO) error {
	var first bo.AlarmRealtimeBO
	if err := l.data.DB().WithContext(ctx).First(&first, realtimeAlarmID).Error; err != nil {
		return err
	}
	newAlarmIntervene := req.ToModel()
	return l.data.DB().WithContext(ctx).Model(&first).Association(basescopes.RealtimeAssociationReplaceIntervenes).Append(newAlarmIntervene)
}

func (l *alarmRealtimeImpl) AlarmUpgrade(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmUpgradeBO) error {
	var first bo.AlarmRealtimeBO
	if err := l.data.DB().WithContext(ctx).First(&first, realtimeAlarmID).Error; err != nil {
		return err
	}
	newAlarmUpgrade := req.ToModel()
	return l.data.DB().WithContext(ctx).Model(&first).Association(basescopes.RealtimeAssociationUpgradeInfo).Append(newAlarmUpgrade)
}

func (l *alarmRealtimeImpl) AlarmSuppress(ctx context.Context, realtimeAlarmID uint32, req *bo.AlarmSuppressBO) error {
	var first bo.AlarmRealtimeBO
	if err := l.data.DB().WithContext(ctx).First(&first, realtimeAlarmID).Error; err != nil {
		return err
	}
	newAlarmSuppress := req.ToModel()
	return l.data.DB().WithContext(ctx).Model(&first).Association(basescopes.RealtimeAssociationSuppressInfo).Append(newAlarmSuppress)
}

func (l *alarmRealtimeImpl) GetRealtimeCount(ctx context.Context, scopes ...basescopes.ScopeMethod) (int64, error) {
	var total int64
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func NewAlarmRealtime(d *data.Data, logger log.Logger) repository.AlarmRealtimeRepo {
	return &alarmRealtimeImpl{
		log:  log.NewHelper(log.With(logger, "module", "repository.alarm.realtime")),
		data: d,
	}
}
