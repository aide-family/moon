package bo

import (
	"encoding"
	"encoding/json"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
)

var _ encoding.BinaryMarshaler = (*AlarmUpgradeBO)(nil)
var _ encoding.BinaryUnmarshaler = (*AlarmUpgradeBO)(nil)

type (
	AlarmUpgradeBO struct {
		ID              uint32  `json:"id"`
		RealtimeAlarmID uint32  `json:"realtimeAlarmID"`
		UserID          uint32  `json:"userID"`
		UpgradedAt      int64   `json:"upgradedAt"`
		Remark          string  `json:"remark"`
		User            *UserBO `json:"user"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

func (l *AlarmUpgradeBO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, l)
}

func (l *AlarmUpgradeBO) MarshalBinary() (data []byte, err error) {
	return json.Marshal(l)
}

// ToModel 转换为model
func (l *AlarmUpgradeBO) ToModel() *model.PromAlarmUpgrade {
	if l == nil {
		return nil
	}
	return &model.PromAlarmUpgrade{
		BaseModel:       query.BaseModel{ID: l.ID},
		RealtimeAlarmID: l.RealtimeAlarmID,
		UserID:          l.UserID,
		UpgradedAt:      l.UpgradedAt,
		Remark:          l.Remark,
	}
}

// GetUser 获取用户
func (l *AlarmUpgradeBO) GetUser() *UserBO {
	if l == nil {
		return nil
	}
	return l.User
}

// ToApi 转换为api
func (l *AlarmUpgradeBO) ToApi() *api.AlarmUpgradeInfo {
	if l == nil {
		return nil
	}
	return &api.AlarmUpgradeInfo{
		UpgradedUser: l.GetUser().ToApiSelectV1(),
		UpgradedAt:   l.UpgradedAt,
		Remark:       l.Remark,
		Id:           l.ID,
	}
}

// AlarmUpgradeModelToBO .
func AlarmUpgradeModelToBO(m *model.PromAlarmUpgrade) *AlarmUpgradeBO {
	if m == nil {
		return nil
	}
	return &AlarmUpgradeBO{
		ID:              m.ID,
		RealtimeAlarmID: m.RealtimeAlarmID,
		UserID:          m.UserID,
		UpgradedAt:      m.UpgradedAt,
		Remark:          m.Remark,
		CreatedAt:       m.CreatedAt.Unix(),
		UpdatedAt:       m.UpdatedAt.Unix(),
		DeletedAt:       int64(m.DeletedAt),
	}
}
