package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type (
	NotifyMemberBO struct {
		Id          uint                 `json:"id"`
		Status      valueobj.Status      `json:"status"`
		CreatedAt   int64                `json:"createdAt"`
		UpdatedAt   int64                `json:"updatedAt"`
		DeletedAt   int64                `json:"deletedAt"`
		MemberId    uint                 `json:"memberId"`
		Member      *UserBO              `json:"member"`
		NotifyTypes valueobj.NotifyTypes `json:"notifyTypes"`
	}
)

func (b *NotifyMemberBO) ToModel() *model.PromAlarmNotifyMember {
	if b == nil {
		return nil
	}
	return &model.PromAlarmNotifyMember{
		BaseModel:   query.BaseModel{ID: b.Id},
		Status:      b.Status,
		NotifyTypes: b.NotifyTypes,
		MemberId:    b.MemberId,
		Member:      b.Member.ToModel(),
	}
}

// ToApi ...
func (b *NotifyMemberBO) ToApi() *api.BeNotifyMemberDetail {
	if b == nil {
		return nil
	}

	return &api.BeNotifyMemberDetail{
		MemberId:    uint32(b.MemberId),
		NotifyTypes: slices.To(b.NotifyTypes, func(i valueobj.NotifyType) int32 { return i.Value() }),
		User:        b.Member.ToApiSelectV1(),
		Status:      b.Status.Value(),
		Id:          uint32(b.Id),
	}
}

// NotifyMemberApiToBO ...
func NotifyMemberApiToBO(a *api.BeNotifyMember) *NotifyMemberBO {
	if a == nil {
		return nil
	}
	return &NotifyMemberBO{
		Id:          uint(a.Id),
		Status:      0,
		CreatedAt:   0,
		UpdatedAt:   0,
		DeletedAt:   0,
		MemberId:    uint(a.MemberId),
		Member:      &UserBO{Id: uint(a.MemberId)},
		NotifyTypes: slices.To(a.NotifyTypes, func(i api.NotifyType) valueobj.NotifyType { return valueobj.NotifyType(i) }),
	}
}

// NotifyMemberModelToBO ...
func NotifyMemberModelToBO(m *model.PromAlarmNotifyMember) *NotifyMemberBO {
	if m == nil {
		return nil
	}
	return &NotifyMemberBO{
		Id:          m.ID,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Unix(),
		UpdatedAt:   m.UpdatedAt.Unix(),
		DeletedAt:   int64(m.DeletedAt),
		MemberId:    m.MemberId,
		Member:      UserModelToBO(m.Member),
		NotifyTypes: m.NotifyTypes,
	}
}
