package biz

import (
	"context"
	"encoding/json"
	promPB "prometheus-manager/api/prom"
	"prometheus-manager/dal/model"
	"prometheus-manager/pkg/times"
)

type V1Repo interface {
	V1(ctx context.Context) string
}

func toDirs(nodeDirs []*model.PromNodeDir) []*promPB.DirItem {
	list := make([]*promPB.DirItem, 0, len(nodeDirs))
	for _, dir := range nodeDirs {
		list = append(list, &promPB.DirItem{
			NodeId:    uint32(dir.NodeID),
			Path:      dir.Path,
			CreatedAt: times.TimeToUnix(dir.CreatedAt),
			UpdatedAt: times.TimeToUnix(dir.UpdatedAt),
			Files:     toFiles(dir.Files),
			Id:        uint32(dir.ID),
		})
	}
	return list
}

func toFiles(files []*model.PromNodeDirFile) []*promPB.FileItem {
	list := make([]*promPB.FileItem, 0, len(files))
	for _, file := range files {
		list = append(list, &promPB.FileItem{
			Filename:  file.Filename,
			DirId:     uint32(file.DirID),
			CreatedAt: times.TimeToUnix(file.CreatedAt),
			UpdatedAt: times.TimeToUnix(file.UpdatedAt),
			Id:        uint32(file.ID),
			Groups:    toGroups(file.Groups),
		})
	}
	return list
}

func toGroups(groups []*model.PromNodeDirFileGroup) []*promPB.GroupItem {
	list := make([]*promPB.GroupItem, 0, len(groups))
	for _, group := range groups {
		list = append(list, &promPB.GroupItem{
			Name:      group.Name,
			Remark:    group.Remark,
			FileId:    uint32(group.FileID),
			CreatedAt: times.TimeToUnix(group.CreatedAt),
			UpdatedAt: times.TimeToUnix(group.UpdatedAt),
			Id:        uint32(group.ID),
			Rules:     toRules(group.Strategies),
		})
	}
	return list
}

func toRules(rules []*model.PromNodeDirFileGroupStrategy) []*promPB.RuleItem {
	list := make([]*promPB.RuleItem, 0, len(rules))
	for _, rule := range rules {
		labels := make(map[string]string)
		annotations := make(map[string]string)
		_ = json.Unmarshal([]byte(rule.Labels), &labels)
		_ = json.Unmarshal([]byte(rule.Annotations), &annotations)
		list = append(list, &promPB.RuleItem{
			GroupId:     uint32(rule.GroupID),
			Alert:       rule.Alert,
			Expr:        rule.Expr,
			For:         rule.For,
			Labels:      labels,
			Annotations: annotations,
			CreatedAt:   times.TimeToUnix(rule.CreatedAt),
			UpdatedAt:   times.TimeToUnix(rule.UpdatedAt),
			Id:          uint32(rule.ID),
		})
	}
	return list
}
