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
		Id          uint32               `json:"id"`
		Status      valueobj.Status      `json:"status"`
		CreatedAt   int64                `json:"createdAt"`
		UpdatedAt   int64                `json:"updatedAt"`
		DeletedAt   int64                `json:"deletedAt"`
		MemberId    uint32               `json:"memberId"`
		Member      *UserBO              `json:"member"`
		NotifyTypes valueobj.NotifyTypes `json:"notifyTypes"`
	}
)

// GetMember 获取用户详情
func (b *NotifyMemberBO) GetMember() *UserBO {
	if b == nil {
		return nil
	}
	return b.Member
}

// GetNotifyTypes 获取通知类型
func (b *NotifyMemberBO) GetNotifyTypes() valueobj.NotifyTypes {
	if b == nil {
		return nil
	}
	return b.NotifyTypes
}

func (b *NotifyMemberBO) ToModel() *model.PromAlarmNotifyMember {
	if b == nil {
		return nil
	}
	return &model.PromAlarmNotifyMember{
		BaseModel:   query.BaseModel{ID: b.Id},
		Status:      b.Status,
		NotifyTypes: b.GetNotifyTypes(),
		MemberId:    b.MemberId,
		Member:      b.GetMember().ToModel(),
	}
}

// ToApi ...
func (b *NotifyMemberBO) ToApi() *api.BeNotifyMemberDetail {
	if b == nil {
		return nil
	}

	return &api.BeNotifyMemberDetail{
		MemberId:    b.MemberId,
		NotifyTypes: slices.To(b.GetNotifyTypes(), func(i valueobj.NotifyType) int32 { return i.Value() }),
		User:        b.GetMember().ToApiSelectV1(),
		Status:      b.Status.Value(),
		Id:          b.Id,
	}
}

// NotifyMemberApiToBO ...
func NotifyMemberApiToBO(a *api.BeNotifyMember) *NotifyMemberBO {
	if a == nil {
		return nil
	}
	return &NotifyMemberBO{
		Id:          a.Id,
		MemberId:    a.GetMemberId(),
		Member:      &UserBO{Id: a.GetMemberId()},
		NotifyTypes: slices.To(a.GetNotifyTypes(), func(i api.NotifyType) valueobj.NotifyType { return valueobj.NotifyType(i) }),
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
		Member:      UserModelToBO(m.GetMember()),
		NotifyTypes: m.GetNotifyTypes(),
	}
}
