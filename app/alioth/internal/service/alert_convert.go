package service

import (
	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func datasourceItem(item bo.Datasource) *v1.DatasourceItem {
	return &v1.DatasourceItem{
		Uid: item.UID, Name: item.Name, Type: item.Type, Driver: item.Driver, Endpoint: item.Endpoint,
		IndexPattern: item.IndexPattern, TimeField: item.TimeField,
		Username: item.Username, SecretSet: item.SecretSet, Status: item.Status,
		LevelUid: item.LevelUID, LevelName: item.LevelName, Labels: item.Labels, Remark: item.Remark,
	}
}

func datasourceItems(items []bo.Datasource) []*v1.DatasourceItem {
	out := make([]*v1.DatasourceItem, 0, len(items))
	for _, item := range items {
		out = append(out, datasourceItem(item))
	}
	return out
}

func groupItem(item bo.AlertRuleGroup) *v1.AlertRuleGroupItem {
	return &v1.AlertRuleGroupItem{Uid: item.UID, Name: item.Name, Status: item.Status}
}

func groupItems(items []bo.AlertRuleGroup) []*v1.AlertRuleGroupItem {
	out := make([]*v1.AlertRuleGroupItem, 0, len(items))
	for _, item := range items {
		out = append(out, groupItem(item))
	}
	return out
}

func ruleItem(item bo.AlertRule) *v1.AlertRuleItem {
	levels := make([]*v1.AlertRuleLevelItem, 0, len(item.Levels))
	for _, level := range item.Levels {
		levels = append(levels, &v1.AlertRuleLevelItem{
			Uid: level.UID, Condition: level.Condition, Sample: level.Sample, Threshold: level.Threshold,
			ForSeconds: int32(level.ForSeconds), LevelUid: level.LevelUID, LevelName: level.LevelName,
		})
	}
	return &v1.AlertRuleItem{
		Uid: item.UID, GroupUid: item.GroupUID, Name: item.Name, Type: item.Type, Driver: item.Driver,
		Status: item.Status, Expr: item.Expr, Labels: item.Labels, DatasourceUid: item.DatasourceUID,
		DatasourceLabels: item.DatasourceLabels,
		Summary:          item.Summary, Description: item.Description, Levels: levels,
	}
}

func ruleItems(items []bo.AlertRule) []*v1.AlertRuleItem {
	out := make([]*v1.AlertRuleItem, 0, len(items))
	for _, item := range items {
		out = append(out, ruleItem(item))
	}
	return out
}

func levelInputs(items []*v1.AlertRuleLevelInput) []bo.AlertRuleLevel {
	out := make([]bo.AlertRuleLevel, 0, len(items))
	for _, item := range items {
		out = append(out, bo.AlertRuleLevel{
			Condition: item.GetCondition(), Sample: item.GetSample(), Threshold: item.GetThreshold(),
			ForSeconds: int(item.GetForSeconds()), LevelUID: item.GetLevelUid(),
		})
	}
	return out
}

func dictionaryItem(item bo.AlertDictionary) *v1.AlertDictionaryItem {
	return &v1.AlertDictionaryItem{Uid: item.UID, Name: item.Name, Color: item.Color, Type: item.Type, Status: item.Status}
}

func dictionaryItems(items []bo.AlertDictionary) []*v1.AlertDictionaryItem {
	out := make([]*v1.AlertDictionaryItem, 0, len(items))
	for _, item := range items {
		out = append(out, dictionaryItem(item))
	}
	return out
}
