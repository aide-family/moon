package data

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"
	"prometheus-manager/dal/model"
	"prometheus-manager/dal/query"
	buildQuery "prometheus-manager/pkg/build_query"

	"prometheus-manager/apps/master/internal/biz"
)

type (
	PromRepo struct {
		logger *log.Helper
		data   *Data
		db     *query.Query
	}
)

func (p *PromRepo) GetStrategyByName(ctx context.Context, groupID int32, name string) (*model.PromStrategy, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.GetStrategyByName")
	defer span.End()

	promStrategy := p.db.PromStrategy
	promStrategyDB := promStrategy.WithContext(ctx)

	first, err := promStrategyDB.Where(promStrategy.Alert.Eq(name), promStrategy.GroupID.Eq(groupID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, perrors.ErrorLogicDataNotFound("name %s not found", name)
		}
		p.logger.WithContext(ctx).Errorw("name", name, "err", err)
		return nil, perrors.ErrorServerDatabaseError("get strategy by name error").WithMetadata(map[string]string{
			"name":     name,
			"group_id": string(groupID),
			"err":      err.Error(),
		})
	}

	return first, nil
}

func (p *PromRepo) Strategies(ctx context.Context, req *pb.ListStrategyRequest) ([]*model.PromStrategy, int64, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.Strategies")
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
				promStrategyDB = promStrategyDB.Where(buildQuery.GetKeywords(key, promStrategy.Alert)...)
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
		}
	}

	return promStrategyDB.Preload(field.Associations).FindByPage(int(offset), int(limit))
}

func (p *PromRepo) Groups(ctx context.Context, req *pb.ListGroupRequest) ([]*model.PromGroup, int64, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.Groups")
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
				promGroupDB = promGroupDB.Where(buildQuery.GetKeywords(key, promGroup.Name)...)
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

func (p *PromRepo) StrategyDetail(ctx context.Context, id int32) (*model.PromStrategy, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.StrategyDetail")
	defer span.End()

	promStrategy := p.db.PromStrategy
	return promStrategy.WithContext(ctx).Preload(field.Associations).Where(promStrategy.ID.Eq(id)).First()
}

func (p *PromRepo) DeleteStrategyByID(ctx context.Context, id int32) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.DeleteStrategyByID")
	defer span.End()

	promStrategy := p.db.PromStrategy

	first, err := promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).First()
	if err != nil {
		p.logger.WithContext(ctx).Errorw("PromRepo.DeleteStrategyByID", id, "err", err)
		return err
	}

	return p.db.Transaction(func(tx *query.Query) error {
		promStrategy = tx.PromStrategy
		promGroup := tx.PromGroup
		inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(first.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Sub(1))
		if err != nil {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteStrategyByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteStrategyByID", id, "err", "RowsAffected != 1")
		}

		inf, err = promStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.DeleteStrategyByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteStrategyByID", id, "err", "RowsAffected != 1")
		}

		return nil
	})
}

func (p *PromRepo) UpdateStrategyByID(ctx context.Context, id int32, m *model.PromStrategy) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.UpdateStrategyByID")
	defer span.End()

	promStrategy := p.db.PromStrategy
	first, err := p.db.PromStrategy.WithContext(ctx).Where(promStrategy.ID.Eq(id)).First()
	if err != nil {
		p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
		return err
	}

	return p.db.Transaction(func(tx *query.Query) error {
		promStrategy = tx.PromStrategy
		if err = promStrategy.AlarmPages.WithContext(ctx).Model(&model.PromStrategy{ID: id}).Replace(m.AlarmPages...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID AlarmPages Replace", id, "m.AlarmPages", m.AlarmPages, "err", err)
			return err
		}

		if err = promStrategy.Categories.WithContext(ctx).Model(&model.PromStrategy{ID: id}).Replace(m.Categories...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID Categories Replace", id, "m.Categories", m.Categories, "err", err)
			return err
		}

		if first.GroupID != m.GroupID {
			promGroup := tx.PromGroup
			// 源group strategy_count -1, 目标group strategy_count +1
			simple, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(first.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Sub(1))
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
				return err
			}
			if simple.RowsAffected != 1 {
				p.logger.WithContext(ctx).Warnw("PromRepo.UpdateStrategyByID", first.GroupID, "err", "RowsAffected != 1")
			}

			simple, err = promGroup.WithContext(ctx).Where(promGroup.ID.Eq(m.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Add(1))
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
				return err
			}
			if simple.RowsAffected != 1 {
				p.logger.WithContext(ctx).Warnw("PromRepo.UpdateStrategyByID", m.GroupID, "err", "RowsAffected != 1")
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
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateStrategyByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.UpdateStrategyByID", id, "err", "RowsAffected != 1")
		}

		return nil
	})
}

func (p *PromRepo) CreateStrategy(ctx context.Context, m *model.PromStrategy) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.CreateStrategy")
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
			p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", m.AlertLevelID, "err", err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return perrors.ErrorLogicDataNotFound("AlertLevel is not found")
			}
			return perrors.ErrorServerDatabaseError("server database error, %v", err)
		}

		m.AlertLevel = alertLevelInfo

		rows, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(m.GroupID)).UpdateColumnSimple(promGroup.StrategyCount.Add(1))
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", m.GroupID, "err", err)
			return perrors.ErrorServerDatabaseError("server database error, %v", err)
		}

		if rows.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.CreateStrategy", m.GroupID, "err", "RowsAffected != 1")
			return perrors.ErrorServerDataNotFound("PromeGroup is not found")
		}

		if len(alarmPageIds) > 0 {
			alarmPageList, err := promAlarmPage.WithContext(ctx).Where(promAlarmPage.ID.In(alarmPageIds...)).Select(promAlarmPage.ID).Find()
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", alarmPageIds, "err", err)
				return perrors.ErrorServerDatabaseError("server database error, %v", err)
			}

			if len(alarmPageList) != len(alarmPageIds) {
				p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", alarmPageIds, "err", "alarmPageList != alarmPageIds")
				return perrors.ErrorServerDataNotFound("PromAlarmPage is not found")
			}
			m.AlarmPages = alarmPageList
		}

		if len(categoriesIds) > 0 {
			categoriesList, err := promDict.WithContext(ctx).Where(promDict.ID.In(categoriesIds...)).Select(promDict.ID).Find()
			if err != nil {
				p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", categoriesIds, "err", err)
				return perrors.ErrorServerDatabaseError("server database error, %v", err)
			}
			if len(categoriesList) != len(categoriesIds) {
				p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", categoriesIds, "err", "categoriesList != categoriesIds")
				return perrors.ErrorServerDataNotFound("PromDict is not found")
			}
			m.Categories = categoriesList
		}

		if err = promStrategy.WithContext(ctx).Create(m); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.CreateStrategy", m, "err", err)
			return perrors.ErrorServerDatabaseError("server database error, %v", err)
		}

		return nil
	})
}

func (p *PromRepo) GroupDetail(ctx context.Context, id int32) (*model.PromGroup, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.GroupDetail")
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

func (p *PromRepo) DeleteGroupByID(ctx context.Context, id int32) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.DeleteGroupByID")
	defer span.End()

	return p.db.Transaction(func(tx *query.Query) error {
		promGroup := tx.PromGroup
		// 清除关联关系
		promStrategy := tx.PromStrategy
		inf, err := promStrategy.WithContext(ctx).Where(promStrategy.GroupID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.DeleteGroupByID PromStrategy", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteGroupByID PromStrategy", id, "err", "RowsAffected != 1")
		}

		if err = promGroup.Categories.WithContext(ctx).Model(&model.PromGroup{ID: id}).Clear(); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.DeleteGroupByID Categories Clear", id, "err", err)
			return err
		}

		// 删除主数据
		inf, err = promGroup.WithContext(ctx).Where(promGroup.ID.Eq(id)).Delete()
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.DeleteGroupByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.DeleteGroupByID", id, "err", "RowsAffected != 1")
		}

		return nil
	})
}

func (p *PromRepo) UpdateGroupByID(ctx context.Context, id int32, m *model.PromGroup) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.UpdateGroupByID")
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
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateGroupByID Categories Find", id, "m.Categories", m.Categories, "err", err)
			return err
		}

		if err = promGroup.Categories.WithContext(ctx).Model(&model.PromGroup{ID: id}).Replace(dictList...); err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateGroupByID Categories Replace", id, "m.Categories", m.Categories, "err", err)
			return err
		}

		inf, err := promGroup.WithContext(ctx).Where(promGroup.ID.Eq(id)).UpdateColumnSimple(
			promGroup.Name.Value(m.Name),
			promGroup.Remark.Value(m.Remark),
		)
		if err != nil {
			p.logger.WithContext(ctx).Errorw("PromRepo.UpdateGroupByID", id, "err", err)
			return err
		}

		if inf.RowsAffected != 1 {
			p.logger.WithContext(ctx).Warnw("PromRepo.UpdateGroupByID", id, "err", "RowsAffected != 1")
		}

		return nil
	})
}

func (p *PromRepo) CreateGroup(ctx context.Context, m *model.PromGroup) error {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromRepo.CreateGroup")
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
			p.logger.WithContext(ctx).Errorw("PromRepo.CreateGroup", categorieIds, "err", err)
			return err
		}

		m.Categories = dictList
		promGroup := tx.PromGroup
		return promGroup.WithContext(ctx).Create(m)
	})
}

func (p *PromRepo) V1(_ context.Context) string {
	return "/prom/v1"
}

var _ biz.IPromRepo = (*PromRepo)(nil)

func NewPromRepo(data *Data, logger log.Logger) *PromRepo {
	return &PromRepo{
		data:   data,
		logger: log.NewHelper(log.With(logger, "module", "data/Prom")),
		db:     query.Use(data.DB()),
	}
}
