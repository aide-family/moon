package biz

import (
	"context"
	"encoding/json"

	"github.com/google/wire"

	"prometheus-manager/api/prom"
	"prometheus-manager/dal/model"
	"prometheus-manager/pkg/times"

	"prometheus-manager/apps/master/internal/service"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewPingLogic,
	wire.Bind(new(service.IPingLogic), new(*PingLogic)),
	NewPushLogic,
	wire.Bind(new(service.IPushLogic), new(*PushLogic)),
	NewPromLogic,
	wire.Bind(new(service.IPromV1Logic), new(*PromLogic)),
)

type V1Repo interface {
	V1(ctx context.Context) string
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
