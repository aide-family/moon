package data

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"

	buildQuery "prometheus-manager/pkg/build_query"
	"prometheus-manager/pkg/dal/model"
	"prometheus-manager/pkg/dal/query"
	"prometheus-manager/pkg/helper"
	"prometheus-manager/pkg/util/stringer"

	"prometheus-manager/apps/master/internal/biz"
)

type (
	PromV1Repo struct {
		logger *log.Helper
		data   *Data
		db     *query.Query
	}
)

var _ biz.IPromV1Repo = (*PromV1Repo)(nil)

// NewPromV1Repo 初始化data.PromV1Repo, 用于操作底层数据库或缓存
//
//	data:   data配置实例, 内部可以包含db、cache等
//	logger: 日志组件接口, 用于记录日志, 实现了改接口的所有log都可以使用
func NewPromV1Repo(data *Data, logger log.Logger) *PromV1Repo {
	return &PromV1Repo{
		data:   data,
		logger: log.NewHelper(log.With(logger, "module", promModuleName)),
		db:     query.Use(data.DB()),
	}
}

func (p *PromV1Repo) SimpleGroups(ctx context.Context, req *pb.ListSimpleGroupRequest) ([]*model.PromGroup, int64, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.SimpleGroups")
	defer span.End()

	promGroup := p.db.PromGroup
	promGroupDB := promGroup.WithContext(ctx)
	offset, limit := buildQuery.GetPage(req.GetPage())
	if req != nil {
		if req.GetKeyword() != "" {
			promGroupDB = promGroupDB.Where(buildQuery.GetConditionKeywords("%"+req.GetKeyword()+"%", promGroup.Name)...)
		}
	}

	return promGroupDB.Select(promGroup.ID, promGroup.Name).FindByPage(int(offset), int(limit))
}

// GetStrategyByName 根据策略名称获取策略, 如果不存在则返回错误, 一般用于策略名称判重
//
//	ctx: 	 上下文
//	groupID: 策略所属组ID
//	name: 	 查询的策略名称
func (p *PromV1Repo) GetStrategyByName(ctx context.Context, groupID int32, name string) (*model.PromStrategy, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.GetStrategyByName")
	defer span.End()

	promStrategy := p.db.PromStrategy
	promStrategyDB := promStrategy.WithContext(ctx)

	first, err := promStrategyDB.Where(promStrategy.Alert.Eq(name), promStrategy.GroupID.Eq(groupID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perrors.ErrorLogicDataNotFound("name %s not found", name).WithMetadata(map[string]string{
				"name":     name,
				"group_id": string(groupID),
			}).WithCause(err)
		}
		p.logger.WithContext(ctx).Errorw("name", name, "err", err)
		return nil, perrors.ErrorServerDatabaseError("get strategy by name error").WithMetadata(map[string]string{
			"name":     name,
			"group_id": string(groupID),
		}).WithCause(err)
	}

	return first, nil
}

// Strategies 根据查询条件获取策略列表
//
//	ctx: 上下文
//	req: 查询条件
func (p *PromV1Repo) Strategies(ctx context.Context, req *pb.ListStrategyRequest) ([]*model.PromStrategy, int64, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.Strategies")
	defer span.End()

	promStrategy := p.db.PromStrategy
	offset, limit := buildQuery.GetPage(req.GetQuery().GetPage())
	promStrategyDB := promStrategy.WithContext(ctx)

	if req != nil {
		queryPrams := req.GetQuery()
		if queryPrams != nil {
			sorts := queryPrams.GetSort()
			iSorts := make([]buildQuery.ISort, 0, len(sorts))
			for _, sort := range sorts {
				iSorts = append(iSorts, sort)
			}
			promStrategyDB = promStrategyDB.Order(buildQuery.GetSorts(&promStrategy, iSorts...)...)
			promStrategyDB = promStrategyDB.Select(buildQuery.GetSlectExprs(&promStrategy, queryPrams)...)
			keyword := queryPrams.GetKeyword()
			if keyword != "" {
				key := "%" + keyword + "%"
				promStrategyDB = promStrategyDB.Where(buildQuery.GetConditionKeywords(key, promStrategy.Alert)...)
			}
			if queryPrams.GetStartAt() > 0 && queryPrams.GetEndAt() > 0 {
				promStrategyDB = promStrategyDB.Where(promStrategy.CreatedAt.Between(
					time.Unix(queryPrams.GetStartAt(), 0),
					time.Unix(queryPrams.GetEndAt(), 0),
				))
			}
		}

		strategyQuery := req.GetStrategy()
		if strategyQuery != nil {
			if strategyQuery.GetId() != 0 {
				promStrategyDB = promStrategyDB.Where(promStrategy.ID.Eq(strategyQuery.GetId()))
			}
			if strategyQuery.GetAlert() != "" {
				promStrategyDB = promStrategyDB.Where(promStrategy.Alert.Eq(strategyQuery.GetAlert()))
			}
			if strategyQuery.GetGroupId() != 0 {
				promStrategyDB = promStrategyDB.Where(promStrategy.GroupID.Eq(strategyQuery.GetGroupId()))
			}
			if strategyQuery.GetStatus() != prom.Status_Status_NONE {
				promStrategyDB = promStrategyDB.Where(promStrategy.Status.Eq(int32(strategyQuery.GetStatus())))
			}
		}
	}

	return promStrategyDB.Preload(field.Associations).FindByPage(int(offset), int(limit))
}

// Groups 根据查询条件获取策略组列表
//
//	ctx: 上下文
//	req: 查询条件
func (p *PromV1Repo) Groups(ctx context.Context, req *pb.ListGroupRequest) ([]*model.PromGroup, int64, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.Groups")
	defer span.End()

	promGroup := p.db.PromGroup
	offset, limit := buildQuery.GetPage(req.GetQuery().GetPage())
	promGroupDB := promGroup.WithContext(ctx)
	var promDictPreloadExpr []field.Expr
	if req != nil {
		queryPrams := req.GetQuery()
		if queryPrams != nil {
			sorts := queryPrams.GetSort()
			iSorts := make([]buildQuery.ISort, 0, len(sorts))
			for _, sort := range sorts {
				iSorts = append(iSorts, sort)
			}
			promGroupDB = promGroupDB.Order(buildQuery.GetSorts(&promGroup, iSorts...)...)
			promGroupDB = promGroupDB.Select(buildQuery.GetSlectExprs(&promGroup, queryPrams)...)
			keyword := queryPrams.GetKeyword()
			if keyword != "" {
				key := "%" + keyword + "%"
				promGroupDB = promGroupDB.Where(buildQuery.GetConditionKeywords(key, promGroup.Name)...)
			}
			if queryPrams.GetStartAt() > 0 && queryPrams.GetEndAt() > 0 {
				promGroupDB = promGroupDB.Where(promGroup.CreatedAt.Between(
					time.Unix(queryPrams.GetStartAt(), 0),
					time.Unix(queryPrams.GetEndAt(), 0),
				))
			}
		}
		groupQuery := req.GetGroup()
		if groupQuery != nil {
			if groupQuery.GetId() != 0 {
				promGroupDB = promGroupDB.Where(promGroup.ID.Eq(groupQuery.GetId()))
			}
			if groupQuery.GetName() != "" {
				promGroupDB = promGroupDB.Where(promGroup.Name.Eq(groupQuery.GetName()))
			}
			if groupQuery.GetStrategyCount() > 0 {
				promGroupDB = promGroupDB.Where(promGroup.StrategyCount.Gte(groupQuery.GetStrategyCount()))
			}
			categoriesIds := groupQuery.GetCategoriesIds()
			if len(categoriesIds) > 0 {
				promDict := p.db.PromDict
				promDictPreloadExpr = append(promDictPreloadExpr, promDict.ID.In(categoriesIds...))
			}
			if groupQuery.Status != prom.Status_Status_NONE {
				promGroupDB = promGroupDB.Where(promGroup.Status.Eq(int32(groupQuery.Status.Number())))
			}
		}
	}

	return promGroupDB.Preload(promGroup.Categories.On(promDictPreloadExpr...)).FindByPage(int(offset), int(limit))
}

// AllGroups 获取所有策略组
//
//	ctx: 上下文
func (p *PromV1Repo) AllGroups(ctx context.Context, ids []int32) ([]*model.PromGroup, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.AllGroups")
	defer span.End()

	if len(ids) == 0 {
		return nil, nil
	}

	promGroup := p.db.PromGroup
	promStrategy := p.db.PromStrategy
	return promGroup.WithContext(ctx).
		Where(promGroup.Status.Eq(int32(prom.Status_Status_ENABLE)), promGroup.ID.In(ids...)).
		Preload(promGroup.PromStrategies.Where(promStrategy.Status.Eq(int32(prom.Status_Status_ENABLE)))).
		Find()
}

// StrategyDetail 根据ID获取策略详情
//
//	ctx: 上下文
//	id: 策略ID
func (p *PromV1Repo) StrategyDetail(ctx context.Context, id int32) (*model.PromStrategy, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.StrategyDetail")
	defer span.End()

	promStrategy := p.db.PromStrategy
	return promStrategy.WithContext(ctx).Preload(field.Associations).Where(promStrategy.ID.Eq(id)).First()
}

// DeleteStrategyByID 根据ID删除策略
//
//	ctx: 上下文
//	id: 策略ID
func (p *PromV1Repo) DeleteStrategyByID(ctx context.Context, id int32) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.DeleteStrategyByID")
	defer span.End()

	promStrategy := p.db.PromStrategy

	first, err := promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		p.logger.WithContext(ctx).Errorw("PromV1Repo.DeleteStrategyByID", id, "err", err)
		return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	return p.db.Transaction(func(tx *query.Query) error {
		promStrategy = tx.PromStrategy
		promGroup := tx.PromGroup
		inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(first.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Sub(1))
		if err != nil {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.DeleteStrategyByID", id, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
			})
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.DeleteStrategyByID", id, "err", "RowsAffected != 1")
		}

		inf, err = promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.DeleteStrategyByID", id, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
			})
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.DeleteStrategyByID", id, "err", "RowsAffected != 1")
		}

		if err = p.StoreChangeGroupNode(ctx, first.GroupID); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreChangeGroupNode", id, "err", err)
			return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
		}

		return nil
	})
}

// UpdateStrategiesStatusByIds 根据ID批量更新策略状态
//
//	ctx: 上下文
//	ids: 策略ID列表
//	status: 状态
func (p *PromV1Repo) UpdateStrategiesStatusByIds(ctx context.Context, ids []int32, status prom.Status) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.UpdateGroupStatusByID")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promStrategy := tx.PromStrategy
		promStrategyDB := promStrategy.WithContext(ctx)
		promStrategyQueryDB := promStrategy.WithContext(ctx)
		switch len(ids) {
		case 0:
			return nil
		case 1:
			promStrategyDB = promStrategyDB.Where(promStrategy.ID.Eq(ids[0]))
			promStrategyQueryDB = promStrategyQueryDB.Where(promStrategy.ID.Eq(ids[0]))
		default:
			promStrategyDB = promStrategyDB.Where(promStrategy.ID.In(ids...))
			promStrategyQueryDB = promStrategyQueryDB.Where(promStrategy.ID.In(ids...))
		}

		inf, err := promStrategyDB.UpdateColumnSimple(promStrategy.Status.Value(int32(status)))
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateStrategiesStatusByIds", ids, "err", err)
			return perrors.ErrorServerDatabaseError("server database error, %v", err).WithCause(err).WithMetadata(map[string]string{
				"statusCode": strconv.Itoa(int(status)),
				"status":     status.String(),
				"ids":        stringer.New(ids).String(),
			})
		}

		if inf.RowsAffected != int64(len(ids)) {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.UpdateStrategiesStatusByIds", ids, "err", "RowsAffected != 1")
			return perrors.ErrorClientNotFound("PromGroup is not found").WithMetadata(map[string]string{
				"ids": stringer.New(ids).String(),
			})
		}

		var groupIds []any
		if err := promStrategyQueryDB.Select(promStrategy.GroupID).Pluck(promStrategy.GroupID, &groupIds); err != nil {
			p.logger.Errorf("")
			return perrors.ErrorServerDatabaseError("database server unknown err").WithCause(err)
		}

		switch status {
		case prom.Status_Status_DISABLE:
			if err = p.StoreChangeGroupNode(ctx, groupIds...); err != nil {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreChangeGroupNode", groupIds, "err", err)
				return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
			}
		case prom.Status_Status_ENABLE:
			if err = p.StoreDeleteGroupNode(ctx, groupIds...); err != nil {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreDeleteGroupNode", groupIds, "err", err)
				return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
			}
		}

		return nil
	})
}

// UpdateStrategyByID 根据ID更新策略
//
//	ctx: 上下文
//	id: 策略ID
//	m: 策略实体
func (p *PromV1Repo) UpdateStrategyByID(ctx context.Context, id int32, m *model.PromStrategy) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.UpdateStrategyByID")
	defer span.End()

	promStrategy := p.db.PromStrategy
	first, err := p.db.PromStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return perrors.ErrorClientNotFound("PromStrategy is not found").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
			})
		}
		p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateStrategyByID", id, "err", err)
		return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
			"id": strconv.Itoa(int(id)),
		})
	}

	var groupIds []any
	return p.db.Transaction(func(tx *query.Query) error {
		groupIds = append(groupIds, m.GroupID)
		promStrategy = tx.PromStrategy
		if err = promStrategy.AlarmPages.WithContext(ctx).Model(&model.PromStrategy{ID: id}).Replace(m.AlarmPages...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateStrategyByID AlarmPages Replace", id, "m.AlarmPages", m.AlarmPages, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
			})
		}

		if err = promStrategy.Categories.WithContext(ctx).Model(&model.PromStrategy{ID: id}).Replace(m.Categories...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateStrategyByID Categories Replace", id, "m.Categories", m.Categories, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
			})
		}

		if first.GroupID != m.GroupID {
			groupIds = append(groupIds, first.GroupID)
			promGroup := tx.PromGroup
			// 源group strategy_count -1, 目标group strategy_count +1
			simple, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(first.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Sub(1))
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateStrategyByID", id, "err", err)
				return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
					"id": strconv.Itoa(int(id)),
				})
			}
			if simple.RowsAffected != 1 {
				p.logger.WithContext(ctx).Warnw("PromV1Repo.UpdateStrategyByID", first.GroupID, "err", "RowsAffected != 1")
			}

			simple, err = promGroup.WithContext(ctx).Where(promGroup.ID.Eq(m.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Add(1))
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateStrategyByID", id, "err", err)
				return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
					"id": strconv.Itoa(int(id)),
				})
			}
			if simple.RowsAffected != 1 {
				p.logger.WithContext(ctx).Warnw("PromV1Repo.UpdateStrategyByID", m.GroupID, "err", "RowsAffected != 1")
			}
		}

		inf, err := promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).UpdateColumnSimple(
			promStrategy.Alert.Value(m.Alert),
			promStrategy.Expr.Value(m.Expr),
			promStrategy.For.Value(m.For),
			promStrategy.Labels.Value(m.Labels),
			promStrategy.Annotations.Value(m.Annotations),
			promStrategy.AlertLevelID.Value(m.AlertLevelID),
			promStrategy.GroupID.Value(m.GroupID),
			promStrategy.GroupID.Value(m.GroupID),
		)
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateStrategyByID", id, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
			})
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.UpdateStrategyByID", id, "err", "RowsAffected != 1")
		}

		if err = p.StoreChangeGroupNode(ctx, groupIds...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreChangeGroupNode", groupIds, "err", err)
			return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
		}

		return nil
	})
}

// CreateStrategy 创建策略
//
//	ctx: 上下文
//	m: 策略实体
func (p *PromV1Repo) CreateStrategy(ctx context.Context, m *model.PromStrategy) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.CreateStrategy")
	defer span.End()

	alarmPageIds := make([]int32, 0, len(m.AlarmPages))
	for _, alarmPage := range m.AlarmPages {
		alarmPageIds = append(alarmPageIds, alarmPage.ID)
	}
	categoriesIds := make([]int32, 0, len(m.Categories))
	for _, categories := range m.Categories {
		categoriesIds = append(categoriesIds, categories.ID)
	}

	return p.db.Transaction(func(tx *query.Query) error {
		promDict := tx.PromDict
		promGroup := tx.PromGroup
		promAlarmPage := tx.PromAlarmPage
		promStrategy := tx.PromStrategy

		alertLevelInfo, err := promDict.WithContext(ctx).Where(promDict.ID.Eq(m.AlertLevelID)).Select(promDict.ID).First()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.CreateStrategy", m.AlertLevelID, "err", err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return perrors.ErrorLogicDataNotFound("AlertLevel is not found")
			}
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"promStrategy": stringer.New(m).String(),
			})
		}

		m.AlertLevel = alertLevelInfo

		groupStrategyTotal, err := promStrategy.WithContext(ctx).Where(promStrategy.GroupID.Eq(m.GroupID)).Count()
		if err != nil {
			p.logger.WithContext(ctx).Errorf("query group strategy total err: %v", err)
			return perrors.ErrorServerDatabaseError("query group strategy total err").WithCause(err)
		}

		// 这里直接inc会有问题, 如果数据本来就是错误的, 那么永远不能自动修正
		rows, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(m.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Value(groupStrategyTotal + 1))
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.CreateStrategy", m.GroupID, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"promStrategy": stringer.New(m).String(),
			})
		}

		if rows.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.CreateStrategy", m.GroupID, "err", "RowsAffected != 1")
			return perrors.ErrorServerDataNotFound("PromeGroup is not found")
		}

		if len(alarmPageIds) > 0 {
			alarmPageList, err := promAlarmPage.WithContext(ctx).Where(promAlarmPage.ID.In(alarmPageIds...)).Select(promAlarmPage.ID).Find()
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.CreateStrategy", alarmPageIds, "err", err)
				return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
					"promStrategy": stringer.New(m).String(),
				})
			}

			if len(alarmPageList) != len(alarmPageIds) {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.CreateStrategy", alarmPageIds, "err", "alarmPageList != alarmPageIds")
				return perrors.ErrorServerDataNotFound("PromAlarmPage is not found")
			}
			m.AlarmPages = alarmPageList
		}

		if len(categoriesIds) > 0 {
			categoriesList, err := promDict.WithContext(ctx).Where(promDict.ID.In(categoriesIds...)).Select(promDict.ID).Find()
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.CreateStrategy", categoriesIds, "err", err)
				return perrors.ErrorServerDatabaseError("server database error, %v", err)
			}
			if len(categoriesList) != len(categoriesIds) {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.CreateStrategy", categoriesIds, "err", "categoriesList != categoriesIds")
				return perrors.ErrorServerDataNotFound("PromDict is not found")
			}
			m.Categories = categoriesList
		}

		if err = promStrategy.WithContext(ctx).Create(m); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.CreateStrategy", m, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"promStrategy": stringer.New(m).String(),
			})
		}

		if err = p.StoreChangeGroupNode(ctx, m.GroupID); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreChangeGroupNode", m.GroupID, "err", err)
			return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
		}

		return nil
	})
}

// GroupDetail 获取分组详情
//
//	ctx: 上下文
//	id: 规则组ID
func (p *PromV1Repo) GroupDetail(ctx context.Context, id int32) (*model.PromGroup, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.GroupDetail")
	defer span.End()

	promGroup := p.db.PromGroup
	return promGroup.WithContext(ctx).Preload(
		promGroup.Categories,
		promGroup.PromStrategies.AlertLevel,
		promGroup.PromStrategies.AlarmPages,
		promGroup.PromStrategies.Categories,
		promGroup.PromStrategies.Limit(int(buildQuery.DefaultLimit)),
	).Where(promGroup.ID.Eq(id)).First()
}

// DeleteGroupByID 删除分组
//
//	ctx: 上下文
//	id: 规则组ID
func (p *PromV1Repo) DeleteGroupByID(ctx context.Context, id int32) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.DeleteGroupByID")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promGroup := tx.PromGroup
		// 清除关联关系
		promStrategy := tx.PromStrategy
		inf, err := promStrategy.WithContext(ctx).Where(promStrategy.GroupID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.DeleteGroupByID PromStrategy", id, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"groupId": stringer.New(id).String(),
			})
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.DeleteGroupByID PromStrategy", id, "err", "RowsAffected != 1")
		}

		if err = promGroup.Categories.WithContext(ctx).Model(&model.PromGroup{ID: id}).Clear(); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.DeleteGroupByID Categories Clear", id, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"groupId": stringer.New(id).String(),
			})
		}

		// 删除主数据
		inf, err = promGroup.WithContext(ctx).Where(promGroup.ID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.DeleteGroupByID", id, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
			})
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.DeleteGroupByID", id, "err", "RowsAffected != 1")
		}

		if err = p.StoreDeleteGroupNode(ctx, id); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreDeleteGroupNode", id, "err", err)
			return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
		}

		return nil
	})
}

// UpdateGroupsStatusByIds 批量更新分组状态
//
//	ctx: 上下文
//	ids: 分组ID列表
//	status: 状态
func (p *PromV1Repo) UpdateGroupsStatusByIds(ctx context.Context, ids []int32, status prom.Status) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.UpdateGroupStatusByID")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promGroup := tx.PromGroup
		promGroupDB := promGroup.WithContext(ctx)
		promGroupQueryDB := promGroup.WithContext(ctx)
		switch len(ids) {
		case 0:
			return nil
		case 1:
			promGroupDB = promGroupDB.Where(promGroup.ID.Eq(ids[0]))
			promGroupQueryDB = promGroupQueryDB.Where(promGroup.ID.Eq(ids[0]))
		default:
			promGroupDB = promGroupDB.Where(promGroup.ID.In(ids...))
			promGroupQueryDB = promGroupQueryDB.Where(promGroup.ID.In(ids...))
		}

		inf, err := promGroupDB.UpdateColumnSimple(promGroup.Status.Value(int32(status)))
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateGroupsStatusByIds", ids, "err", err)
			return perrors.ErrorServerDatabaseError("server database error, %v", err).WithMetadata(map[string]string{
				"statusCode": strconv.Itoa(int(status)),
				"status":     status.String(),
				"ids":        stringer.New(ids).String(),
			})
		}

		if inf.RowsAffected != int64(len(ids)) {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.UpdateGroupsStatusByIds", ids, "err", "RowsAffected != 1")
			return perrors.ErrorClientNotFound("PromGroup is not found").WithMetadata(map[string]string{
				"ids": stringer.New(ids).String(),
			})
		}

		var groupIds []any

		if err := promGroupQueryDB.Select(promGroup.ID).Pluck(promGroup.ID, &groupIds); err != nil {
			return perrors.ErrorServerDatabaseError("database err").WithCause(err)
		}

		switch status {
		case prom.Status_Status_ENABLE:
			if err = p.StoreChangeGroupNode(ctx, groupIds...); err != nil {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreChangeGroupNode", groupIds, "err", err)
				return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
			}
		case prom.Status_Status_DISABLE:
			if err = p.StoreDeleteGroupNode(ctx, groupIds...); err != nil {
				p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreDeleteGroupNode", groupIds, "err", err)
				return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
			}
		}

		return nil
	})
}

// UpdateGroupByID 根据ID更新分组
//
//	ctx: 上下文
//	id: 分组ID
//	m: 规则组实体
func (p *PromV1Repo) UpdateGroupByID(ctx context.Context, id int32, m *model.PromGroup) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.UpdateGroupByID")
	defer span.End()

	categorieIds := make([]int32, 0, len(m.Categories))
	for _, c := range m.Categories {
		categorieIds = append(categorieIds, c.ID)
	}

	return p.db.Transaction(func(tx *query.Query) error {
		promGroup := tx.PromGroup
		promDict := tx.PromDict
		dictList, err := promDict.WithContext(ctx).Where(
			promDict.ID.In(categorieIds...),
			promDict.Category.Eq(int32(prom.Category_CATEGORY_GROUP)),
		).Select(promDict.ID).Find()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateGroupByID Categories Find", id, "m.Categories", m.Categories, "err", err)
			return err
		}

		if err = promGroup.Categories.WithContext(ctx).Model(&model.PromGroup{ID: id}).Replace(dictList...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateGroupByID Categories Replace", id, "m.Categories", m.Categories, "err", err)
			return err
		}

		inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(id)).UpdateColumnSimple(
			promGroup.Name.Value(m.Name),
			promGroup.Remark.Value(m.Remark),
		)
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.UpdateGroupByID", id, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"id": strconv.Itoa(int(id)),
				"m":  stringer.New(m).String(),
			})
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromV1Repo.UpdateGroupByID", id, "err", "RowsAffected != 1")
		}

		if err = p.StoreChangeGroupNode(ctx, id); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreChangeGroupNode", id, "err", err)
			return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
		}

		return nil
	})
}

// CreateGroup 创建分组
//
//	ctx: 上下文
//	m: 规则组实体
func (p *PromV1Repo) CreateGroup(ctx context.Context, m *model.PromGroup) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.CreateGroup")
	defer span.End()

	categorieIds := make([]int32, 0, len(m.Categories))
	for _, c := range m.Categories {
		categorieIds = append(categorieIds, c.ID)
	}

	return p.db.Transaction(func(tx *query.Query) error {
		promDict := tx.PromDict
		dictList, err := promDict.WithContext(ctx).Where(
			promDict.ID.In(categorieIds...),
			promDict.Category.Eq(int32(prom.Category_CATEGORY_GROUP)),
		).Select(promDict.ID).Find()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.CreateGroup", categorieIds, "err", err)
			return perrors.ErrorServerDatabaseError("database err").WithCause(err).WithMetadata(map[string]string{
				"categorieIds": stringer.New(categorieIds).String(),
			})
		}

		m.Categories = dictList
		promGroup := tx.PromGroup
		if err := promGroup.WithContext(ctx).Create(m); err != nil {
			return perrors.ErrorServerDatabaseError("database err").WithCause(err)
		}

		if err = p.StoreChangeGroupNode(ctx, m.ID); err != nil {
			p.logger.WithContext(ctx).Errorw("PromV1Repo.StoreChangeGroupNode", m.ID, "err", err)
			return perrors.ErrorServerUnknown("server unknown error").WithCause(err)
		}

		return nil
	})
}

func (p *PromV1Repo) StoreChangeGroupNode(ctx context.Context, groupIds ...any) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.StoreChangeGroupNode")
	defer span.End()
	if len(groupIds) == 0 {
		return nil
	}
	// 把变更的group id存入redis的集合类型中
	return p.data.cache.SAdd(ctx, helper.PromGroupChangeKey.String(), groupIds...).Err()
}

func (p *PromV1Repo) StoreDeleteGroupNode(ctx context.Context, groupIds ...any) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.StoreDeleteGroupNode")
	defer span.End()
	if len(groupIds) == 0 {
		return nil
	}
	// 把删除的group id存入redis的集合类型中
	return p.data.cache.SAdd(ctx, helper.PromGroupDeleteKey.String(), groupIds...).Err()
}

// V1 服务版本
//
//	ctx: 上下文
func (p *PromV1Repo) V1(ctx context.Context) string {
	_, span := otel.Tracer(promModuleName).Start(ctx, "PromV1Repo.V1")
	defer span.End()
	return "/prom/v1"
}
