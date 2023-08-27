package data

import (
	"prometheus-manager/pkg/alert"
	"prometheus-manager/pkg/dal/model"
	"time"
)

func alertDataToModel(info *alert.Data) *model.PromAlarmHistory {
	if info == nil {
		return nil
	}

	return &model.PromAlarmHistory{
		ID:         0,
		Node:       "",
		Status:     "",
		Info:       "",
		CreatedAt:  time.Time{},
		StartAt:    0,
		EndAt:      0,
		Duration:   0,
		StrategyID: 0,
		LevelID:    0,
		Md5:        "",
		Pages:      nil,
	}
}
