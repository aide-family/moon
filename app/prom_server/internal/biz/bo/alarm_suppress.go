package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
)

type (
	AlarmSuppressBO struct {
		ID              uint    `json:"id"`
		RealtimeAlarmID uint    `json:"realtimeAlarmID"`
		UserID          uint    `json:"userID"`
		SuppressedAt    int64   `json:"suppressedAt"`
		Remark          string  `json:"remark"`
		Duration        int64   `json:"duration"`
		User            *UserBO `json:"user"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// GetUser 获取用户
func (l *AlarmSuppressBO) GetUser() *UserBO {
	if l == nil {
		return nil
	}
	return l.User
}

// ToModel 转换为model
func (l *AlarmSuppressBO) ToModel() *model.PromAlarmSuppress {
	if l == nil {
		return nil
	}
	return &model.PromAlarmSuppress{
		BaseModel:       query.BaseModel{ID: l.ID},
		RealtimeAlarmID: l.RealtimeAlarmID,
		UserID:          l.UserID,
		SuppressedAt:    l.SuppressedAt,
		Remark:          l.Remark,
		Duration:        l.Duration,
	}
}

// ToApi 转换为api
func (l *AlarmSuppressBO) ToApi() *api.AlarmSuppressInfo {
	if l == nil {
		return nil
	}
	return &api.AlarmSuppressInfo{
		SuppressedUser: l.GetUser().ToApiSelectV1(),
		SuppressedAt:   l.SuppressedAt,
		Remark:         l.Remark,
		Duration:       l.Duration,
		Id:             uint32(l.ID),
	}
}

// AlarmSuppressModelToBO model转换为bo
func AlarmSuppressModelToBO(m *model.PromAlarmSuppress) *AlarmSuppressBO {
	if m == nil {
		return nil
	}
	return &AlarmSuppressBO{
		ID:              m.ID,
		RealtimeAlarmID: m.RealtimeAlarmID,
		UserID:          m.UserID,
		SuppressedAt:    m.SuppressedAt,
		Remark:          m.Remark,
		Duration:        m.Duration,
		CreatedAt:       m.CreatedAt.Unix(),
		UpdatedAt:       m.UpdatedAt.Unix(),
		DeletedAt:       int64(m.DeletedAt),
	}
}
