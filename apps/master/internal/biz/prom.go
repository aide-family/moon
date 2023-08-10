package biz

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/api"
	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"

	"prometheus-manager/dal/model"
	"prometheus-manager/pkg/times"

	"prometheus-manager/apps/master/internal/service"
)

type (
	IPromRepo interface {
		V1Repo
		CreateGroup(ctx context.Context, m *model.PromGroup) error
		UpdateGroupByID(ctx context.Context, id int32, m *model.PromGroup) error
		DeleteGroupByID(ctx context.Context, id int32) error
		GroupDetail(ctx context.Context, id int32) (*model.PromGroup, error)
		Groups(ctx context.Context, req *pb.ListGroupRequest) ([]*model.PromGroup, int64, error)

		CreateStrategy(ctx context.Context, m *model.PromStrategy) error
		UpdateStrategyByID(ctx context.Context, id int32, m *model.PromStrategy) error
		DeleteStrategyByID(ctx context.Context, id int32) error
		StrategyDetail(ctx context.Context, id int32) (*model.PromStrategy, error)
		Strategies(ctx context.Context, req *pb.ListStrategyRequest) ([]*model.PromStrategy, int64, error)
		GetStrategyByName(ctx context.Context, groupID int32, name string) (*model.PromStrategy, error)
	}

	PromLogic struct {
		logger *log.Helper
		repo   IPromRepo
	}
)

var _ service.IPromLogic = (*PromLogic)(nil)

func NewPromLogic(repo IPromRepo, logger log.Logger) *PromLogic {
	return &PromLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Prom"))}
}

func buildInsertCategories(categorieIds []int32) []*model.PromDict {
	result := make([]*model.PromDict, 0, len(categorieIds))
	for _, categoryId := range categorieIds {
		result = append(result, &model.PromDict{
			ID: categoryId,
		})
	}
	return result
}

func buildModelPromGroup(groupItem *prom.GroupItem) *model.PromGroup {
	return &model.PromGroup{
		Name:       groupItem.GetName(),
		Remark:     groupItem.GetRemark(),
		Categories: buildInsertCategories(groupItem.GetCategoriesIds()),
	}
}

func buildLabels(labelsStr string) map[string]string {
	result := make(map[string]string)
	if labelsStr != "" {
		_ = json.Unmarshal([]byte(labelsStr), &result)
	}
	return result
}

func buildAnnotations(annotationsStr string) map[string]string {
	result := make(map[string]string)
	if annotationsStr != "" {
		_ = json.Unmarshal([]byte(annotationsStr), &result)
	}
	return result
}

func buildPromStrategies(strategyItems []*model.PromStrategy) []*prom.StrategyItem {
	result := make([]*prom.StrategyItem, 0, len(strategyItems))
	for _, strategyItem := range strategyItems {
		result = append(result, buildStrategyItem(strategyItem))
	}
	return result
}

func buildDictItem(category *model.PromDict) *prom.DictItem {
	if category == nil {
		return nil
	}
	return &prom.DictItem{
		Id:        category.ID,
		Name:      category.Name,
		Remark:    category.Remark,
		Category:  prom.Category(category.Category),
		Color:     category.Color,
		CreatedAt: times.TimeToUnix(category.CreatedAt),
		UpdatedAt: times.TimeToUnix(category.UpdatedAt),
	}
}

func buidlPromCategories(categories []*model.PromDict) []*prom.DictItem {
	result := make([]*prom.DictItem, 0, len(categories))
	for _, category := range categories {
		result = append(result, buildDictItem(category))
	}
	return result
}

func buildAlarmPageItem(alarmPage *model.PromAlarmPage) *prom.AlarmPageItem {
	return &prom.AlarmPageItem{
		Id:        alarmPage.ID,
		Name:      alarmPage.Name,
		Remark:    alarmPage.Remark,
		Icon:      alarmPage.Icon,
		Color:     alarmPage.Color,
		CreatedAt: times.TimeToUnix(alarmPage.CreatedAt),
		UpdatedAt: times.TimeToUnix(alarmPage.UpdatedAt),
	}
}

func buildAlarmPages(alarmPages []*model.PromAlarmPage) []*prom.AlarmPageItem {
	result := make([]*prom.AlarmPageItem, 0, len(alarmPages))
	for _, alarmPage := range alarmPages {
		result = append(result, buildAlarmPageItem(alarmPage))
	}
	return result
}

func buildStrategyItem(strategyItem *model.PromStrategy) *prom.StrategyItem {
	return &prom.StrategyItem{
		GroupId:     strategyItem.GroupID,
		Alert:       strategyItem.Alert,
		Expr:        strategyItem.Expr,
		For:         strategyItem.For,
		Labels:      buildLabels(strategyItem.Labels),
		Annotations: buildAnnotations(strategyItem.Annotations),
		CreatedAt:   times.TimeToUnix(strategyItem.CreatedAt),
		UpdatedAt:   times.TimeToUnix(strategyItem.UpdatedAt),
		Categories:  buidlPromCategories(strategyItem.Categories),
		CategorieIds: func() []int32 {
			result := make([]int32, 0, len(strategyItem.Categories))
			for _, category := range strategyItem.Categories {
				result = append(result, category.ID)
			}
			return result
		}(),
		AlertLevelId: strategyItem.AlertLevelID,
		AlertLevel:   buildDictItem(strategyItem.AlertLevel),
		AlarmPages:   buildAlarmPages(strategyItem.AlarmPages),
		AlarmPageIds: func() []int32 {
			result := make([]int32, 0, len(strategyItem.AlarmPages))
			for _, alarmPage := range strategyItem.AlarmPages {
				result = append(result, alarmPage.ID)
			}
			return result
		}(),
		Status: prom.Status(strategyItem.Status),
		Id:     strategyItem.ID,
	}
}

func buildGroupItem(group *model.PromGroup) *prom.GroupItem {
	return &prom.GroupItem{
		Id:             group.ID,
		Name:           group.Name,
		Remark:         group.Remark,
		CreatedAt:      times.TimeToUnix(group.CreatedAt),
		UpdatedAt:      times.TimeToUnix(group.UpdatedAt),
		PromStrategies: buildPromStrategies(group.PromStrategies),
		Categories:     buidlPromCategories(group.Categories),
		StrategyCount:  group.StrategyCount,
		Status:         prom.Status(group.Status),
		CategoriesIds: func() []int32 {
			result := make([]int32, 0, len(group.Categories))
			for _, category := range group.Categories {
				result = append(result, category.ID)
			}
			return result
		}(),
	}
}

func annotationsToString(annotations map[string]string) string {
	result, _ := json.Marshal(annotations)
	return string(result)
}

func labelsToString(labels map[string]string) string {
	result, _ := json.Marshal(labels)
	return string(result)
}

func buildModelAlarmPage(alarmPageId int32) *model.PromAlarmPage {
	return &model.PromAlarmPage{ID: alarmPageId}
}

func buildModelAlarmPages(alarmPages []int32) []*model.PromAlarmPage {
	result := make([]*model.PromAlarmPage, 0, len(alarmPages))
	for _, alarmPageId := range alarmPages {
		result = append(result, buildModelAlarmPage(alarmPageId))
	}
	return result
}

func buildModelPromStrategy(strategyItem *prom.StrategyItem) *model.PromStrategy {
	return &model.PromStrategy{
		AlarmPages:   buildModelAlarmPages(strategyItem.GetAlarmPageIds()),
		Categories:   buildInsertCategories(strategyItem.GetCategorieIds()),
		GroupID:      strategyItem.GetGroupId(),
		Alert:        strategyItem.GetAlert(),
		Expr:         strategyItem.GetExpr(),
		For:          strategyItem.GetFor(),
		Labels:       labelsToString(strategyItem.GetLabels()),
		Annotations:  annotationsToString(strategyItem.GetAnnotations()),
		AlertLevelID: strategyItem.GetAlertLevelId(),
		Status:       int32(strategyItem.GetStatus()),
	}
}

func (s *PromLogic) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.CreateGroup")
	defer span.End()

	insertModel := buildModelPromGroup(req.GetGroup())
	if err := s.repo.CreateGroup(ctx, insertModel); err != nil {
		s.logger.WithContext(ctx).Errorw("创建Prometheus分组失败", insertModel, "err", err)
		return nil, perrors.ErrorLogicCreatePrometheusGroupFailed("创建Prometheus分组失败")
	}

	return &pb.CreateGroupReply{Response: &api.Response{Message: "创建Prometheus group成功"}}, nil
}

func (s *PromLogic) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.UpdateGroup")
	defer span.End()

	edieModel := buildModelPromGroup(req.GetGroup())
	if err := s.repo.UpdateGroupByID(ctx, req.GetId(), edieModel); err != nil {
		s.logger.WithContext(ctx).Errorw("更新Prometheus分组失败", edieModel, "err", err)
		return nil, perrors.ErrorLogicEditPrometheusGroupFailed("更新Prometheus分组失败")
	}

	return &pb.UpdateGroupReply{Response: &api.Response{Message: "更新Prometheus group成功"}}, nil
}

func (s *PromLogic) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.DeleteGroup")
	defer span.End()

	if err := s.repo.DeleteGroupByID(ctx, req.GetId()); err != nil {
		s.logger.WithContext(ctx).Errorw("删除Prometheus分组失败", "id", req.GetId(), "err", err)
		return nil, perrors.ErrorLogicDeletePrometheusGroupFailed("删除Prometheus分组失败")
	}

	return &pb.DeleteGroupReply{Response: &api.Response{Message: "删除Prometheus group成功"}}, nil
}

func (s *PromLogic) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.GetGroup")
	defer span.End()

	groupDetail, err := s.repo.GroupDetail(ctx, req.GetId())
	if err != nil {
		s.logger.WithContext(ctx).Errorw("获取Prometheus分组失败", "id", req.GetId(), "err", err)
		return nil, perrors.ErrorServerDatabaseError("获取Prometheus分组失败")
	}

	return &pb.GetGroupReply{
		Response: &api.Response{Message: "获取Prometheus group成功"},
		Group:    buildGroupItem(groupDetail),
	}, nil
}

func (s *PromLogic) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.ListGroup")
	defer span.End()

	groups, total, err := s.repo.Groups(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorw("获取Prometheus分组列表失败", req, "err", err)
		return nil, perrors.ErrorServerDatabaseError("获取Prometheus分组列表失败")
	}

	list := make([]*prom.GroupItem, 0, len(groups))
	for _, group := range groups {
		list = append(list, buildGroupItem(group))
	}

	return &pb.ListGroupReply{
		Response: &api.Response{Message: "获取Prometheus group列表成功"},
		Result: &api.ListQueryResult{
			Page: &api.PageReply{
				Current: req.GetQuery().GetPage().GetCurrent(),
				Size:    req.GetQuery().GetPage().GetSize(),
				Total:   total,
			},
			// TODO 暂时不返回fields
			Fields: nil,
		},
		Groups: list,
	}, nil
}

func (s *PromLogic) CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.CreateStrategy")
	defer span.End()

	firstStrategyInfo, err := s.repo.GetStrategyByName(ctx, req.GetStrategy().GetGroupId(), req.GetStrategy().Alert)
	if err != nil && !perrors.IsLogicDataNotFound(err) {
		s.logger.WithContext(ctx).Errorw("创建Prometheus策略失败", req.GetStrategy(), "err", err)
		return nil, err
	}
	if firstStrategyInfo != nil && firstStrategyInfo.ID != 0 {
		return nil, perrors.ErrorLogicDataDuplicate("策略名称已存在").WithMetadata(map[string]string{
			"alert":    req.GetStrategy().GetAlert(),
			"gourp_id": strconv.Itoa(int(req.GetStrategy().GetGroupId())),
		})
	}

	if err = s.repo.CreateStrategy(ctx, buildModelPromStrategy(req.GetStrategy())); err != nil {
		s.logger.WithContext(ctx).Errorw("创建Prometheus策略失败", req.GetStrategy(), "err", err)
		return nil, err
	}

	return &pb.CreateStrategyReply{Response: &api.Response{Message: "创建Prometheus strategy成功"}}, nil
}

func (s *PromLogic) UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.UpdateStrategy")
	defer span.End()

	firstStrategyInfo, err := s.repo.GetStrategyByName(ctx, req.GetStrategy().GetGroupId(), req.GetStrategy().Alert)
	if err != nil && !perrors.IsLogicDataNotFound(err) {
		s.logger.WithContext(ctx).Errorw("修改Prometheus策略失败", req.GetStrategy(), "err", err)
		return nil, err
	}
	if firstStrategyInfo != nil && firstStrategyInfo.ID != req.GetId() {
		return nil, perrors.ErrorLogicDataDuplicate("策略名称已存在").WithMetadata(map[string]string{
			"alert":    req.GetStrategy().GetAlert(),
			"group_id": strconv.Itoa(int(req.GetStrategy().GetGroupId())),
		})
	}

	if err = s.repo.UpdateStrategyByID(ctx, req.GetId(), buildModelPromStrategy(req.GetStrategy())); err != nil {
		s.logger.WithContext(ctx).Errorw("修改Prometheus策略失败", req.GetStrategy(), "err", err)
		return nil, err
	}

	return &pb.UpdateStrategyReply{Response: &api.Response{Message: "修改Prometheus strategy成功"}}, nil
}

func (s *PromLogic) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.DeleteStrategy")
	defer span.End()

	if err := s.repo.DeleteStrategyByID(ctx, req.GetId()); err != nil {
		s.logger.WithContext(ctx).Errorw("删除Prometheus策略失败", "id", req.GetId(), "err", err)
		return nil, err
	}
	return &pb.DeleteStrategyReply{Response: &api.Response{Message: "删除Prometheus策略成功"}}, nil
}

func (s *PromLogic) GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.GetStrategy")
	defer span.End()

	strategyDetail, err := s.repo.StrategyDetail(ctx, req.GetId())
	if err != nil {
		s.logger.WithContext(ctx).Errorw("获取Prometheus策略失败", "id", req.GetId(), "err", err)
		return nil, err
	}

	return &pb.GetStrategyReply{
		Response: &api.Response{Message: "获取Prometheus策略成功"},
		Strategy: buildStrategyItem(strategyDetail),
	}, nil
}

func (s *PromLogic) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.ListStrategy")
	defer span.End()

	strategies, total, err := s.repo.Strategies(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorw("获取Prometheus策略列表失败", req, "err", err)
		return nil, perrors.ErrorServerDatabaseError("获取Prometheus策略列表失败")
	}

	return &pb.ListStrategyReply{
		Response: &api.Response{Message: "获取Prometheus策略列表成功"},
		Result: &api.ListQueryResult{
			Page: &api.PageReply{
				Current: req.GetQuery().GetPage().GetCurrent(),
				Size:    req.GetQuery().GetPage().GetSize(),
				Total:   total,
			},
			// TODO 暂时不返回fields
			Fields: nil,
		},
		Strategies: buildPromStrategies(strategies),
	}, nil
}
